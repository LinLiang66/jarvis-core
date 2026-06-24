package auth

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"jarvis/backend/internal/config"
	"jarvis/backend/internal/database"
	"jarvis/backend/internal/middleware"
	"jarvis/backend/internal/model"
	"jarvis/backend/internal/pkg/response"
	"jarvis/backend/internal/pkg/serialize"
	"jarvis/backend/internal/service/rbac"
)

type Handler struct {
	app *database.App
	cfg *config.Config
}

func New(app *database.App, cfg *config.Config) *Handler {
	return &Handler{app: app, cfg: cfg}
}

func (h *Handler) Register(rg *gin.RouterGroup) {
	rg.POST("/login", h.Login)
	rg.GET("/userinfo", middleware.Auth(h.cfg, h.app.Session), h.UserInfo)
	rg.POST("/logout", middleware.Auth(h.cfg, h.app.Session), h.Logout)
}

func (h *Handler) Login(c *gin.Context) {
	username, password, err := parseLoginCredentials(c)
	if err != nil || username == "" || password == "" {
		response.Fail(c, 400, "参数错误")
		return
	}
	user, err := h.app.Stores.SysUser.GetByUsername(c.Request.Context(), username)
	if err != nil {
		response.Fail(c, 401, "用户名或密码错误")
		return
	}
	if user.Status == "1" {
		response.Fail(c, 401, "用户已停用")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		response.Fail(c, 401, "用户名或密码错误")
		return
	}
	accessToken, err := h.signToken(*user, false)
	if err != nil {
		response.Fail(c, 500, "生成令牌失败")
		return
	}
	refreshToken, err := h.signToken(*user, true)
	if err != nil {
		response.Fail(c, 500, "生成令牌失败")
		return
	}
	_ = h.app.Session.SaveToken(c.Request.Context(), accessToken, user.ID)
	expiresIn := int(h.tokenExpire(false).Seconds())
	super := rbac.IsSuperAdmin(*user)
	response.OK(c, gin.H{
		"token":        accessToken,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"expiresIn":    expiresIn,
		"tokenType":    "Bearer",
		"user":         serialize.LoginUserDTO(*user, super),
	})
}

func (h *Handler) UserInfo(c *gin.Context) {
	uid, _ := c.Get("user_id")
	user, err := h.app.Stores.SysUser.GetByID(c.Request.Context(), uid)
	if err != nil {
		response.Fail(c, 404, "用户不存在")
		return
	}
	perms, _ := rbac.PermissionsForUser(c.Request.Context(), h.app.DB, user.ID)
	response.OK(c, serialize.UserInfoDTO(*user, perms))
}

func (h *Handler) Logout(c *gin.Context) {
	if token, ok := c.Get("token"); ok {
		if s, ok := token.(string); ok {
			_ = h.app.Session.DeleteToken(c.Request.Context(), s)
		}
	}
	response.OK(c, nil)
}

func parseLoginCredentials(c *gin.Context) (username, password string, err error) {
	ct := strings.ToLower(c.GetHeader("Content-Type"))
	if strings.Contains(ct, "application/x-www-form-urlencoded") {
		return c.PostForm("username"), c.PostForm("password"), nil
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return "", "", err
	}
	return req.Username, req.Password, nil
}

func (h *Handler) tokenExpire(refresh bool) time.Duration {
	if refresh {
		days := 7
		if h.cfg.JWTRefreshDays > 0 {
			days = h.cfg.JWTRefreshDays
		}
		return time.Duration(days) * 24 * time.Hour
	}
	expire := h.cfg.JWTExpire
	if expire <= 0 {
		expire = 24 * time.Hour
	}
	return expire
}

func (h *Handler) signToken(user model.SysUser, refresh bool) (string, error) {
	claims := middleware.Claims{
		UserID:   user.ID,
		Username: user.Username,
		IsRefresh: refresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.tokenExpire(refresh))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(h.cfg.JWTSecret))
}
