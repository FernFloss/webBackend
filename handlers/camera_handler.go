package handlers

import (
	"net/http"
	"strconv"
	"web_backend_v2/forms"
	"web_backend_v2/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var CameraModel = new(models.CameraModel)

// CreateCamera handles POST /v1/cameras
func (h *CameraController) CreateCamera(c *gin.Context) {
	var req forms.CreateCameraRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	camera, err := CameraModel.CreateCamera(req.Mac)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp := forms.CameraResponse{
		ID:  camera.ID,
		Mac: camera.Mac,
	}
	c.JSON(http.StatusCreated, resp)
}

// GetCamera handles GET /v1/cameras/:camera_id
func (h *CameraController) GetCamera(c *gin.Context) {
	cameraID, err := parseUintParam(c, "camera_id")
	if err != nil {
		return
	}

	camera, err := CameraModel.GetCameraWithAssignment(cameraID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "camera not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	resp := forms.CameraResponse{
		ID:           camera.ID,
		Mac:          camera.Mac,
		AuditoriumID: camera.AuditoriumID,
	}
	c.JSON(http.StatusOK, resp)
}

// GetFreeCameras handles GET /v1/cameras
func (h *CameraController) GetFreeCameras(c *gin.Context) {
	cameras, err := CameraModel.GetFreeCameras()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]forms.CameraResponse, len(cameras))
	for i := range cameras {
		resp[i] = forms.CameraResponse{
			ID:  cameras[i].ID,
			Mac: cameras[i].Mac,
		}
	}
	c.JSON(http.StatusOK, resp)
}

// GetCamerasByAuditorium handles GET /v1/cities/:city_id/buildings/:building_id/auditories/:auditorium_id/cameras
func (h *CameraController) GetCamerasByAuditorium(c *gin.Context) {
	auditoriumID, err := parseUintParam(c, "auditorium_id")
	if err != nil {
		return
	}

	cameras, err := CameraModel.GetCamerasByAuditorium(auditoriumID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]forms.CameraResponse, len(cameras))
	for i := range cameras {
		resp[i] = forms.CameraResponse{
			ID:           cameras[i].ID,
			Mac:          cameras[i].Mac,
			AuditoriumID: &auditoriumID,
		}
	}
	c.JSON(http.StatusOK, resp)
}

// AttachCamera handles POST /v1/cities/:city_id/buildings/:building_id/auditories/:auditorium_id/cameras
func (h *CameraController) AttachCamera(c *gin.Context) {
	auditoriumID, err := parseUintParam(c, "auditorium_id")
	if err != nil {
		return
	}

	var req forms.AttachCameraRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := CameraModel.AttachCameraToAuditorium(req.CameraID, auditoriumID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

type CameraController struct{}

// GetAttachedCameras handles GET /v1/cameras/attached
func (h *CameraController) GetAttachedCameras(c *gin.Context) {
	cameras, err := CameraModel.GetAttachedCameras()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]forms.CameraResponse, len(cameras))
	for i := range cameras {
		resp[i] = forms.CameraResponse{
			ID:           cameras[i].ID,
			Mac:          cameras[i].Mac,
			AuditoriumID: cameras[i].AuditoriumID,
		}
	}
	c.JSON(http.StatusOK, resp)
}

// DetachCamera handles DELETE /v1/cameras/:camera_id/attachment
func (h *CameraController) DetachCamera(c *gin.Context) {
	cameraID, err := parseUintParam(c, "camera_id")
	if err != nil {
		return
	}

	// Check camera existence
	_, err = CameraModel.GetCameraWithAssignment(cameraID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "camera not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if err := CameraModel.DetachCameraFromAuditorium(cameraID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteCamera handles DELETE /v1/cameras/:camera_id
func (h *CameraController) DeleteCamera(c *gin.Context) {
	cameraID, err := parseUintParam(c, "camera_id")
	if err != nil {
		return
	}

	confirm := c.Query("confirm") == "true"

	cam, err := CameraModel.GetCameraWithAssignment(cameraID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "camera not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if cam.AuditoriumID != nil && !confirm {
		c.JSON(http.StatusConflict, gin.H{
			"error":            "camera is attached to an auditorium",
			"auditorium_id":    cam.AuditoriumID,
			"require_confirm":  true,
			"confirm_hint":     "repeat request with ?confirm=true to delete and detach",
		})
		return
	}

	if err := CameraModel.DeleteCamera(cameraID); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "camera not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func parseUintParam(c *gin.Context, name string) (uint, error) {
	valStr := c.Param(name)
	if valStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": name + " parameter is required"})
		return 0, strconv.ErrSyntax
	}
	valUint64, err := strconv.ParseUint(valStr, 10, 32)
	if err != nil || valUint64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": name + " must be a positive integer"})
		return 0, err
	}
	return uint(valUint64), nil
}


