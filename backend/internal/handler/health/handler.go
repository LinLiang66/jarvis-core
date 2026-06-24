package health

import (
	"github.com/gin-gonic/gin"

	"jarvis/backend/internal/database"
	"jarvis/backend/internal/pkg/response"
)

type Handler struct {
	app *database.App
}

func New(app *database.App) *Handler {
	return &Handler{app: app}
}

func (h *Handler) Health(c *gin.Context) {
	response.OK(c, h.app.Health(c.Request.Context()))
}
