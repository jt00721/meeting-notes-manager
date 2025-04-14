package usecase_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jt00721/meeting-notes-manager/internal/domain"
	"github.com/jt00721/meeting-notes-manager/internal/usecase"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockNoteRepository struct {
	notes       []domain.Note
	forceDBFail bool
}

func (m *mockNoteRepository) Create(n *domain.Note) error {
	m.notes = append(m.notes, *n)
	return nil
}

// GetAll implements repository.NoteRepository.
func (m *mockNoteRepository) GetAll() ([]domain.Note, error) {
	if m.forceDBFail {
		return []domain.Note{}, errors.New("db error")
	}
	return m.notes, nil
}

// GetByID implements repository.NoteRepository.
func (m *mockNoteRepository) GetByID(id uint) (domain.Note, error) {
	// 1. Simulate hardcoded error (like db failure)
	if id == 3 {
		return domain.Note{}, errors.New("db error")
	}

	// 2. Look through mock slice for testable notes
	for _, n := range m.notes {
		if n.ID == id {
			return n, nil
		}
	}

	// 3. Default: not found
	return domain.Note{}, gorm.ErrRecordNotFound
}

// GetPaginated implements repository.NoteRepository.
func (m *mockNoteRepository) GetPaginated(limit int, offset int) ([]domain.Note, error) {
	panic("unimplemented")
}

// Update implements repository.NoteRepository.
func (m *mockNoteRepository) Update(n *domain.Note) error {
	if n.ID == 999 {
		return errors.New("db error")
	}
	return nil
}

func (m *mockNoteRepository) Delete(id uint) error {
	if m.forceDBFail {
		return errors.New("db error")
	}

	newNotes := make([]domain.Note, 0)
	for _, note := range m.notes {
		if note.ID != id {
			newNotes = append(newNotes, note)
		}
	}
	m.notes = newNotes
	return nil
}

// Search implements repository.NoteRepository.
func (m *mockNoteRepository) Search(keyword string) ([]domain.Note, error) {
	panic("unimplemented")
}

// Filter implements repository.NoteRepository.
func (m *mockNoteRepository) Filter(filter domain.NoteFilter) ([]domain.Note, error) {
	if m.forceDBFail {
		return nil, errors.New("db error")
	}

	var result []domain.Note
	for _, note := range m.notes {
		match := true

		if filter.Keyword != "" {
			keyword := strings.ToLower(filter.Keyword)
			if !strings.Contains(strings.ToLower(note.Title), keyword) &&
				!strings.Contains(strings.ToLower(note.Content), keyword) {
				match = false
			}
		}

		if filter.Category != "" && note.Category != filter.Category {
			match = false
		}

		if filter.FromDate != nil && note.MeetingDate.Before(*filter.FromDate) {
			match = false
		}

		if filter.ToDate != nil && note.MeetingDate.After(*filter.ToDate) {
			match = false
		}

		if match {
			result = append(result, note)
		}
	}

	return result, nil
}

func TestCreateNote(t *testing.T) {
	tests := []struct {
		name        string
		input       domain.Note
		wantErr     bool
		errContains error
	}{
		{
			name: "valid note",
			input: domain.Note{
				Title:   "Team Meeting",
				Content: "Discussed sprint planning",
			},
			wantErr: false,
		},
		{
			name: "empty title",
			input: domain.Note{
				Title:   "",
				Content: "Discussed sprint planning",
			},
			wantErr:     true,
			errContains: usecase.ErrEmptyTitle,
		},
		{
			name: "empty content",
			input: domain.Note{
				Title:   "Team Meeting",
				Content: "",
			},
			wantErr:     true,
			errContains: usecase.ErrEmptyContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockNoteRepository{}
			noteUC := usecase.NewNoteUsecase(mockRepo)
			err := noteUC.CreateNote(&tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains.Error())
				assert.Len(t, mockRepo.notes, 0)
			} else {
				assert.NoError(t, err)
				assert.Len(t, mockRepo.notes, 1)
			}
		})
	}
}

