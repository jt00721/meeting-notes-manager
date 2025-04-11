package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jt00721/meeting-notes-manager/internal/domain"
	"github.com/jt00721/meeting-notes-manager/internal/usecase"
)

type NoteHandler struct {
	Usecase usecase.NoteUsecase
}

func NewNoteHandler(u usecase.NoteUsecase) *NoteHandler {
	return &NoteHandler{Usecase: u}
}

func (handler *NoteHandler) CreateNoteApi(c *gin.Context) {
	var note domain.Note
	if err := c.ShouldBindJSON(&note); err != nil {
		log.Printf("Error binding json request body to create note: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input to create note"})
		return
	}

	err := handler.Usecase.CreateNote(&note)
	if err != nil {
		if err.Error() == "note title cannot be empty" {
			log.Println("Error: Cannot create note without title")
			c.JSON(http.StatusBadRequest, gin.H{"error": "note title cannot be empty"})
			return
		} else if err.Error() == "note content cannot be empty" {
			log.Println("Error: Cannot create note without content")
			c.JSON(http.StatusBadRequest, gin.H{"error": "note content cannot be empty"})
			return
		}

		log.Printf("Error creating note: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create note. Please try again later."})
		return
	}

	log.Println("Successfully created note")
	c.JSON(http.StatusCreated, note)
}

func (handler *NoteHandler) GetAllNotesApi(c *gin.Context) {
	notes, err := handler.Usecase.GetAllNotes()
	if err != nil {
		log.Printf("Error retrieving all notes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve all notes. Please try again later.",
		})
		return
	}

	if len(notes) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No notes found",
			"notes":   notes,
		})
		return
	}

	log.Println("Successfully retrieved all notes")
	c.JSON(http.StatusOK, notes)
}

func (handler *NoteHandler) GetNoteByIDApi(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Error converting note ID URL query: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	note, err := handler.Usecase.GetNoteByID(uint(id))
	if err != nil {
		log.Printf("Error retrieving note with ID(%d): %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Note not found",
		})
		return
	}

	log.Println("Successfully retrieved note")
	c.JSON(http.StatusOK, note)
}

func (handler *NoteHandler) UpdateNoteApi(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Error converting note ID URL query: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	var note domain.Note
	if err := c.ShouldBindJSON(&note); err != nil {
		log.Printf("Error binding json request body to update note: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input to update note",
		})
		return
	}

	note.ID = uint(id)
	err = handler.Usecase.UpdateNote(&note)
	if err != nil {
		if err.Error() == "note title cannot be empty" {
			log.Println("Error: Cannot create note without title")
			c.JSON(http.StatusBadRequest, gin.H{"error": "note title cannot be empty"})
			return
		} else if err.Error() == "note content cannot be empty" {
			log.Println("Error: Cannot create note without content")
			c.JSON(http.StatusBadRequest, gin.H{"error": "note content cannot be empty"})
			return
		}

		log.Printf("Error updating note with ID(%d): %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update note. Please try again later."})
		return
	}

	log.Println("Successfully updated note")
	c.JSON(http.StatusOK, note)
}

func (handler *NoteHandler) DeleteNoteApi(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Error converting note ID URL query: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	err = handler.Usecase.DeleteNote(uint(id))
	if err != nil {
		log.Printf("Error deleting note with ID(%d): %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete note. Please try again later."})
		return
	}

	log.Println("Successfully deleted note")
	c.JSON(http.StatusOK, gin.H{"message": "Note deleted"})
}
