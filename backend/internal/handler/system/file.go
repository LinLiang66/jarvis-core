package system

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"jarvis-core/backend/internal/config"
	"jarvis-core/backend/internal/database"
	"jarvis-core/backend/internal/middleware"
	"jarvis-core/backend/internal/pkg/parseid"
	"jarvis-core/backend/internal/pkg/response"
	"jarvis-core/backend/internal/pkg/serialize"
	storagesvc "jarvis-core/backend/internal/service/storage"
	"jarvis-core/backend/internal/store"
)

type FileHandler struct {
	app *database.App
	cfg *config.Config
	svc *storagesvc.Service
}

func NewFileHandler(app *database.App, cfg *config.Config) *FileHandler {
	return &FileHandler{
		app: app,
		cfg: cfg,
		svc: storagesvc.NewService(cfg, app.Stores),
	}
}

func (h *FileHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/file")
	g.GET("/list", h.List)
	g.GET("/statistics", h.Statistics)
	g.POST("/upload", h.Upload)
	g.POST("/dir", h.CreateDir)
	g.POST("/delete", middleware.RequireSuperAdmin(h.app), h.Delete)
}

func (h *FileHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	storageID, _ := strconv.ParseInt(c.Query("storageId"), 10, 64)
	filter := store.SysFileFilter{
		StorageID:    storageID,
		ParentPath:   storagesvc.NormalizeParentPath(c.DefaultQuery("parentPath", "/")),
		OriginalName: c.Query("originalName"),
	}
	if t := c.Query("type"); t != "" {
		v, _ := strconv.Atoi(t)
		filter.Type = &v
	}
	list, total, err := h.app.Stores.SysFile.List(c.Request.Context(), store.PageQuery{Page: page, Size: size}, filter)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	var out []map[string]any
	for _, row := range list {
		out = append(out, serialize.FileDTO(row))
	}
	response.Page(c, out, int(total), page, size)
}

func (h *FileHandler) Statistics(c *gin.Context) {
	fileCount, dirCount, totalSize, err := h.app.Stores.SysFile.Statistics(c.Request.Context())
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, map[string]any{
		"fileCount": fileCount,
		"dirCount":  dirCount,
		"totalSize": totalSize,
	})
}

func (h *FileHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, 400, "请选择文件")
		return
	}
	f, err := file.Open()
	if err != nil {
		response.Fail(c, 400, "读取文件失败")
		return
	}
	defer f.Close()

	storageID, _ := strconv.ParseInt(c.PostForm("storageId"), 10, 64)
	parentPath := c.PostForm("parentPath")
	contentType := file.Header.Get("Content-Type")

	row, err := h.svc.Upload(c.Request.Context(), storageID, parentPath, file.Filename, file.Size, f, contentType)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, map[string]any{
		"id":  serialize.IDStr(row.ID),
		"url": row.URL,
	})
}

func (h *FileHandler) CreateDir(c *gin.Context) {
	var req struct {
		ParentPath   string `json:"parentPath"`
		OriginalName string `json:"originalName"`
		StorageID    string `json:"storageId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	storageID, _ := strconv.ParseInt(req.StorageID, 10, 64)
	row, err := h.svc.CreateDir(c.Request.Context(), storageID, req.ParentPath, req.OriginalName)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, serialize.FileDTO(*row))
}

func (h *FileHandler) Delete(c *gin.Context) {
	ids, err := parseid.BindDeleteIDs(c)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	if err := h.svc.DeleteFiles(c.Request.Context(), ids); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, nil)
}