func TestGetAllNotes(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func() usecase.NoteUsecase
		wantErr     bool
		errContains error
	}{
		{
			name: "valid notes",
			setupRepo: func() usecase.NoteUsecase {
				mockRepo := &mockNoteRepository{
					notes: []domain.Note{
						{
							ID:          1,
							Title:       "Note Title 1",
							Content:     "Note Content 1",
							Category:    "Team Meeting",
							MeetingDate: time.Date(2025, time.October, 12, 11, 30, 0, 0, time.UTC),
						},
						{
							ID:          2,
							Title:       "Note Title 2",
							Content:     "Note Content 2",
							Category:    "Team Meeting",
							MeetingDate: time.Date(2025, time.July, 12, 11, 30, 0, 0, time.UTC),
						},
						{
							ID:          3,
							Title:       "Note Title 3",
							Content:     "Note Content 3",
							Category:    "Team Meeting",
							MeetingDate: time.Date(2025, time.June, 12, 11, 30, 0, 0, time.UTC),
						}}}
				return usecase.NewNoteUsecase(mockRepo)
			},
			wantErr: false,
		},
		{
			name: "no notes",
			setupRepo: func() usecase.NoteUsecase {
				return usecase.NewNoteUsecase(&mockNoteRepository{
					notes: []domain.Note{},
				})
			},
			wantErr: false,
		},
		{
			name: "repo error",
			setupRepo: func() usecase.NoteUsecase {
				mockRepo := &mockNoteRepository{
					notes: []domain.Note{
						{
							ID:          1,
							Title:       "Note Title 1",
							Content:     "Note Content 1",
							Category:    "Team Meeting",
							MeetingDate: time.Date(2025, time.October, 12, 11, 30, 0, 0, time.UTC),
						},
						{
							ID:          2,
							Title:       "Note Title 2",
							Content:     "Note Content 2",
							Category:    "Team Meeting",
							MeetingDate: time.Date(2025, time.July, 12, 11, 30, 0, 0, time.UTC),
						}},
					forceDBFail: true,
				}
				return usecase.NewNoteUsecase(mockRepo)
			},
			wantErr:     true,
			errContains: errors.New("failed to get notes"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			noteUC := tt.setupRepo()
			notes, err := noteUC.GetAllNotes()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains.Error())
			} else {
				assert.NoError(t, err)
				if tt.name == "no notes" {
					assert.Len(t, notes, 0)
				} else {
					assert.Len(t, notes, 3)
				}
			}
		})
	}
}

func TestGetNoteByID(t *testing.T) {
	tests := []struct {
		name        string
		input       uint
		wantErr     bool
		errContains error
	}{
		{
			name:    "valid ID",
			input:   1,
			wantErr: false,
		},
		{
			name:        "missing ID",
			input:       0,
			wantErr:     true,
			errContains: usecase.ErrNoteNotFound,
		},
		{
			name:        "repo error",
			input:       3,
			wantErr:     true,
			errContains: errors.New("failed to retrieve note"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockNoteRepository{
				notes: []domain.Note{{
					ID:      1,
					Title:   "Valid",
					Content: "Exists",
				}},
			}
			noteUC := usecase.NewNoteUsecase(mockRepo)
			note, err := noteUC.GetNoteByID(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, uint(1), note.ID)
			}
		})
	}
}

