package system

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"jarvis-core/backend/internal/database"
	"jarvis-core/backend/internal/middleware"
	"jarvis-core/backend/internal/model"
	"jarvis-core/backend/internal/pkg/parseid"
	"jarvis-core/backend/internal/pkg/response"
	"jarvis-core/backend/internal/pkg/serialize"
	"jarvis-core/backend/internal/store"
)

type RoleHandler struct{ app *database.App }

func NewRoleHandler(app *database.App) *RoleHandler { return &RoleHandler{app: app} }

func (h *RoleHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/role")
	g.GET("/options", h.Options)
	g.Use(middleware.RequireSuperAdmin(h.app))
	g.GET("/list", h.List)
	g.GET("/:id", h.Detail)
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.GET("/:id/menus", h.GetMenus)
	g.PUT("/:id/menus", h.SetMenus)
	g.POST("/delete", h.Delete)
}

func (h *RoleHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	list, total, err := h.app.Stores.SysRole.List(c.Request.Context(), store.PageQuery{Page: page, Size: size},
		c.Query("code"), c.Query("name"), c.Query("status"))
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	var out []map[string]any
	for _, r := range list {
		out = append(out, serialize.RoleDTO(r))
	}
	response.Page(c, out, int(total), page, size)
}

func (h *RoleHandler) Detail(c *gin.Context) {
	id, err := parseid.GinKey(c, "id")
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	r, err := h.app.Stores.SysRole.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, 404, "角色不存在")
		return
	}
	response.OK(c, serialize.RoleDTO(*r))
}

func (h *RoleHandler) Options(c *gin.Context) {
	list, err := h.app.Stores.SysRole.AllEnabled(c.Request.Context())
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	var out []map[string]any
	for _, r := range list {
		out = append(out, map[string]any{"id": serialize.IDStr(r.ID), "code": r.Code, "name": r.Name})
	}
	response.OK(c, out)
}

func (h *RoleHandler) Create(c *gin.Context) {
	var r model.SysRole
	if err := c.ShouldBindJSON(&r); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if r.Status == "" {
		r.Status = "0"
	}
	if err := h.app.Stores.SysRole.Create(c.Request.Context(), &r); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, serialize.RoleDTO(r))
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, err := parseid.GinKey(c, "id")
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	r, err := h.app.Stores.SysRole.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, 404, "角色不存在")
		return
	}
	if err := c.ShouldBindJSON(r); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := h.app.Stores.SysRole.Save(c.Request.Context(), r); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, serialize.RoleDTO(*r))
}

func (h *RoleHandler) GetMenus(c *gin.Context) {
	idKey, err := parseid.GinKey(c, "id")
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	roleID, _ := parseid.Path(idKey)
	role, err := h.app.Stores.SysRole.GetByID(c.Request.Context(), roleID)
	if err != nil {
		response.Fail(c, 404, "角色不存在")
		return
	}
	if role.Code == model.RoleAdminCode || role.IsSystem {
		response.OK(c, gin.H{"menuIds": []string{}})
		return
	}
	ids, err := h.app.Stores.SysRole.MenuIDs(c.Request.Context(), roleID)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	if ids == nil {
		ids = []string{}
	}
	response.OK(c, gin.H{"menuIds": ids})
}

func (h *RoleHandler) SetMenus(c *gin.Context) {
	idKey, err := parseid.GinKey(c, "id")
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	roleID, _ := parseid.Path(idKey)
	role, err := h.app.Stores.SysRole.GetByID(c.Request.Context(), roleID)
	if err != nil {
		response.Fail(c, 404, "角色不存在")
		return
	}
	if role.Code == model.RoleAdminCode || role.IsSystem {
		response.Fail(c, 400, "超级管理员角色无需分配菜单")
		return
	}
	var req struct {
		MenuIDs []string `json:"menuIds"`
		MenuIDsAlt []string `json:"menu_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	menuIDInputs := req.MenuIDs
	if len(menuIDInputs) == 0 {
		menuIDInputs = req.MenuIDsAlt
	}
	menuIDs, err := parseid.NormalizeStrings(menuIDInputs)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	menuIDsInt := make([]int64, len(menuIDs))
	for i, s := range menuIDs {
		menuIDsInt[i], _ = parseid.Path(s)
	}
	if err := h.app.Stores.SysRole.ReplaceMenus(c.Request.Context(), roleID, menuIDsInt); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	ids, err := parseid.BindDeleteIDs(c)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	_ = h.app.Stores.SysRole.DeleteByIDs(c.Request.Context(), ids)
	response.OK(c, nil)
}
