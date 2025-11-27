package handlers

import (
	"net/http"
	"time"

	"securityrota-api/database"
	"securityrota-api/models"

	"github.com/gin-gonic/gin"
)

// ImportCurrentStateInput represents the current manual schedule state
type ImportCurrentStateInput struct {
	WeekStart    string `json:"week_start" binding:"required"`     // Current week's Sunday (YYYY-MM-DD)
	DayShiftTeam int    `json:"day_shift_team" binding:"required"` // Which team (1 or 2) is currently on day shift
}

// ImportCurrentState godoc
// @Summary Import current manual schedule state
// @Description Set the current rotation state to sync with existing manual schedule
// @Tags admin
// @Accept json
// @Produce json
// @Param input body ImportCurrentStateInput true "Current schedule state"
// @Success 201 {object} models.WeekRotation
// @Failure 400 {object} map[string]string
// @Router /admin/import-state [post]
func ImportCurrentState(c *gin.Context) {
	var input ImportCurrentStateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	weekStart, err := time.Parse("2006-01-02", input.WeekStart)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use YYYY-MM-DD"})
		return
	}

	if weekStart.Weekday() != time.Sunday {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Week start must be a Sunday"})
		return
	}

	if input.DayShiftTeam != 1 && input.DayShiftTeam != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "day_shift_team must be 1 or 2"})
		return
	}

	// Check if rotation already exists
	var existing models.WeekRotation
	if database.DB.Where("week_start = ?", weekStart).First(&existing).Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rotation already exists for this week. Delete it first if you want to re-import."})
		return
	}

	// Create the rotation record to establish current state
	rotation := models.WeekRotation{
		WeekStart:    weekStart,
		DayShiftTeam: input.DayShiftTeam,
	}
	database.DB.Create(&rotation)

	nightShiftTeam := 1
	if input.DayShiftTeam == 1 {
		nightShiftTeam = 2
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":          "Current state imported successfully",
		"week_start":       input.WeekStart,
		"day_shift_team":   input.DayShiftTeam,
		"night_shift_team": nightShiftTeam,
		"next_step":        "Now you can generate future weeks using POST /shifts/generate",
	})
}

// BulkImportShiftInput represents a single shift to import
type BulkImportShiftInput struct {
	Name      string `json:"name" binding:"required"`       // Officer name
	Date      string `json:"date" binding:"required"`       // YYYY-MM-DD
	ShiftType string `json:"shift_type" binding:"required"` // day or night
	Status    string `json:"status" binding:"required"`     // on_duty or off_duty
}

// BulkImportShiftsInput represents bulk import request
type BulkImportShiftsInput struct {
	Shifts []BulkImportShiftInput `json:"shifts" binding:"required"`
}

// BulkImportShifts godoc
// @Summary Bulk import existing shifts
// @Description Import historical or current shifts from manual schedule
// @Tags admin
// @Accept json
// @Produce json
// @Param input body BulkImportShiftsInput true "Shifts to import"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /admin/import-shifts [post]
func BulkImportShifts(c *gin.Context) {
	var input BulkImportShiftsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var created, failed int
	var errors []string

	for _, s := range input.Shifts {
		// Find officer by name
		var officer models.Officer
		if err := database.DB.Where("name = ?", s.Name).First(&officer).Error; err != nil {
			failed++
			errors = append(errors, "Officer not found: "+s.Name)
			continue
		}

		date, err := time.Parse("2006-01-02", s.Date)
		if err != nil {
			failed++
			errors = append(errors, "Invalid date for "+s.Name+": "+s.Date)
			continue
		}

		shift := models.Shift{
			OfficerID: officer.ID,
			Date:      date,
			ShiftType: models.ShiftType(s.ShiftType),
			Status:    models.DutyStatus(s.Status),
		}

		if err := database.DB.Create(&shift).Error; err != nil {
			failed++
			errors = append(errors, "Failed to create shift for "+s.Name)
			continue
		}
		created++
	}

	c.JSON(http.StatusCreated, gin.H{
		"created": created,
		"failed":  failed,
		"errors":  errors,
	})
}
