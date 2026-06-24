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

type DictHandler struct{ app *database.App }

func NewDictHandler(app *database.App) *DictHandler { return &DictHandler{app: app} }

func (h *DictHandler) Register(rg *gin.RouterGroup) {
	t := rg.Group("/dict/type")
	t.Use(middleware.RequireSuperAdmin(h.app))
	t.GET("/list", h.TypeList)
	t.POST("", h.TypeCreate)
	t.PUT("/:id", h.TypeUpdate)
	t.POST("/delete", h.TypeDelete)

	d := rg.Group("/dict/data")
	d.GET("/by-code/:code", h.ByCode)
	d.Use(middleware.RequireSuperAdmin(h.app))
	d.GET("/list", h.DataList)
	d.POST("", h.DataCreate)
	d.PUT("/:id", h.DataUpdate)
	d.PUT("/:id/status", h.DataStatus)
	d.POST("/delete", h.DataDelete)
}

func (h *DictHandler) TypeList(c *gin.Context) {
	list, err := h.app.Stores.SysDict.ListTypes(c.Request.Context(), c.Query("name"), c.Query("status"))
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	var out []map[string]any
	for _, row := range list {
		out = append(out, serialize.DictTypeDTO(row))
	}
	response.OK(c, out)
}

func (h *DictHandler) TypeCreate(c *gin.Context) {
	var raw map[string]any
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	row := model.SysDictType{Status: "0"}
	if v, ok := raw["name"].(string); ok {
		row.Name = v
	}
	if v, ok := raw["code"].(string); ok {
		row.Code = v
	}
	if v, ok := raw["status"].(string); ok {
		row.Status = v
	}
	if v, ok := raw["sort"].(float64); ok {
		row.Sort = int(v)
	}
	if v, ok := raw["remark"].(string); ok {
		row.Remark = v
	}
	if err := h.app.Stores.SysDict.CreateType(c.Request.Context(), &row); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, serialize.DictTypeDTO(row))
}

func (h *DictHandler) TypeUpdate(c *gin.Context) {
	id, err := parseid.GinKey(c, "id")
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	row, err := h.app.Stores.SysDict.GetTypeByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, 404, "记录不存在")
		return
	}
	var raw map[string]any
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if v, ok := raw["name"].(string); ok {
		row.Name = v
	}
	if v, ok := raw["code"].(string); ok {
		row.Code = v
	}
	if v, ok := raw["status"].(string); ok {
		row.Status = v
	}
	if v, ok := raw["sort"].(float64); ok {
		row.Sort = int(v)
	}
	if v, ok := raw["remark"].(string); ok {
		row.Remark = v
	}
	_ = h.app.Stores.SysDict.SaveType(c.Request.Context(), row)
	response.OK(c, serialize.DictTypeDTO(*row))
}

func (h *DictHandler) TypeDelete(c *gin.Context) {
	ids, err := parseid.BindDeleteIDs(c)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	_ = h.app.Stores.SysDict.DeleteTypes(c.Request.Context(), ids)
	response.OK(c, nil)
}

func (h *DictHandler) DataList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	list, total, err := h.app.Stores.SysDict.ListData(c.Request.Context(), store.PageQuery{Page: page, Size: size},
		c.Query("typeId"), c.Query("label"), c.Query("status"))
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	var out []map[string]any
	for _, row := range list {
		out = append(out, serialize.DictDataDTO(row))
	}
	response.Page(c, out, int(total), page, size)
}

func (h *DictHandler) DataCreate(c *gin.Context) {
	var raw map[string]any
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	row := model.SysDictData{Status: "0"}
	if v, ok := raw["typeId"].(string); ok {
		tid, _ := strconv.ParseInt(v, 10, 64)
		row.TypeID = tid
	}
	if v, ok := raw["label"].(string); ok {
		row.Label = v
	}
	if v, ok := raw["value"].(string); ok {
		row.Value = v
	}
	if v, ok := raw["status"].(string); ok {
		row.Status = v
	}
	if v, ok := raw["sort"].(float64); ok {
		row.Sort = int(v)
	}
	if err := h.app.Stores.SysDict.CreateData(c.Request.Context(), &row); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, serialize.DictDataDTO(row))
}

func (h *DictHandler) DataUpdate(c *gin.Context) {
	id, err := parseid.GinKey(c, "id")
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	row, err := h.app.Stores.SysDict.GetDataByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, 404, "记录不存在")
		return
	}
	var raw map[string]any
	_ = c.ShouldBindJSON(&raw)
	if v, ok := raw["label"].(string); ok {
		row.Label = v
	}
	if v, ok := raw["value"].(string); ok {
		row.Value = v
	}
	if v, ok := raw["status"].(string); ok {
		row.Status = v
	}
	_ = h.app.Stores.SysDict.SaveData(c.Request.Context(), row)
	response.OK(c, serialize.DictDataDTO(*row))
}

func (h *DictHandler) DataStatus(c *gin.Context) {
	id, err := parseid.GinKey(c, "id")
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	row, _ := h.app.Stores.SysDict.GetDataByID(c.Request.Context(), id)
	var req struct {
		Status string `json:"status"`
	}
	_ = c.ShouldBindJSON(&req)
	row.Status = req.Status
	_ = h.app.Stores.SysDict.SaveData(c.Request.Context(), row)
	response.OK(c, serialize.DictDataDTO(*row))
}

func (h *DictHandler) DataDelete(c *gin.Context) {
	ids, err := parseid.BindDeleteIDs(c)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	_ = h.app.Stores.SysDict.DeleteData(c.Request.Context(), ids)
	response.OK(c, nil)
}

func (h *DictHandler) ByCode(c *gin.Context) {
	rows, err := h.app.Stores.SysDict.OptionsByCode(c.Request.Context(), c.Param("code"))
	if err != nil {
		response.Fail(c, 404, "字典不存在")
		return
	}
	var out []map[string]string
	for _, r := range rows {
		out = append(out, serialize.DictOption(r.Label, r.Value))
	}
	response.OK(c, out)
}
