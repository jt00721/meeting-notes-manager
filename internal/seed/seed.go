package seed

import (
	"log"
	"time"

	"github.com/jt00721/meeting-notes-manager/internal/domain"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	notes := []domain.Note{
		{Title: "Performance Review", Content: "Went over my performance over the year with my boss", Category: "1:1", MeetingDate: time.Date(2025, time.April, 1, 13, 30, 0, 0, time.UTC)},
		{Title: "Team Standup", Content: "Went over items in the current sprint", Category: "Standup", MeetingDate: time.Date(2025, time.May, 10, 14, 30, 0, 0, time.UTC)},
		{Title: "All-Hands Meeting", Content: "Quarterly meeting covering recent company news or updates", Category: "Company-wide", MeetingDate: time.Date(2025, time.June, 15, 10, 30, 0, 0, time.UTC)},
	}

	for _, note := range notes {
		if err := db.Create(&note).Error; err != nil {
			log.Println("Failed to seed note:", note.Title, "Error:", err)
			return err
		}
	}

	log.Println("Seeded initial notes successfully")
	return nil
}
