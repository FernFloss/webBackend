package models

import (
	"errors"
	"fmt"
	"web_backend_v2/db"
	"web_backend_v2/forms"

	"gorm.io/gorm"
)

var (
	// ErrCameraNotFound is returned when the camera MAC is unknown.
	ErrCameraNotFound = errors.New("camera not found")
	// ErrCameraNotAttached is returned when the camera has no auditorium assignment.
	ErrCameraNotAttached = errors.New("camera is not attached to an auditorium")
)

// OccupancyModel encapsulates occupancy-related operations.
type OccupancyModel struct{}

// SaveEvent stores occupancy info from a camera event.
func (o *OccupancyModel) SaveEvent(event *forms.CameraEvent) error {
	if event == nil {
		return fmt.Errorf("camera event is nil")
	}

	return db.GetDB().Transaction(func(tx *gorm.DB) error {
		var camera forms.Camera
		if err := tx.Table("camera").
			Where("mac = ?", event.IDCamera).
			First(&camera).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCameraNotFound
			}
			return fmt.Errorf("failed to load camera by mac %s: %w", event.IDCamera, err)
		}

		var assignment forms.CamerasInAuditorium
		if err := tx.Table("camerasinauditorium").
			Where("camera_id = ?", camera.ID).
			First(&assignment).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCameraNotAttached
			}
			return fmt.Errorf("failed to load camera assignment: %w", err)
		}

		record := forms.Occupancy{
			AuditoriumID: assignment.AuditoriumID,
			PersonCount:  event.PersonCount,
			Timestamp:    event.Timestamp.UTC(),
		}

		if err := tx.Table("occupancy").Create(&record).Error; err != nil {
			return fmt.Errorf("failed to create occupancy record: %w", err)
		}

		return nil
	})
}

