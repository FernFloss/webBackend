package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"web_backend_v2/forms"
	"web_backend_v2/models"

	"github.com/gin-gonic/gin"
)

// GetAuditoriumsByBuilding handles GET /v1/buildings/:building_id/auditoriums
// Returns a list of auditoriums in a specific building
var AuditoriumModel = new(models.AuditoryModel)

type AuditoriumController struct{}

func (b *AuditoriumController) GetAuditoriumsByBuilding(c *gin.Context) {
	BuildingIDStr := c.Param("building_id")

	if BuildingIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "city_id parameter is required",
		})
		return
	}

	BuildingIDUint64, err := strconv.ParseUint(BuildingIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid city_id format. Must be a positive integer",
		})
		return
	}

	BuildingID := uint(BuildingIDUint64)
	if BuildingID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "city_id must be greater than zero",
		})
		return
	}

	auditoriums, err := AuditoriumModel.GetAuditoriumsByBuilding(BuildingID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintln(err),
		})
	}
	// Convert to response format
	response := make([]forms.AuditoriumResponse, len(auditoriums))
	for i := range auditoriums {
		response[i] = auditoriums[i].ToAuditoriumResponse()
	}

	c.JSON(http.StatusOK, response)
}
