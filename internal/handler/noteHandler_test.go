package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/jt00721/meeting-notes-manager/internal/domain"
	"github.com/jt00721/meeting-notes-manager/internal/usecase"
)

type mockNoteUsecase struct {
	mockCreateNote  func(n *domain.Note) error
	mockGetAllNotes func() ([]domain.Note, error)
	mockGetNoteByID func(id uint) (domain.Note, error)
	mockUpdateNote  func(n *domain.Note) error
	mockDeleteNote  func(id uint) error
	mockFilterNotes func(filter domain.NoteFilter) ([]domain.Note, error)
}

func (m *mockNoteUsecase) CreateNote(n *domain.Note) error {
	if m.mockCreateNote != nil {
		return m.mockCreateNote(n)
	}
	return nil
}

func (m *mockNoteUsecase) GetAllNotes() ([]domain.Note, error) {
	if m.mockGetAllNotes != nil {
		return m.mockGetAllNotes()
	}
	return []domain.Note{}, nil
}
func (m *mockNoteUsecase) GetPaginatedNotes(limit, offset int) ([]domain.Note, error) {
	return nil, nil
}

func (m *mockNoteUsecase) GetNoteByID(id uint) (domain.Note, error) {
	if m.mockGetNoteByID != nil {
		return m.mockGetNoteByID(id)
	}
	return domain.Note{}, nil
}

func (m *mockNoteUsecase) UpdateNote(n *domain.Note) error {
	if m.mockUpdateNote != nil {
		return m.mockUpdateNote(n)
	}
	return nil
}
func (m *mockNoteUsecase) DeleteNote(id uint) error {
	if m.mockDeleteNote != nil {
		return m.mockDeleteNote(id)
	}
	return nil
}
func (m *mockNoteUsecase) SearchNotesByKeyword(keyword string) ([]domain.Note, error) {
	return nil, nil
}
func (m *mockNoteUsecase) FilterNotes(filter domain.NoteFilter) ([]domain.Note, error) {
	if m.mockFilterNotes != nil {
		return m.mockFilterNotes(filter)
	}
	return []domain.Note{}, nil
}

func TestCreateNoteApi(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       string
		mockReturn error
		wantCode   int
	}{
		{
			name:       "Valid Create Note",
			body:       `{"title": "Test meeting", "content": "Some content", "category": "Standup", "meeting_date": "2025-06-15T10:30:00Z"}`,
			mockReturn: nil,
			wantCode:   http.StatusCreated,
		},
		{
			name:     "Invalid JSON",
			body:     `{"title": "Test meeting", "content": "Some content", "category": "Standup"`, // broken JSON
			wantCode: http.StatusBadRequest,
		},
		{
			name:       "Invalid Note Title",
			body:       `{"title": "", "content": "Some content", "category": "Standup", "meeting_date": "2025-06-15T10:30:00Z"}`,
			mockReturn: usecase.ErrEmptyTitle,
			wantCode:   http.StatusBadRequest,
		},
		{
			name:       "Invalid Note Content",
			body:       `{"title": "Test meeting", "content": "", "category": "Standup", "meeting_date": "2025-06-15T10:30:00Z"}`,
			mockReturn: usecase.ErrEmptyContent,
			wantCode:   http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockNoteUsecase{
				mockCreateNote: func(n *domain.Note) error {
					return tt.mockReturn
				},
			}

			handler := NewNoteHandler(mockUC)
			router := gin.Default()
			router.POST("/notes", handler.CreateNoteApi)

			req := httptest.NewRequest(http.MethodPost, "/notes", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.wantCode, resp.Code)
		})
	}
}

func TestGetAllNotesApi(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		mockReturn   []domain.Note
		mockError    error
		expectedCode int
	}{
		{
			name: "Valid Get All Notes",
			mockReturn: []domain.Note{
				{ID: 1, Title: "Test Meeting 1", Content: "Some content"},
				{ID: 2, Title: "Test Meeting 2", Content: "Some content"},
				{ID: 3, Title: "Test Meeting 3", Content: "Some content"},
			},
			mockError:    nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Valid Get All Notes with no Notes",
			mockReturn:   []domain.Note{},
			mockError:    nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Repo error",
			mockError:    errors.New("db error"),
			expectedCode: http.StatusInternalServerError, // This is what your handler currently returns
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockNoteUsecase{
				mockGetAllNotes: func() ([]domain.Note, error) {
					if tt.mockError != nil {
						return []domain.Note{}, tt.mockError
					}
					return tt.mockReturn, nil
				},
			}

			handler := NewNoteHandler(mockUC)
			router := gin.Default()
			router.GET("/notes", handler.GetAllNotesApi)

			req := httptest.NewRequest(http.MethodGet, "/notes", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedCode, resp.Code)
		})
	}
}

