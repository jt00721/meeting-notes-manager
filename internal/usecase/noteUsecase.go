package usecase

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/jt00721/meeting-notes-manager/internal/domain"
	"github.com/jt00721/meeting-notes-manager/internal/repository"
	"gorm.io/gorm"
)

type NoteUsecase interface {
	CreateNote(n *domain.Note) error
	GetAllNotes() ([]domain.Note, error)
	GetPaginatedNotes(limit, offset int) ([]domain.Note, error)
	GetNoteByID(id uint) (domain.Note, error)
	UpdateNote(n *domain.Note) error
	DeleteNote(id uint) error
	SearchNotesByKeyword(keyword string) ([]domain.Note, error)
	FilterNotes(filter domain.NoteFilter) ([]domain.Note, error)
}

type noteUsecase struct {
	repo repository.NoteRepository
}

func NewNoteUsecase(r repository.NoteRepository) *noteUsecase {
	return &noteUsecase{repo: r}
}

func (uc *noteUsecase) CreateNote(n *domain.Note) error {
	if n.Title == "" {
		return ErrEmptyTitle
	}

	if n.Content == "" {
		return ErrEmptyContent
	}

	if err := uc.repo.Create(n); err != nil {
		log.Println("Error creating note:", err)
		return fmt.Errorf("failed to create note")
	}

	log.Printf("Note (%d) created successfully", n.ID)
	return nil
}

func (uc *noteUsecase) GetAllNotes() ([]domain.Note, error) {
	notes, err := uc.repo.GetAll()
	if err != nil {
		log.Println("Error retrieving all notes:", err)
		return nil, fmt.Errorf("failed to get notes")
	}

	sort.Slice(notes, func(i, j int) bool {
		return notes[i].MeetingDate.After(notes[j].MeetingDate)
	})

	log.Println("All notes retrieved successfully")
	return notes, nil
}

func (uc *noteUsecase) GetPaginatedNotes(limit, offset int) ([]domain.Note, error) {
	notes, err := uc.repo.GetPaginated(limit, offset)
	if err != nil {
		log.Println("Error retrieving paginated notes:", err)
		return nil, fmt.Errorf("failed to get notes")
	}

	sort.Slice(notes, func(i, j int) bool {
		return notes[i].MeetingDate.After(notes[j].MeetingDate)
	})

	log.Println("Paginated notes retrieved successfully")
	return notes, nil
}

func (uc *noteUsecase) GetNoteByID(id uint) (domain.Note, error) {
	note, err := uc.repo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return domain.Note{}, ErrNoteNotFound
		}
		log.Printf("Error retrieving note with ID(%d): %v", id, err)
		return domain.Note{}, fmt.Errorf("failed to retrieve note")
	}

	log.Printf("Note (%d) retrieved successfully", note.ID)
	return note, nil
}

func (uc *noteUsecase) UpdateNote(n *domain.Note) error {
	existingNote, err := uc.GetNoteByID(n.ID)
	if err != nil {
		log.Println("Error retrieving note while trying to update note:", err)
		return ErrNoteNotFound
	}

	if n.Title == "" {
		return ErrEmptyTitle
	}

	if n.Content == "" {
		return ErrEmptyContent
	}

	existingNote.Title = n.Title
	existingNote.Content = n.Content
	existingNote.Category = n.Category
	existingNote.MeetingDate = n.MeetingDate

	err = uc.repo.Update(&existingNote)
	if err != nil {
		log.Printf("Error updating note with ID(%d): %v", n.ID, err)
		return fmt.Errorf("failed to update note")
	}

	log.Printf("Note (%d) updated successfully", n.ID)
	return nil
}

func (uc *noteUsecase) DeleteNote(id uint) error {
	if _, err := uc.GetNoteByID(id); err != nil {
		log.Println("Error: Tried to delete non-existing note with ID:", id)
		return ErrNoteNotFound
	}

	err := uc.repo.Delete(id)
	if err != nil {
		log.Println("Error deleting note:", err)
		return fmt.Errorf("failed to delete note")
	}

	log.Println("Note deleted successfully")
	return nil
}

func (uc *noteUsecase) SearchNotesByKeyword(keyword string) ([]domain.Note, error) {
	if strings.TrimSpace(keyword) == "" {
		return nil, fmt.Errorf("search keyword cannot be empty")
	}

	searchResult, err := uc.repo.Search(keyword)
	if err != nil {
		log.Printf("Error searching for notes with keyword (%s): %v", keyword, err)
		return nil, fmt.Errorf("failed to find notes")
	}

	sort.Slice(searchResult, func(i, j int) bool {
		return searchResult[i].MeetingDate.After(searchResult[j].MeetingDate)
	})

	log.Println("Successful Search")
	return searchResult, nil
}

func (uc *noteUsecase) FilterNotes(filter domain.NoteFilter) ([]domain.Note, error) {
	filter.Keyword = strings.TrimSpace(filter.Keyword)

	filter.Category = strings.TrimSpace(filter.Category)

	if filter.FromDate != nil && filter.ToDate != nil {
		if filter.FromDate.After(*filter.ToDate) {
			return nil, fmt.Errorf("fromDate must be before toDate")
		}
	}

	filterResults, err := uc.repo.Filter(filter)
	if err != nil {
		log.Printf("Error filtering for notes: %v", err)
		return nil, fmt.Errorf("failed to filter notes")
	}

	sort.Slice(filterResults, func(i, j int) bool {
		return filterResults[i].MeetingDate.After(filterResults[j].MeetingDate)
	})

	log.Println("Successful Filter")
	return filterResults, nil
}
