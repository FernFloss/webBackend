package models

import (
	"fmt"
	"time"
	"web_backend_v2/db"

	"gorm.io/gorm"
)

// AggregateDailyOccupancy aggregates Occupancy records for a given day into DailyLoad
// and then deletes the aggregated Occupancy rows. Intended to be run by a daily cron.
func AggregateDailyOccupancy(targetDay time.Time) error {
	start := time.Date(targetDay.Year(), targetDay.Month(), targetDay.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)

	return db.GetDB().Transaction(func(tx *gorm.DB) error {
		// Remove previous aggregates for the same day to keep the job idempotent.
		if err := tx.Exec(`DELETE FROM dailyload WHERE day = $1`, start).Error; err != nil {
			return fmt.Errorf("delete existing dailyload for day: %w", err)
		}

		// Insert aggregated averages per auditorium/hour.
		if err := tx.Exec(`
			INSERT INTO dailyload (auditorium_id, day, hour, avg_person_count)
			SELECT
				o.auditorium_id,
				DATE_TRUNC('day', o.timestamp)::date AS day,
				EXTRACT(hour FROM o.timestamp)::int AS hour,
				AVG(o.person_count)::float8 AS avg_person_count
			FROM occupancy o
			WHERE o.timestamp >= $1 AND o.timestamp < $2
			GROUP BY o.auditorium_id, day, hour
		`, start, end).Error; err != nil {
			return fmt.Errorf("insert dailyload aggregates: %w", err)
		}

		// Delete the source occupancy rows that were aggregated.
		if err := tx.Exec(`
			DELETE FROM occupancy
			WHERE timestamp >= $1 AND timestamp < $2
		`, start, end).Error; err != nil {
			return fmt.Errorf("cleanup occupancy: %w", err)
		}

		return nil
	})
}

