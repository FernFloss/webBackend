package models

import (
	"fmt"
	"time"
	"web_backend_v2/db"
	"web_backend_v2/forms"

	"gorm.io/gorm"
)

type AuditoryModel struct{}

func (a *AuditoryModel) GetAuditoriumsByBuilding(uidBuilding uint) ([]forms.Auditorium, error) {
	var auditories []forms.Auditorium

	result := db.GetDB().Table("auditorium").
		Where("building_id = ?", uidBuilding).
		Order("id").
		Find(&auditories)

	if result.Error != nil {
		return nil, fmt.Errorf("error fetching audotories for city %d: %w", uidBuilding, result.Error)
	}
	if len(auditories) < 1 {
		return nil, fmt.Errorf("auditories with building ID:%d do not exist", uidBuilding)
	}

	return auditories, nil
}

func (a *AuditoryModel) GetOccupancyForAuditorium(auditoriumID uint, queryTimestamp time.Time, maxTimeDiffMinutes int) (forms.Occupancy, error) {
	var occupancy forms.Occupancy
	result := db.GetDB().Table("occupancy").
		Where("auditorium_id = ? AND timestamp <= ?", auditoriumID, queryTimestamp).
		Order("timestamp DESC").
		First(&occupancy)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return occupancy, fmt.Errorf("no occupancy records found for auditorium %d before %s",
				auditoriumID, queryTimestamp.Format(time.RFC3339))
		}
		return occupancy, fmt.Errorf("database error: %w", result.Error)
	}
	return occupancy, nil
}
