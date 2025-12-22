package models

import (
	"fmt"
	"time"
	"web_backend_v2/db"
	"web_backend_v2/forms"

	"gorm.io/gorm"
)

type AuditoryModel struct{}

// Exists checks if auditorium with given ID exists.
func (a *AuditoryModel) Exists(auditoriumID uint) (bool, error) {
	var count int64
	if err := db.GetDB().Table("auditorium").Where("id = ?", auditoriumID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("error checking auditorium existence: %w", err)
	}
	return count > 0, nil
}

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

// GetAuditoriumStats returns hourly statistics for a specific auditorium on a specific day.
// statsType: 1 = Absolute Count (Average), 2 = Occupancy Rate (Percentage)
// Returns strictly hours 9 to 21.
// The boolean flag noData is true when neither aggregated nor raw data exist for that day.
func (a *AuditoryModel) GetAuditoriumStats(auditoriumID uint, day time.Time, statsType int) ([]forms.HourlyStatsResponse, bool, error) {
	// Initialize map for hours 9-21
	statsMap := make(map[int]float64)
	// We want to return data for 9..21, but if no data exists, we might return 0.

	// Ensure day is at 00:00:00
	startOfDay := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.AddDate(0, 0, 1)

	// 1. Query DailyLoad (aggregated data)
	var dailyRows []forms.DailyLoad
	err := db.GetDB().Table("dailyload").
		Where("auditorium_id = ? AND day = ?", auditoriumID, startOfDay).
		Find(&dailyRows).Error
	if err != nil {
		return nil, false, fmt.Errorf("error fetching dailyload: %w", err)
	}

	for _, r := range dailyRows {
		if r.Hour >= 9 && r.Hour <= 21 {
			statsMap[r.Hour] = r.AvgPersonCount
		}
	}

	// 2. Query Occupancy (raw data, typically for today)
	type Result struct {
		Hour int
		Avg  float64
	}
	var occupancyRows []Result
	// We use EXTRACT(HOUR FROM timestamp) which depends on DB timezone. Assuming consistency.
	err = db.GetDB().Table("occupancy").
		Select("EXTRACT(hour FROM timestamp)::int as hour, AVG(person_count)::float8 as avg").
		Where("auditorium_id = ? AND timestamp >= ? AND timestamp < ?", auditoriumID, startOfDay, endOfDay).
		Group("EXTRACT(hour FROM timestamp)").
		Scan(&occupancyRows).Error
	if err != nil {
		return nil, false, fmt.Errorf("error fetching occupancy stats: %w", err)
	}

	for _, r := range occupancyRows {
		if r.Hour >= 9 && r.Hour <= 21 {
			// Overwrite if exists (raw data assumed more precise/current if overlap,
			// though overlap shouldn't exist due to aggregation logic)
			statsMap[r.Hour] = r.Avg
		}
	}

	// 3. Construct response sorted by hour
	var response []forms.HourlyStatsResponse
	for h := 9; h <= 21; h++ {
		val := statsMap[h] // 0 if missing

		// Currently only returning absolute count regardless of type
		// If future stats types need different processing or different response fields,
		// logic will diverge here or in different method.
		
		response = append(response, forms.HourlyStatsResponse{
			Hour:           h,
			AvgPersonCount: val,
		})
	}

	noData := len(dailyRows) == 0 && len(occupancyRows) == 0
	return response, noData, nil
}
