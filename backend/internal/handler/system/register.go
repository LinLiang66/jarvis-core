package system

import (
	"github.com/gin-gonic/gin"

	"jarvis-core/backend/internal/config"
	"jarvis-core/backend/internal/database"
)

func Register(rg *gin.RouterGroup, app *database.App, cfg *config.Config) {
	NewUserHandler(app).Register(rg)
	NewRoleHandler(app).Register(rg)
	NewMenuHandler(app).Register(rg)
	NewDictHandler(app).Register(rg)
	NewStorageHandler(app, cfg).Register(rg)
	NewFileHandler(app, cfg).Register(rg)
}
