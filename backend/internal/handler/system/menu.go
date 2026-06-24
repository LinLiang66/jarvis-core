package system



import (

	"sort"

	"strconv"

	"strings"



	"github.com/gin-gonic/gin"



	"jarvis-core/backend/internal/database"

	"jarvis-core/backend/internal/middleware"

	"jarvis-core/backend/internal/model"

	"jarvis-core/backend/internal/pkg/parseid"

	"jarvis-core/backend/internal/pkg/response"

	"jarvis-core/backend/internal/pkg/serialize"

	"jarvis-core/backend/internal/service/rbac"

)



type MenuHandler struct{ app *database.App }



func NewMenuHandler(app *database.App) *MenuHandler { return &MenuHandler{app: app} }



func (h *MenuHandler) Register(rg *gin.RouterGroup) {

	g := rg.Group("/menu")

	g.GET("/routes", h.Routes)

	g.Use(middleware.RequireSuperAdmin(h.app))

	g.GET("/tree", h.Tree)

	g.POST("", h.Create)

	g.PUT("/:id", h.Update)

	g.POST("/delete", h.Delete)

}



func (h *MenuHandler) Tree(c *gin.Context) {

	all, err := h.app.Stores.SysMenu.All(c.Request.Context())

	if err != nil {

		response.Fail(c, 500, err.Error())

		return

	}

	response.OK(c, buildMenuTree(all, 0))

}



// Routes 对齐上游 FastAPI GET /menu/routes

func (h *MenuHandler) Routes(c *gin.Context) {

	uid, _ := c.Get("user_id")

	user, err := h.app.Stores.SysUser.GetByID(c.Request.Context(), uid)

	if err != nil {

		response.Fail(c, 401, "未登录")

		return

	}

	menus, err := h.app.Stores.SysMenu.ListForRoutes(c.Request.Context(), *user, rbac.IsSuperAdmin(*user))

	if err != nil {

		response.Fail(c, 500, err.Error())

		return

	}

	byID := make(map[int64]model.SysMenu, len(menus))

	for _, m := range menus {

		byID[m.ID] = m

	}

	var items []map[string]any

	for _, m := range menus {

		if m.Type != 1 && m.Type != 2 {

			continue

		}

		if m.Status != "0" {

			continue

		}

		if m.Hidden {

			continue

		}

		items = append(items, menuToRouteItem(m, byID))

	}

	response.OK(c, buildRouteTreeByMap(items))

}



func (h *MenuHandler) Create(c *gin.Context) {

	var m model.SysMenu

	if err := c.ShouldBindJSON(&m); err != nil {

		response.Fail(c, 400, "参数错误")

		return

	}

	if err := bindMenuJSON(c, &m); err != nil {

		response.Fail(c, 400, "参数错误")

		return

	}

	if m.Status == "" {

		m.Status = "0"

	}

	if err := h.app.Stores.SysMenu.Create(c.Request.Context(), &m); err != nil {

		response.Fail(c, 500, err.Error())

		return

	}

	response.OK(c, serialize.MenuDTO(m))

}



func (h *MenuHandler) Update(c *gin.Context) {

	id, err := parseid.GinKey(c, "id")

	if err != nil {

		response.Fail(c, 400, err.Error())

		return

	}

	m, err := h.app.Stores.SysMenu.GetByID(c.Request.Context(), id)

	if err != nil {

		response.Fail(c, 404, "菜单不存在")

		return

	}

	if err := bindMenuJSON(c, m); err != nil {

		response.Fail(c, 400, "参数错误")

		return

	}

	if err := h.app.Stores.SysMenu.Save(c.Request.Context(), m); err != nil {

		response.Fail(c, 500, err.Error())

		return

	}

	response.OK(c, serialize.MenuDTO(*m))

}



func (h *MenuHandler) Delete(c *gin.Context) {

	ids, err := parseid.BindDeleteIDs(c)

	if err != nil {

		response.Fail(c, 400, err.Error())

		return

	}

	_ = h.app.Stores.SysMenu.DeleteByIDs(c.Request.Context(), ids)

	response.OK(c, nil)

}



