package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"web_backend_v2/forms"
	"web_backend_v2/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetAuditoriumsByBuilding handles GET /v1/buildings/:building_id/auditoriums
// Returns a list of auditoriums in a specific building
var AuditoriumModel = new(models.AuditoryModel)

type AuditoriumController struct{}

const maxFreshMinutes = 5

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
		return
	}

	if len(auditoriums) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"warning":     "no auditoriums found for this building",
			"auditoriums": []forms.AuditoriumResponse{},
		})
		return
	}

	response := make([]forms.AuditoriumResponse, len(auditoriums))
	for i := range auditoriums {
		response[i] = auditoriums[i].ToAuditoriumResponse()
	}

	c.JSON(http.StatusOK, response)
}

// GetOccupancyByBuilding handles GET /v1/cities/:city_id/buildings/:building_id/auditories/occupancy
func (b *AuditoriumController) GetOccupancyByBuilding(c *gin.Context) {
	buildingID, err := parseUintParam(c, "building_id")
	if err != nil {
		return
	}

	var q forms.OccupancyQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "timestamp is required in RFC3339"})
		return
	}

	occupancies, err := AuditoriumModel.GetLatestOccupancyByBuilding(buildingID, q.Timestamp, maxFreshMinutes)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"warning":     "no occupancy data found for this building",
				"occupancies": []forms.AuditoriumOccupancyResponse{},
			})
			return
		}
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, forms.BuildingOccupancyResponse(occupancies))
}

// GetOccupancyByAuditorium handles GET /v1/cities/:city_id/buildings/:building_id/auditories/:auditorium_id/occupancy
func (b *AuditoriumController) GetOccupancyByAuditorium(c *gin.Context) {
	auditoriumID, err := parseUintParam(c, "auditorium_id")
	if err != nil {
		return
	}

	var q forms.OccupancyQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "timestamp is required in RFC3339"})
		return
	}

	occupancy, err := AuditoriumModel.GetLatestOccupancyForAuditorium(auditoriumID, q.Timestamp, maxFreshMinutes)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "no occupancy data found"})
			return
		}
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, occupancy)
}
