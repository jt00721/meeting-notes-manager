package repository

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/jt00721/meeting-notes-manager/internal/domain"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var testRepo *noteRepository
var DB *gorm.DB

func SetupTestDB(m *testing.M) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("Failed to load env variables for Repo test")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_TEST_DB"),
		os.Getenv("POSTGRES_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to test DB:", err)
	}

	err = db.AutoMigrate(&domain.Note{})
	if err != nil {
		log.Fatal("Failed to migrate schema:", err)
	}

	DB = db

	testRepo = NewNoteRepository(DB)

	code := m.Run()

	os.Exit(code)
}

func cleanDB(t *testing.T) {
	err := DB.Exec("TRUNCATE notes RESTART IDENTITY CASCADE").Error
	assert.NoError(t, err)
}

func TestMain(m *testing.M) {
	SetupTestDB(m)
}

func TestCreate(t *testing.T) {
	cleanDB(t)

	note := domain.Note{
		Title:       "Test Meeting",
		Content:     "Some notes",
		Category:    "Planning",
		MeetingDate: time.Now(),
	}

	err := testRepo.Create(&note)
	assert.NoError(t, err)
	assert.NotZero(t, note.ID)
}

func TestGetByID(t *testing.T) {
	cleanDB(t)

	note := domain.Note{
		Title:       "Test Meeting",
		Content:     "Some notes",
		Category:    "Planning",
		MeetingDate: time.Now(),
	}

	err := testRepo.Create(&note)
	assert.NoError(t, err)

	fetchedNote, err := testRepo.GetByID(note.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Test Meeting", fetchedNote.Title)
}

func TestGetAll(t *testing.T) {
	cleanDB(t)

	testRepo.Create(&domain.Note{
		Title:       "Test Meeting 1",
		Content:     "Some notes",
		Category:    "Planning",
		MeetingDate: time.Now(),
	})

	testRepo.Create(&domain.Note{
		Title:       "Test Meeting 2",
		Content:     "Some notes",
		Category:    "1:1",
		MeetingDate: time.Now(),
	})

	testRepo.Create(&domain.Note{
		Title:       "Test Meeting 3",
		Content:     "Some notes",
		Category:    "Standup",
		MeetingDate: time.Now(),
	})

	notes, err := testRepo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, notes, 3)
}

func TestUpdate(t *testing.T) {
	cleanDB(t)

	note := domain.Note{
		Title:       "Test Meeting",
		Content:     "Some notes",
		Category:    "Planning",
		MeetingDate: time.Now(),
	}

	err := testRepo.Create(&note)
	assert.NoError(t, err)

	createdNote := domain.Note{
		ID:          note.ID,
		Title:       "Updated Test Meeting",
		Content:     "Updated notes",
		Category:    "Updated category",
		MeetingDate: time.Date(2025, time.June, 15, 10, 30, 0, 0, time.UTC),
	}

	err = testRepo.Update(&createdNote)
	assert.NoError(t, err)

	updatedNote, err := testRepo.GetByID(note.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Test Meeting", updatedNote.Title)
}

func TestDelete(t *testing.T) {
	cleanDB(t)

	note := domain.Note{
		Title:       "Test Meeting",
		Content:     "Some notes",
		Category:    "Planning",
		MeetingDate: time.Now(),
	}

	err := testRepo.Create(&note)
	assert.NoError(t, err)

	err = testRepo.Delete(note.ID)
	assert.NoError(t, err)

	notes, err := testRepo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, notes, 0)
}

func TestFilter(t *testing.T) {
	cleanDB(t)

	validFromDate := time.Date(2025, time.May, 12, 11, 30, 0, 0, time.UTC)
	validToDate := time.Date(2025, time.July, 12, 11, 30, 0, 0, time.UTC)

	testRepo.Create(&domain.Note{
		Title:       "Test Meeting 1",
		Content:     "Keyword in notes",
		Category:    "Planning",
		MeetingDate: time.Now(),
	})

	testRepo.Create(&domain.Note{
		Title:       "Test Meeting 2",
		Content:     "Some notes",
		Category:    "1:1",
		MeetingDate: time.Date(2025, time.June, 15, 10, 30, 0, 0, time.UTC),
	})

	testRepo.Create(&domain.Note{
		Title:       "Test Meeting 3",
		Content:     "Some notes",
		Category:    "Standup",
		MeetingDate: time.Date(2025, time.October, 17, 12, 30, 0, 0, time.UTC),
	})

	tests := []struct {
		name    string
		input   domain.NoteFilter
		wantLen int
	}{
		{
			name: "Keyword only",
			input: domain.NoteFilter{
				Keyword:  "Keyword",
				Category: "",
				FromDate: nil,
				ToDate:   nil,
			},
			wantLen: 1,
		},
		{
			name: "Category only",
			input: domain.NoteFilter{
				Keyword:  "",
				Category: "Standup",
				FromDate: nil,
				ToDate:   nil,
			},
			wantLen: 1,
		},
		{
			name: "Date range only",
			input: domain.NoteFilter{
				Keyword:  "",
				Category: "",
				FromDate: &validFromDate,
				ToDate:   &validToDate,
			},
			wantLen: 1,
		},
		{
			name: "Combined filters (keyword + category + date)",
			input: domain.NoteFilter{
				Keyword:  "Test",
				Category: "1:1",
				FromDate: &validFromDate,
				ToDate:   &validToDate,
			},
			wantLen: 1,
		},
		{
			name: "No match",
			input: domain.NoteFilter{
				Keyword:  "None",
				Category: "N/A",
				FromDate: &validFromDate,
				ToDate:   &validToDate,
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searchResults, err := testRepo.Filter(tt.input)
			assert.NoError(t, err)
			assert.Len(t, searchResults, tt.wantLen)
		})
	}
}
