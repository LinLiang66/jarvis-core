package system

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"jarvis-core/backend/internal/config"
	"jarvis-core/backend/internal/database"
	"jarvis-core/backend/internal/middleware"
	"jarvis-core/backend/internal/model"
	"jarvis-core/backend/internal/pkg/parseid"
	"jarvis-core/backend/internal/pkg/response"
	"jarvis-core/backend/internal/pkg/serialize"
)

type StorageHandler struct {
	app *database.App
	cfg *config.Config
}

func NewStorageHandler(app *database.App, cfg *config.Config) *StorageHandler {
	return &StorageHandler{app: app, cfg: cfg}
}

func (h *StorageHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/storage")
	g.Use(middleware.RequireSuperAdmin(h.app))
	g.GET("/list", h.List)
	g.GET("/:id", h.Detail)
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.PUT("/:id/status", h.UpdateStatus)
	g.PUT("/:id/default", h.SetDefault)
	g.POST("/delete", h.Delete)
}

func (h *StorageHandler) List(c *gin.Context) {
	storageType, _ := strconv.Atoi(c.Query("type"))
	list, err := h.app.Stores.SysStorage.List(c.Request.Context(), storageType)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	var out []map[string]any
	for _, row := range list {
		out = append(out, serialize.StorageDTO(row, true))
	}
	response.OK(c, out)
}

func (h *StorageHandler) Detail(c *gin.Context) {
	id, err := parseid.GinKey(c, "id")
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	row, err := h.app.Stores.SysStorage.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, 404, "存储不存在")
		return
	}
	response.OK(c, serialize.StorageDTO(*row, true))
}

func (h *StorageHandler) Create(c *gin.Context) {
	var raw map[string]any
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	row, err := bindStorage(raw, model.SysStorage{Status: "0"})
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	exists, err := h.app.Stores.SysStorage.ExistsCode(c.Request.Context(), row.Code, 0)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	if exists {
		response.Fail(c, 400, "存储编码已存在")
		return
	}
	if err := validateStorage(row, true); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	row.IsDefault = false
	if err := h.app.Stores.SysStorage.Create(c.Request.Context(), &row); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, serialize.StorageDTO(row, true))
}

func (h *StorageHandler) Update(c *gin.Context) {
	id, err := parseid.GinKey(c, "id")
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	old, err := h.app.Stores.SysStorage.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, 404, "存储不存在")
		return
	}
	var raw map[string]any
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	row, err := bindStorage(raw, *old)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	row.ID = old.ID
	row.Code = old.Code
	row.Type = old.Type
	row.IsDefault = old.IsDefault
	if v, ok := raw["secretKey"].(string); !ok || v == "" {
		row.SecretKey = old.SecretKey
	}
	if err := validateStorage(row, false); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	if row.IsDefault && row.Status == "1" {
		response.Fail(c, 400, "默认存储不允许禁用")
		return
	}
	if err := h.app.Stores.SysStorage.Save(c.Request.Context(), &row); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, serialize.StorageDTO(row, true))
}

func (h *StorageHandler) UpdateStatus(c *gin.Context) {
	id, err := parseid.GinKey(c, "id")
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	row, err := h.app.Stores.SysStorage.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, 404, "存储不存在")
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if row.IsDefault && req.Status == "1" {
		response.Fail(c, 400, "默认存储不允许禁用")
		return
	}
	row.Status = req.Status
	if err := h.app.Stores.SysStorage.Save(c.Request.Context(), row); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, serialize.StorageDTO(*row, true))
}

func (h *StorageHandler) SetDefault(c *gin.Context) {
	id, err := parseid.GinKey(c, "id")
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	row, err := h.app.Stores.SysStorage.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, 404, "存储不存在")
		return
	}
	if row.Status == "1" {
		response.Fail(c, 400, "请先启用所选存储")
		return
	}
	if err := h.app.Stores.SysStorage.SetDefault(c.Request.Context(), row.ID); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	row.IsDefault = true
	response.OK(c, serialize.StorageDTO(*row, true))
}

func (h *StorageHandler) Delete(c *gin.Context) {
	ids, err := parseid.BindDeleteIDs(c)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	for _, id := range ids {
		row, err := h.app.Stores.SysStorage.GetByID(c.Request.Context(), id)
		if err != nil {
			continue
		}
		if row.IsDefault {
			response.Fail(c, 400, "默认存储不允许删除")
			return
		}
	}
	n, err := h.app.Stores.SysFile.CountByStorageIDs(c.Request.Context(), ids)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	if n > 0 {
		response.Fail(c, 400, "所选存储存在文件关联，请删除后重试")
		return
	}
	if err := h.app.Stores.SysStorage.DeleteByIDs(c.Request.Context(), ids); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, nil)
}

func bindStorage(raw map[string]any, base model.SysStorage) (model.SysStorage, error) {
	row := base
	if v, ok := raw["name"].(string); ok {
		row.Name = v
	}
	if v, ok := raw["code"].(string); ok && v != "" {
		row.Code = v
	}
	if v, ok := raw["type"].(float64); ok {
		row.Type = int(v)
	}
	if v, ok := raw["accessKey"].(string); ok {
		row.AccessKey = v
	}
	if v, ok := raw["secretKey"].(string); ok {
		row.SecretKey = v
	}
	if v, ok := raw["endpoint"].(string); ok {
		row.Endpoint = v
	}
	if v, ok := raw["bucketName"].(string); ok {
		row.BucketName = v
	}
	if v, ok := raw["baseUrl"].(string); ok {
		row.BaseURL = v
	}
	if v, ok := raw["domain"].(string); ok {
		row.Domain = v
	}
	if v, ok := raw["description"].(string); ok {
		row.Description = v
	}
	if v, ok := raw["status"].(string); ok {
		row.Status = v
	}
	if v, ok := raw["sort"].(float64); ok {
		row.Sort = int(v)
	}
	return row, nil
}

func validateStorage(row model.SysStorage, creating bool) error {
	if row.Name == "" {
		return errMsg("名称不能为空")
	}
	if creating && row.Code == "" {
		return errMsg("编码不能为空")
	}
	if row.Type == model.StorageTypeLocal {
		if row.BucketName == "" {
			return errMsg("存储路径不能为空")
		}
		if row.Domain == "" {
			return errMsg("访问路径不能为空")
		}
	}
	if row.Type == model.StorageTypeOSS {
		if row.AccessKey == "" {
			return errMsg("Access Key 不能为空")
		}
		if creating && row.SecretKey == "" {
			return errMsg("Secret Key 不能为空")
		}
		if row.Endpoint == "" {
			return errMsg("Endpoint 不能为空")
		}
		if row.BucketName == "" {
			return errMsg("Bucket 不能为空")
		}
	}
	return nil
}

type simpleError string

func (e simpleError) Error() string { return string(e) }

func errMsg(msg string) error { return simpleError(msg) }
