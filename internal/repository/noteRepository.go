package repository

import (
	"github.com/jt00721/meeting-notes-manager/internal/domain"
	"gorm.io/gorm"
)

type NoteRepository interface {
	Create(n *domain.Note) error
	GetAll() ([]domain.Note, error)
	GetPaginated(limit, offset int) ([]domain.Note, error)
	GetByID(id uint) (domain.Note, error)
	Update(n *domain.Note) error
	Delete(id uint) error
	Search(keyword string) ([]domain.Note, error)
	Filter(filter domain.NoteFilter) ([]domain.Note, error)
}

type noteRepository struct {
	DB *gorm.DB
}

func NewNoteRepository(DB *gorm.DB) *noteRepository {
	return &noteRepository{DB: DB}
}

func (r *noteRepository) Create(n *domain.Note) error {
	return r.DB.Create(n).Error
}

func (r *noteRepository) GetAll() ([]domain.Note, error) {
	var notes []domain.Note
	err := r.DB.Find(&notes).Error
	return notes, err
}

func (r *noteRepository) GetPaginated(limit, offset int) ([]domain.Note, error) {
	var notes []domain.Note
	err := r.DB.Limit(limit).Offset(offset).Find(&notes).Error
	return notes, err
}

func (r *noteRepository) GetByID(id uint) (domain.Note, error) {
	var note domain.Note
	err := r.DB.First(&note, id).Error
	return note, err
}

func (r *noteRepository) Update(n *domain.Note) error {
	return r.DB.Save(n).Error
}

func (r *noteRepository) Delete(id uint) error {
	return r.DB.Delete(&domain.Note{}, id).Error
}

func (r *noteRepository) Search(keyword string) ([]domain.Note, error) {
	var notes []domain.Note
	err := r.DB.
		Where("title ILIKE ? OR content ILIKE ?", "%"+keyword+"%", "%"+keyword+"%").
		Find(&notes).Error
	return notes, err
}

func (r *noteRepository) Filter(filter domain.NoteFilter) ([]domain.Note, error) {
	var notes []domain.Note

	tx := r.DB // Start building the query

	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		tx = tx.Where("title ILIKE ? OR content ILIKE ?", like, like)
	}

	if filter.Category != "" {
		tx = tx.Where("category = ?", filter.Category)
	}

	if filter.FromDate != nil {
		tx = tx.Where("meeting_date >= ?", *filter.FromDate)
	}

	if filter.ToDate != nil {
		tx = tx.Where("meeting_date <= ?", *filter.ToDate)
	}

	err := tx.Find(&notes).Error
	return notes, err
}