func bindMenuJSON(c *gin.Context, m *model.SysMenu) error {

	var raw map[string]any

	if err := c.ShouldBindJSON(&raw); err != nil {

		return err

	}

	if v, ok := raw["title"].(string); ok {

		m.Title = v

	}

	if v, ok := raw["path"].(string); ok {

		m.Path = v

	}

	if v, ok := raw["component"].(string); ok {

		m.Component = v

	}

	if v, ok := raw["redirect"].(string); ok {

		m.Redirect = v

	}

	if v, ok := raw["icon"].(string); ok {

		m.Icon = v

	}

	if v, ok := raw["permission"].(string); ok {

		m.Permission = v

	}

	if v, ok := raw["status"].(string); ok {

		m.Status = v

	}

	if v, ok := raw["activeMenu"].(string); ok {

		m.ActiveMenu = v

	}

	if v, ok := raw["parentId"].(string); ok && v != "" && v != "0" {

		pid, _ := strconv.ParseInt(v, 10, 64)

		m.ParentID = pid

	}

	if v, ok := raw["type"].(float64); ok {

		m.Type = int(v)

	}

	if v, ok := raw["sort"].(float64); ok {

		m.Sort = int(v)

	}

	if v, ok := raw["hidden"].(bool); ok {

		m.Hidden = v

	}

	if v, ok := raw["keepAlive"].(bool); ok {

		m.KeepAlive = v

	}

	if v, ok := raw["affix"].(bool); ok {

		m.Affix = v

	}

	if v, ok := raw["alwaysShow"].(bool); ok {

		m.AlwaysShow = v

	}

	return nil

}



func buildMenuTree(all []model.SysMenu, parentID int64) []map[string]any {

	var nodes []map[string]any

	for _, m := range all {

		if m.ParentID == parentID {

			dto := serialize.MenuDTO(m)

			dto["children"] = buildMenuTree(all, m.ID)

			nodes = append(nodes, dto)

		}

	}

	sort.Slice(nodes, func(i, j int) bool {

		return nodes[i]["sort"].(int) < nodes[j]["sort"].(int)

	})

	return nodes

}



// buildRouteTreeByMap 对齐上游 build_tree：按 parentId 挂树，叶子 children 为 []

func buildRouteTreeByMap(items []map[string]any) []map[string]any {

	itemMap := make(map[string]map[string]any, len(items))

	for _, item := range items {

		id, _ := item["id"].(string)

		node := make(map[string]any, len(item))

		for k, v := range item {

			node[k] = v

		}

		node["children"] = []map[string]any{}

		itemMap[id] = node

	}

	var roots []map[string]any

	for _, item := range itemMap {

		parentID, _ := item["parentId"].(string)

		if parentID != "" && parentID != "0" {

			if parent, ok := itemMap[parentID]; ok {

				children := parent["children"].([]map[string]any)

				parent["children"] = append(children, item)

				continue

			}

		}

		roots = append(roots, item)

	}

	sortRouteNodes(roots)

	return roots

}



func sortRouteNodes(nodes []map[string]any) {

	sort.Slice(nodes, func(i, j int) bool {

		return nodes[i]["sort"].(int) < nodes[j]["sort"].(int)

	})

	for _, node := range nodes {

		if children, ok := node["children"].([]map[string]any); ok && len(children) > 0 {

			sortRouteNodes(children)

		}

	}

}



// menuToRouteItem 对齐上游 menu_to_route_item；path 输出完整 route_path

func menuToRouteItem(m model.SysMenu, byID map[int64]model.SysMenu) map[string]any {

	path := resolveRoutePath(m, byID)

	component := m.Component

	if component == "" {

		if m.Type == 1 {

			component = "Layout"

		}

	}

	title := m.Title

	if title == "" {

		title = m.Name

	}

	parentID := "0"

	if m.ParentID > 0 {

		parentID = serialize.IDStr(m.ParentID)

	}

	return map[string]any{

		"id":          serialize.IDStr(m.ID),

		"parentId":    parentID,

		"type":        m.Type,

		"title":       title,

		"path":        path,

		"component":   component,

		"redirect":    m.Redirect,

		"icon":        m.Icon,

		"permission":  m.Permission,

		"sort":        m.Sort,

		"status":      m.Status,

		"hidden":      m.Hidden,

		"keepAlive":   m.KeepAlive,

		"affix":       m.Affix,

		"alwaysShow":  m.AlwaysShow,

		"breadcrumb":  m.Breadcrumb,

		"showInTabs":  m.ShowInTabs,

		"activeMenu":  m.ActiveMenu,

		"roles":       []string{},

		"children":    []map[string]any{},

	}

}



func resolveRoutePath(m model.SysMenu, byID map[int64]model.SysMenu) string {

	path := strings.TrimSpace(m.Path)

	if path == "" {

		return ""

	}

	if strings.HasPrefix(path, "/") {

		return path

	}

	if m.ParentID > 0 {

		if parent, ok := byID[m.ParentID]; ok {

			parentPath := resolveRoutePath(parent, byID)

			if parentPath != "" {

				return strings.TrimSuffix(parentPath, "/") + "/" + path

			}

		}

	}

	if m.Type == 1 {

		return "/" + path

	}

	return path

}


