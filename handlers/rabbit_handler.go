package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"web_backend_v2/forms"
	"web_backend_v2/models"
)

var occupancyModel = new(models.OccupancyModel)

// ProcessCameraEvent parses and stores occupancy data from RabbitMQ message.
func ProcessCameraEvent(messageBody []byte) error {
	var event forms.CameraEvent
	if err := json.Unmarshal(messageBody, &event); err != nil {
		return fmt.Errorf("failed to parse camera event (raw len=%d): %w", len(messageBody), err)
	}

	if err := event.Validate(); err != nil {
		return fmt.Errorf("camera %s validation failed: %w", event.IDCamera, err)
	}

	if err := occupancyModel.SaveEvent(&event); err != nil {
		return fmt.Errorf("camera %s save failed: %w", event.IDCamera, err)
	}

	log.Printf("Stored occupancy from camera %s: %d persons at %s", event.IDCamera, *event.PersonCount, event.Timestamp.UTC().Format("2006-01-02T15:04:05Z07:00"))
	return nil
}
