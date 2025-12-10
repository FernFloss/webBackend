package forms

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

// ErrInvalidCameraEvent indicates that incoming camera event failed validation.
var ErrInvalidCameraEvent = errors.New("invalid camera event")

// CameraEvent represents an incoming message from RabbitMQ about room occupancy.
type CameraEvent struct {
	IDCamera    string    `json:"id_camera" validate:"required,mac"`
	Timestamp   time.Time `json:"timestamp" validate:"required"`
	// PersonCount is a pointer to distinguish "field absent" (nil) from zero.
	PersonCount *int `json:"person_count" validate:"required,gte=0"`
}

// Validate ensures the event has all required fields and sane values via struct tags.
func (e *CameraEvent) Validate() error {
	if e == nil {
		return fmt.Errorf("%w: event is nil", ErrInvalidCameraEvent)
	}

	v := validator.New(validator.WithRequiredStructEnabled())
	if err := v.Struct(e); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidCameraEvent, err)
	}
	return nil
}
