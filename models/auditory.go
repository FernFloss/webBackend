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

// GetLatestOccupancyByBuilding returns the most recent occupancy record for each
// auditorium in a building at or before the provided timestamp.
func (a *AuditoryModel) GetLatestOccupancyByBuilding(buildingID uint, queryTimestamp time.Time, maxTimeDiffMinutes int) ([]forms.AuditoriumOccupancyResponse, error) {
	type row struct {
		AuditoriumID uint
		PersonCount  int
		Timestamp    time.Time
	}

	var rows []row
	// subquery: latest timestamp per auditorium in building up to queryTimestamp
	sub := db.GetDB().Table("occupancy AS o").
		Select("o.auditorium_id, MAX(o.timestamp) AS max_ts").
		Joins("JOIN auditorium a ON a.id = o.auditorium_id").
		Where("a.building_id = ? AND o.timestamp <= ?", buildingID, queryTimestamp).
		Group("o.auditorium_id")

	result := db.GetDB().Table("occupancy AS o").
		Select("o.auditorium_id, o.person_count, o.timestamp").
		Joins("JOIN (?) latest ON latest.auditorium_id = o.auditorium_id AND latest.max_ts = o.timestamp", sub).
		Order("o.auditorium_id").
		Scan(&rows)
	if result.Error != nil {
		return nil, fmt.Errorf("database error: %w", result.Error)
	}
	if len(rows) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	responses := make([]forms.AuditoriumOccupancyResponse, 0, len(rows))
	for _, r := range rows {
		timeDiff := queryTimestamp.Sub(r.Timestamp).Minutes()
		isFresh := timeDiff <= float64(maxTimeDiffMinutes)
		var warning *string
		if !isFresh {
			msg := fmt.Sprintf("Data is stale by %.1f minutes (max %d)", timeDiff, maxTimeDiffMinutes)
			warning = &msg
		}
		responses = append(responses, forms.AuditoriumOccupancyResponse{
			AuditoriumID:    r.AuditoriumID,
			PersonCount:     r.PersonCount,
			ActualTimestamp: r.Timestamp,
			IsFresh:         isFresh,
			TimeDiffMinutes: timeDiff,
			Warning:         warning,
		})
	}
	return responses, nil
}

// GetLatestOccupancyForAuditorium returns the most recent occupancy record for a
// single auditorium at or before the provided timestamp.
func (a *AuditoryModel) GetLatestOccupancyForAuditorium(auditoriumID uint, queryTimestamp time.Time, maxTimeDiffMinutes int) (*forms.AuditoriumOccupancyResponse, error) {
	var row struct {
		AuditoriumID uint
		PersonCount  int
		Timestamp    time.Time
	}

	result := db.GetDB().Table("occupancy").
		Select("auditorium_id, person_count, timestamp").
		Where("auditorium_id = ? AND timestamp <= ?", auditoriumID, queryTimestamp).
		Order("timestamp DESC").
		Limit(1).
		Scan(&row)
	if result.Error != nil {
		return nil, fmt.Errorf("database error: %w", result.Error)
	}
	if row.AuditoriumID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	timeDiff := queryTimestamp.Sub(row.Timestamp).Minutes()
	isFresh := timeDiff <= float64(maxTimeDiffMinutes)
	var warning *string
	if !isFresh {
		msg := fmt.Sprintf("Data is stale by %.1f minutes (max %d)", timeDiff, maxTimeDiffMinutes)
		warning = &msg
	}

	resp := &forms.AuditoriumOccupancyResponse{
		AuditoriumID:    row.AuditoriumID,
		PersonCount:     row.PersonCount,
		ActualTimestamp: row.Timestamp,
		IsFresh:         isFresh,
		TimeDiffMinutes: timeDiff,
		Warning:         warning,
	}
	return resp, nil
}
