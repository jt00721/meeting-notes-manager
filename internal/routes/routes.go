package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jt00721/meeting-notes-manager/internal/handler"
)

func SetupRoutes(r *gin.Engine, noteHandler *handler.NoteHandler) {
	r.POST("/notes", noteHandler.CreateNoteApi)
	r.GET("/notes", noteHandler.GetAllNotesApi)
	r.GET("/notes/:id", noteHandler.GetNoteByIDApi)
	r.PUT("/notes/:id", noteHandler.UpdateNoteApi)
	r.DELETE("/notes/:id", noteHandler.DeleteNoteApi)
}
