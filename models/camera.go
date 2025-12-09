package models

import (
	"errors"
	"fmt"
	"web_backend_v2/db"
	"web_backend_v2/forms"

	"gorm.io/gorm"
)

type CameraModel struct{}

type CameraWithAssignment struct {
	ID           uint   `gorm:"column:id"`
	Mac          string `gorm:"column:mac"`
	AuditoriumID *uint  `gorm:"column:auditorium_id"`
}

// CreateCamera creates a new camera with the given MAC.
func (m *CameraModel) CreateCamera(mac string) (*forms.Camera, error) {
	camera := forms.Camera{
		Mac: mac,
	}
	result := db.GetDB().Create(&camera)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create camera: %w", result.Error)
	}
	return &camera, nil
}

// GetCameraWithAssignment returns camera and an optional auditorium assignment.
func (m *CameraModel) GetCameraWithAssignment(id uint) (*CameraWithAssignment, error) {
	var cam CameraWithAssignment
	tx := db.GetDB().
		Table("camera c").
		Select("c.id, c.mac, cia.auditorium_id").
		Joins("LEFT JOIN camerasinauditorium cia ON cia.camera_id = c.id").
		Where("c.id = ?", id).
		First(&cam)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("failed to fetch camera: %w", tx.Error)
	}
	return &cam, nil
}

// GetCamerasByAuditorium returns all cameras attached to the given auditorium.
func (m *CameraModel) GetCamerasByAuditorium(auditoriumID uint) ([]forms.Camera, error) {
	var cams []forms.Camera
	tx := db.GetDB().
		Table("camera c").
		Joins("JOIN camerasinauditorium cia ON cia.camera_id = c.id").
		Where("cia.auditorium_id = ?", auditoriumID).
		Order("c.id").
		Find(&cams)
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to fetch cameras for auditorium %d: %w", auditoriumID, tx.Error)
	}
	return cams, nil
}

// GetFreeCameras returns cameras not attached to any auditorium.
func (m *CameraModel) GetFreeCameras() ([]forms.Camera, error) {
	var cams []forms.Camera
	tx := db.GetDB().
		Table("camera c").
		Joins("LEFT JOIN camerasinauditorium cia ON cia.camera_id = c.id").
		Where("cia.camera_id IS NULL").
		Order("c.id").
		Find(&cams)
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to fetch free cameras: %w", tx.Error)
	}
	return cams, nil
}

// GetAttachedCameras returns cameras that are assigned to any auditorium with assignment info.
func (m *CameraModel) GetAttachedCameras() ([]CameraWithAssignment, error) {
	var cams []CameraWithAssignment
	tx := db.GetDB().
		Table("camera c").
		Select("c.id, c.mac, cia.auditorium_id").
		Joins("JOIN camerasinauditorium cia ON cia.camera_id = c.id").
		Order("c.id").
		Find(&cams)
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to fetch attached cameras: %w", tx.Error)
	}
	return cams, nil
}

// AttachCameraToAuditorium links camera to an auditorium, ensuring a camera is linked only once.
func (m *CameraModel) AttachCameraToAuditorium(cameraID, auditoriumID uint) error {
	dbConn := db.GetDB()

	return dbConn.Transaction(func(tx *gorm.DB) error {
		// Ensure camera exists
		var count int64
		if err := tx.Table("camera").Where("id = ?", cameraID).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check camera existence: %w", err)
		}
		if count == 0 {
			return fmt.Errorf("camera %d not found", cameraID)
		}

		// Ensure auditorium exists
		if err := tx.Table("auditorium").Where("id = ?", auditoriumID).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check auditorium existence: %w", err)
		}
		if count == 0 {
			return fmt.Errorf("auditorium %d not found", auditoriumID)
		}

		// Check if camera is already attached to another auditorium
		var existing forms.CamerasInAuditorium
		err := tx.Table("camerasinauditorium").
			Where("camera_id = ?", cameraID).
			First(&existing).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check camera assignment: %w", err)
		}
		if err == nil && existing.AuditoriumID != auditoriumID {
			return fmt.Errorf("camera %d already assigned to auditorium %d", cameraID, existing.AuditoriumID)
		}

		// Upsert assignment
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Table("camerasinauditorium").Create(&forms.CamerasInAuditorium{
				CameraID:     cameraID,
				AuditoriumID: auditoriumID,
			}).Error; err != nil {
				return fmt.Errorf("failed to assign camera: %w", err)
			}
		} else {
			if err := tx.Table("camerasinauditorium").
				Where("camera_id = ?", cameraID).
				Update("auditorium_id", auditoriumID).Error; err != nil {
				return fmt.Errorf("failed to update camera assignment: %w", err)
			}
		}
		return nil
	})
}

// DetachCameraFromAuditorium removes camera assignment if exists.
func (m *CameraModel) DetachCameraFromAuditorium(cameraID uint) error {
	tx := db.GetDB().Table("camerasinauditorium").Where("camera_id = ?", cameraID).Delete(&forms.CamerasInAuditorium{})
	if tx.Error != nil {
		return fmt.Errorf("failed to detach camera: %w", tx.Error)
	}
	return nil
}

// DeleteCamera removes camera; assignment is removed via FK cascade.
func (m *CameraModel) DeleteCamera(cameraID uint) error {
	tx := db.GetDB().Table("camera").Where("id = ?", cameraID).Delete(&forms.Camera{})
	if tx.Error != nil {
		return fmt.Errorf("failed to delete camera: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}