func TestUpdateNote(t *testing.T) {
	tests := []struct {
		name        string
		noteID      uint
		input       domain.Note
		setupRepo   func() usecase.NoteUsecase
		wantErr     bool
		errContains error
	}{
		{
			name:   "valid update",
			noteID: 1,
			input: domain.Note{
				ID:          1,
				Title:       "Team Standup",
				Content:     "Discussed issues that may affect other teams",
				Category:    "Standup",
				MeetingDate: time.Date(2025, time.June, 15, 10, 30, 0, 0, time.UTC),
			},
			setupRepo: func() usecase.NoteUsecase {
				mockRepo := &mockNoteRepository{
					notes: []domain.Note{{
						ID:          1,
						Title:       "Update Meeting Title",
						Content:     "Update Meeting Content",
						Category:    "Team Meeting",
						MeetingDate: time.Date(2025, time.October, 12, 11, 30, 0, 0, time.UTC),
					}}}
				return usecase.NewNoteUsecase(mockRepo)
			},
			wantErr: false,
		},
		{
			name:   "missing title",
			noteID: 1,
			input: domain.Note{
				ID:          1,
				Title:       "",
				Content:     "Discussed pay rise",
				Category:    "1:1",
				MeetingDate: time.Date(2025, time.September, 15, 10, 30, 0, 0, time.UTC),
			},
			setupRepo: func() usecase.NoteUsecase {
				mockRepo := &mockNoteRepository{
					notes: []domain.Note{{
						ID:          1,
						Title:       "Update Meeting Title",
						Content:     "Update Meeting Content",
						Category:    "Team Meeting",
						MeetingDate: time.Date(2025, time.October, 12, 11, 30, 0, 0, time.UTC),
					}}}
				return usecase.NewNoteUsecase(mockRepo)
			},
			wantErr:     true,
			errContains: usecase.ErrEmptyTitle,
		},
		{
			name:   "missing content",
			noteID: 1,
			input: domain.Note{
				ID:          1,
				Title:       "All-Hands",
				Content:     "",
				Category:    "Company-wide",
				MeetingDate: time.Date(2025, time.August, 15, 10, 30, 0, 0, time.UTC),
			},
			setupRepo: func() usecase.NoteUsecase {
				mockRepo := &mockNoteRepository{
					notes: []domain.Note{{
						ID:          1,
						Title:       "Update Meeting Title",
						Content:     "Update Meeting Content",
						Category:    "Team Meeting",
						MeetingDate: time.Date(2025, time.October, 12, 11, 30, 0, 0, time.UTC),
					}}}
				return usecase.NewNoteUsecase(mockRepo)
			},
			wantErr:     true,
			errContains: usecase.ErrEmptyContent,
		},
		{
			name:   "Note doesn't exist",
			noteID: 4,
			input: domain.Note{
				ID:          4,
				Title:       "All-Hands",
				Content:     "Discussed issues that may affect other teams",
				Category:    "Company-wide",
				MeetingDate: time.Date(2025, time.August, 15, 10, 30, 0, 0, time.UTC),
			},
			setupRepo: func() usecase.NoteUsecase {
				mockRepo := &mockNoteRepository{
					notes: []domain.Note{{
						ID:          1,
						Title:       "Update Meeting Title",
						Content:     "Update Meeting Content",
						Category:    "Team Meeting",
						MeetingDate: time.Date(2025, time.October, 12, 11, 30, 0, 0, time.UTC),
					}}}
				return usecase.NewNoteUsecase(mockRepo)
			},
			wantErr:     true,
			errContains: usecase.ErrNoteNotFound,
		},
		{
			name:   "repo fails",
			noteID: 999,
			input: domain.Note{
				ID:          999,
				Title:       "All-Hands",
				Content:     "Discussed issues that may affect other teams",
				Category:    "Company-wide",
				MeetingDate: time.Date(2025, time.August, 15, 10, 30, 0, 0, time.UTC),
			},
			setupRepo: func() usecase.NoteUsecase {
				mockRepo := &mockNoteRepository{
					notes: []domain.Note{{
						ID:          999,
						Title:       "Update Meeting Title",
						Content:     "Update Meeting Content",
						Category:    "Team Meeting",
						MeetingDate: time.Date(2025, time.October, 12, 11, 30, 0, 0, time.UTC),
					}}}
				return usecase.NewNoteUsecase(mockRepo)
			},
			wantErr:     true,
			errContains: errors.New("failed to update note"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			noteUC := tt.setupRepo()

			err := noteUC.UpdateNote(&tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteNote(t *testing.T) {
	tests := []struct {
		name        string
		input       uint
		setupRepo   func(repo **mockNoteRepository) usecase.NoteUsecase
		wantErr     bool
		errContains error
		expectLen   *int // only used for success case
	}{
		{
			name:  "valid ID",
			input: 1,
			setupRepo: func(repo **mockNoteRepository) usecase.NoteUsecase {
				*repo = &mockNoteRepository{
					notes: []domain.Note{{
						ID:      1,
						Title:   "Valid",
						Content: "Exists",
					}},
				}
				return usecase.NewNoteUsecase(*repo)
			},
			wantErr: false,
		},
		{
			name:  "missing ID",
			input: 0,
			setupRepo: func(repo **mockNoteRepository) usecase.NoteUsecase {
				*repo = &mockNoteRepository{
					notes: []domain.Note{{
						ID:      1,
						Title:   "Valid",
						Content: "Exists",
					}},
				}
				return usecase.NewNoteUsecase(*repo)
			},
			wantErr:     true,
			errContains: usecase.ErrNoteNotFound,
		},
		{
			name:  "repo error",
			input: 1,
			setupRepo: func(repo **mockNoteRepository) usecase.NoteUsecase {
				*repo = &mockNoteRepository{
					notes: []domain.Note{{
						ID:      1,
						Title:   "Valid",
						Content: "Exists",
					}},
					forceDBFail: true,
				}
				return usecase.NewNoteUsecase(*repo)
			},
			wantErr:     true,
			errContains: errors.New("failed to delete note"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var repo *mockNoteRepository
			noteUC := tt.setupRepo(&repo)

			err := noteUC.DeleteNote(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, repo.notes, 0)
			}
		})
	}
}

func TestFilterNotes(t *testing.T) {
	validFromDate := time.Date(2025, time.January, 12, 11, 30, 0, 0, time.UTC)
	validToDate := time.Date(2025, time.June, 12, 11, 30, 0, 0, time.UTC)
	tests := []struct {
		name        string
		input       domain.NoteFilter
		setupRepo   func() usecase.NoteUsecase
		wantLen     int
		wantErr     bool
		errContains error
	}{
		{
			name: "Valid: keyword only",
			input: domain.NoteFilter{
				Keyword:  "Title",
				Category: "",
				FromDate: nil,
				ToDate:   nil,
			},
			setupRepo: func() usecase.NoteUsecase {
				mockRepo := &mockNoteRepository{
					notes: []domain.Note{{
						ID:          1,
						Title:       "Test Meeting Title",
						Content:     "Test Meeting Content",
						Category:    "Team Meeting",
						MeetingDate: time.Date(2025, time.February, 12, 11, 30, 0, 0, time.UTC),
					},
						{
							ID:          2,
							Title:       "Test Meeting 2",
							Content:     "Test Meeting Content: Title",
							Category:    "Team Meeting",
							MeetingDate: time.Date(2025, time.March, 12, 11, 30, 0, 0, time.UTC),
						},
						{
							ID:          3,
							Title:       "Test Meeting 3",
							Content:     "Test Meeting Content",
							Category:    "1:1",
							MeetingDate: time.Date(2025, time.December, 12, 11, 30, 0, 0, time.UTC),
						},
					}}
				return usecase.NewNoteUsecase(mockRepo)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "Valid: category only",
			input: domain.NoteFilter{
				Keyword:  "",
				Category: "1:1",
				FromDate: nil,
				ToDate:   nil,
			},
			setupRepo: func() usecase.NoteUsecase {
				mockRepo := &mockNoteRepository{
					notes: []domain.Note{{
						ID:          1,
						Title:       "Test Meeting Title",
						Content:     "Test Meeting Content",
						Category:    "Team Meeting",
						MeetingDate: time.Date(2025, time.February, 12, 11, 30, 0, 0, time.UTC),
					},
						{
							ID:          2,
							Title:       "Test Meeting 2",
							Content:     "Test Meeting Content: Title",
							Category:    "Team Meeting",
							MeetingDate: time.Date(2025, time.March, 12, 11, 30, 0, 0, time.UTC),
						},
						{
							ID:          3,
							Title:       "Test Meeting 3",
							Content:     "Test Meeting Content",
							Category:    "1:1",
							MeetingDate: time.Date(2025, time.December, 12, 11, 30, 0, 0, time.UTC),
						},
					}}
				return usecase.NewNoteUsecase(mockRepo)
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "Valid: Full filter",
			input: domain.NoteFilter{
				Keyword:  "Title",
				Category: "Team Meeting",
				FromDate: &validFromDate,
				ToDate:   &validToDate,
			},
			setupRepo: func() usecase.NoteUsecase {
				mockRepo := &mockNoteRepository{
					notes: []domain.Note{{
						ID:          1,
						Title:       "Test Meeting Title",
						Content:     "Test Meeting Content",
						Category:    "Team Meeting",
						MeetingDate: time.Date(2025, time.February, 12, 11, 30, 0, 0, time.UTC),
					},
						{
							ID:          2,
							Title:       "Test Meeting 2",
							Content:     "Test Meeting Content: Title",
							Category:    "Team Meeting",
							MeetingDate: time.Date(2025, time.December, 12, 11, 30, 0, 0, time.UTC),
						},
						{
							ID:          3,
							Title:       "Test Meeting 3",
							Content:     "Test Meeting Content",
							Category:    "1:1",
							MeetingDate: time.Date(2025, time.March, 12, 11, 30, 0, 0, time.UTC),
						},
					}}
				return usecase.NewNoteUsecase(mockRepo)
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "Invalid: bad date range",
			input: domain.NoteFilter{
				Keyword:  "",
				Category: "",
				FromDate: &validToDate,
				ToDate:   &validFromDate,
			},
			setupRepo: func() usecase.NoteUsecase {
				mockRepo := &mockNoteRepository{
					notes: []domain.Note{{
						ID:          1,
						Title:       "Update Meeting Title",
						Content:     "Update Meeting Content",
						Category:    "Team Meeting",
						MeetingDate: time.Date(2025, time.October, 12, 11, 30, 0, 0, time.UTC),
					}}}
				return usecase.NewNoteUsecase(mockRepo)
			},
			wantErr:     true,
			errContains: errors.New("fromDate must be before toDate"),
		},
		{
			name: "Repo fails",
			input: domain.NoteFilter{
				Keyword:  "Title",
				Category: "Team meeting",
				FromDate: &validFromDate,
				ToDate:   &validToDate,
			},
			setupRepo: func() usecase.NoteUsecase {
				mockRepo := &mockNoteRepository{
					notes: []domain.Note{{
						ID:          999,
						Title:       "Update Meeting Title",
						Content:     "Update Meeting Content",
						Category:    "Team Meeting",
						MeetingDate: time.Date(2025, time.October, 12, 11, 30, 0, 0, time.UTC),
					}},
					forceDBFail: true,
				}
				return usecase.NewNoteUsecase(mockRepo)
			},
			wantErr:     true,
			errContains: errors.New("failed to filter notes"),
		},
		{
			name: "No results",
			input: domain.NoteFilter{
				Keyword:  "Title",
				Category: "Team meeting",
				FromDate: &validFromDate,
				ToDate:   &validToDate,
			},
			setupRepo: func() usecase.NoteUsecase {
				mockRepo := &mockNoteRepository{
					notes: []domain.Note{},
				}
				return usecase.NewNoteUsecase(mockRepo)
			},
			wantLen: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			noteUC := tt.setupRepo()

			searchResults, err := noteUC.FilterNotes(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, searchResults, tt.wantLen)
			}
		})
	}
}
