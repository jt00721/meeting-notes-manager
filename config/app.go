package config

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/jt00721/meeting-notes-manager/infrastructure"
	"github.com/jt00721/meeting-notes-manager/internal/handler"
	"github.com/jt00721/meeting-notes-manager/internal/repository"
	"github.com/jt00721/meeting-notes-manager/internal/routes"
	"github.com/jt00721/meeting-notes-manager/internal/usecase"
)

type App struct {
	Router      *gin.Engine
	NoteHandler *handler.NoteHandler
}

func NewApp() *App {
	log.Println("Starting app...")
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Could not load .env file, using system environment variables")
	}

	log.Println("Initialising DB...")
	if err := infrastructure.InitDB(); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	noteRepository := repository.NewNoteRepository(infrastructure.DB)
	noteUsecase := usecase.NewNoteUsecase(noteRepository)
	noteHandler := handler.NewNoteHandler(noteUsecase)

	router := gin.Default()

	router.Static("/static", "./static")

	routes.SetupRoutes(router, noteHandler)

	return &App{
		Router:      router,
		NoteHandler: noteHandler,
	}
}

// Run starts the server
func (app *App) Run() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ip := "0.0.0.0:"
	if env := os.Getenv("ENV"); env == "Dev" || env == "development" {
		ip = ":"
	}

	fmt.Println("Server running on port", port)
	app.Router.Run(ip + port)
}