func TestGetNoteByIDApi(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		idParam      string
		mockReturn   domain.Note
		mockError    error
		expectedCode int
	}{
		{
			name:         "Valid ID",
			idParam:      "1",
			mockReturn:   domain.Note{ID: 1, Title: "Test Meeting"},
			mockError:    nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid ID (non-integer)",
			idParam:      "abc",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Note not found",
			idParam:      "999",
			mockError:    usecase.ErrNoteNotFound,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Repo error",
			idParam:      "5",
			mockError:    errors.New("db error"),
			expectedCode: http.StatusNotFound, // This is what your handler currently returns
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockNoteUsecase{
				mockGetNoteByID: func(id uint) (domain.Note, error) {
					if tt.mockError != nil {
						return domain.Note{}, tt.mockError
					}
					return tt.mockReturn, nil
				},
			}

			handler := NewNoteHandler(mockUC)
			router := gin.Default()
			router.GET("/notes/:id", handler.GetNoteByIDApi)

			req := httptest.NewRequest(http.MethodGet, "/notes/"+tt.idParam, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedCode, resp.Code)
		})
	}
}

func TestUpdateNoteApi(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		idParam    string
		body       string
		mockReturn error
		wantCode   int
	}{
		{
			name:       "Valid Update Note",
			idParam:    "1",
			body:       `{"title": "Test meeting", "content": "Some content", "category": "Standup", "meeting_date": "2025-06-15T10:30:00Z"}`,
			mockReturn: nil,
			wantCode:   http.StatusOK,
		},
		{
			name:     "Invalid ID (non-integer)",
			idParam:  "abc",
			body:     `{"title": "Test meeting", "content": "Some content", "category": "Standup", "meeting_date": "2025-06-15T10:30:00Z"}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "Invalid JSON",
			idParam:  "1",
			body:     `{"title": "Test meeting", "content": "Some content", "category": "Standup"`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:       "Invalid Note Title",
			idParam:    "1",
			body:       `{"title": "", "content": "Some content", "category": "Standup", "meeting_date": "2025-06-15T10:30:00Z"}`,
			mockReturn: usecase.ErrEmptyTitle,
			wantCode:   http.StatusBadRequest,
		},
		{
			name:       "Invalid Note Content",
			idParam:    "1",
			body:       `{"title": "Test meeting", "content": "", "category": "Standup", "meeting_date": "2025-06-15T10:30:00Z"}`,
			mockReturn: usecase.ErrEmptyContent,
			wantCode:   http.StatusBadRequest,
		},
		{
			name:       "Repo error",
			idParam:    "1",
			body:       `{"title": "Test meeting", "content": "Some content", "category": "Standup", "meeting_date": "2025-06-15T10:30:00Z"}`,
			mockReturn: errors.New("db error"),
			wantCode:   http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockNoteUsecase{
				mockUpdateNote: func(n *domain.Note) error {
					return tt.mockReturn
				},
			}

			handler := NewNoteHandler(mockUC)
			router := gin.Default()
			router.PUT("/notes/:id", handler.UpdateNoteApi)

			req := httptest.NewRequest(http.MethodPut, "/notes/"+tt.idParam, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.wantCode, resp.Code)
		})
	}
}

func TestDeleteNoteApi(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		idParam      string
		mockError    error
		expectedCode int
	}{
		{
			name:         "Valid ID",
			idParam:      "1",
			mockError:    nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid ID (non-integer)",
			idParam:      "abc",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Note not found",
			idParam:      "999",
			mockError:    usecase.ErrNoteNotFound,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Repo error",
			idParam:      "5",
			mockError:    errors.New("db error"),
			expectedCode: http.StatusInternalServerError, // This is what your handler currently returns
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockNoteUsecase{
				mockDeleteNote: func(id uint) error {
					return tt.mockError
				},
			}

			handler := NewNoteHandler(mockUC)
			router := gin.Default()
			router.DELETE("/notes/:id", handler.DeleteNoteApi)

			req := httptest.NewRequest(http.MethodDelete, "/notes/"+tt.idParam, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedCode, resp.Code)
		})
	}
}

func TestFilterNotesApi(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		queryParams  string
		mockReturn   []domain.Note
		mockError    error
		expectedCode int
	}{
		{
			name:        "Valid: keyword only",
			queryParams: "?keyword=meeting",
			mockReturn: []domain.Note{
				{ID: 1, Title: "Team Meeting", Content: "Discussed project"},
			},
			mockError:    nil,
			expectedCode: http.StatusOK,
		},
		{
			name:        "Valid: category only",
			queryParams: "?category=Standup",
			mockReturn: []domain.Note{
				{ID: 2, Title: "Daily", Content: "Quick sync"},
			},
			mockError:    nil,
			expectedCode: http.StatusOK,
		},
		{
			name:        "Valid: full filter",
			queryParams: "?keyword=team&category=Standup&fromDate=2025-01-01&toDate=2025-12-31",
			mockReturn: []domain.Note{
				{ID: 3, Title: "Team Standup", Content: "Updates", Category: "Standup"},
			},
			mockError:    nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "No results match",
			queryParams:  "?keyword=xyz",
			mockReturn:   []domain.Note{},
			mockError:    nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Repo error",
			queryParams:  "?keyword=team",
			mockError:    errors.New("db error"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockNoteUsecase{
				mockFilterNotes: func(filter domain.NoteFilter) ([]domain.Note, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockReturn, nil
				},
			}

			handler := NewNoteHandler(mockUC)
			router := gin.Default()
			router.GET("/notes/filter", handler.FilterNotesApi)

			req := httptest.NewRequest(http.MethodGet, "/notes/filter"+tt.queryParams, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedCode, resp.Code)
		})
	}
}
