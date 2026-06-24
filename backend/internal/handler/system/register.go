package system

import (
	"github.com/gin-gonic/gin"

	"jarvis/backend/internal/database"
)

func Register(rg *gin.RouterGroup, app *database.App) {
	NewUserHandler(app).Register(rg)
	NewRoleHandler(app).Register(rg)
	NewMenuHandler(app).Register(rg)
	NewDictHandler(app).Register(rg)
}
