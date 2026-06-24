package system

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"jarvis-core/backend/internal/database"
	"jarvis-core/backend/internal/middleware"
	"jarvis-core/backend/internal/model"
	"jarvis-core/backend/internal/pkg/parseid"
	"jarvis-core/backend/internal/pkg/response"
	"jarvis-core/backend/internal/pkg/serialize"
	"jarvis-core/backend/internal/store"
)

type UserHandler struct{ app *database.App }

func NewUserHandler(app *database.App) *UserHandler { return &UserHandler{app: app} }

func (h *UserHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/user")
	g.Use(middleware.RequireSuperAdmin(h.app))
	g.GET("/list", h.List)
	g.GET("/:id", h.Detail)
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.PUT("/:id/password", h.ResetPassword)
	g.PUT("/:id/status", h.UpdateStatus)
	g.POST("/delete", h.Delete)
}

func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	list, total, err := h.app.Stores.SysUser.List(c.Request.Context(), store.PageQuery{Page: page, Size: size}, store.SysUserFilter{
		Username: c.Query("username"),
		Phone:    c.Query("phone"),
		Status:   c.Query("status"),
	})
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	var out []map[string]any
	for _, u := range list {
		full, _ := h.app.Stores.SysUser.GetByID(c.Request.Context(), u.ID)
		if full != nil {
			dto, _ := serializeUser(*full)
			out = append(out, dto)
		}
	}
	response.Page(c, out, int(total), page, size)
}

func (h *UserHandler) Detail(c *gin.Context) {
	id, err := parseid.GinKey(c, "id")
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	u, err := h.app.Stores.SysUser.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, 404, "用户不存在")
		return
	}
	dto, _ := serializeUser(*u)
	response.OK(c, dto)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Nickname string `json:"nickname"`
		Phone    string `json:"phone"`
		Email    string `json:"email"`
		Avatar   string `json:"avatar"`
		Remark   string `json:"remark"`
		Status   string `json:"status"`
		Sort     int      `json:"sort"`
		RoleIDs  []string `json:"role_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	roleIDs, err := parseid.ToInt64Slice(req.RoleIDs)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	if req.Status == "" {
		req.Status = "0"
	}
	u := model.SysUser{
		Username: req.Username, Password: string(hash), Nickname: req.Nickname,
		Phone: req.Phone, Email: req.Email, Avatar: req.Avatar, Remark: req.Remark,
		Status: req.Status, Sort: req.Sort,
	}
	if err := h.app.Stores.SysUser.Create(c.Request.Context(), &u); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	_ = h.app.Stores.SysUser.ReplaceRoles(c.Request.Context(), u.ID, roleIDs)
	full, _ := h.app.Stores.SysUser.GetByID(c.Request.Context(), u.ID)
	dto, _ := serializeUser(*full)
	response.OK(c, dto)
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := parseid.GinKey(c, "id")
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	u, err := h.app.Stores.SysUser.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, 404, "用户不存在")
		return
	}
	var req struct {
		Nickname string `json:"nickname"`
		Phone    string `json:"phone"`
		Email    string `json:"email"`
		Avatar   string `json:"avatar"`
		Remark   string `json:"remark"`
		Status   string `json:"status"`
		Sort     int      `json:"sort"`
		RoleIDs  []string `json:"role_ids"`
		Password string   `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	u.Nickname = req.Nickname
	u.Phone = req.Phone
	u.Email = req.Email
	u.Avatar = req.Avatar
	u.Remark = req.Remark
	if req.Status != "" {
		u.Status = req.Status
	}
	u.Sort = req.Sort
	if req.Password != "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		u.Password = string(hash)
	}
	if err := h.app.Stores.SysUser.Save(c.Request.Context(), u); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	if req.RoleIDs != nil {
		roleIDs, err := parseid.ToInt64Slice(req.RoleIDs)
		if err != nil {
			response.Fail(c, 400, err.Error())
			return
		}
		_ = h.app.Stores.SysUser.ReplaceRoles(c.Request.Context(), u.ID, roleIDs)
	}
	full, _ := h.app.Stores.SysUser.GetByID(c.Request.Context(), u.ID)
	dto, _ := serializeUser(*full)
	response.OK(c, dto)
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	id, err := parseid.GinKey(c, "id")
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	u, err := h.app.Stores.SysUser.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, 404, "用户不存在")
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	u.Password = string(hash)
	_ = h.app.Stores.SysUser.Save(c.Request.Context(), u)
	response.OK(c, nil)
}

func (h *UserHandler) UpdateStatus(c *gin.Context) {
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	id, err := parseid.GinKey(c, "id")
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	u, err := h.app.Stores.SysUser.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, 404, "用户不存在")
		return
	}
	u.Status = req.Status
	_ = h.app.Stores.SysUser.Save(c.Request.Context(), u)
	response.OK(c, nil)
}

func (h *UserHandler) Delete(c *gin.Context) {
	ids, err := parseid.BindDeleteIDs(c)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	_ = h.app.Stores.SysUser.DeleteByIDs(c.Request.Context(), ids)
	response.OK(c, nil)
}

func serializeUser(u model.SysUser) (map[string]any, error) {
	var roleIDs, roleNames, roles []string
	for _, r := range u.Roles {
		roleIDs = append(roleIDs, serialize.IDStr(r.ID))
		roleNames = append(roleNames, r.Name)
		roles = append(roles, r.Code)
	}
	return serialize.UserDTO(u, roleIDs, roleNames, roles), nil
}
