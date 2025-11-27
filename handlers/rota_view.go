package handlers

import (
	"net/http"
	"time"

	"securityrota-api/database"
	"securityrota-api/models"

	"github.com/gin-gonic/gin"
)

// DayRota represents a single day's rota
type DayRota struct {
	Date       string        `json:"date"`
	DayOfWeek  string        `json:"day_of_week"`
	DayShift   []OfficerDuty `json:"day_shift"`
	NightShift []OfficerDuty `json:"night_shift"`
}

// OfficerDuty represents an officer's duty status
type OfficerDuty struct {
	Name   string `json:"name"`
	Role   string `json:"role"`
	Status string `json:"status"` // on_duty or off_duty
}

// WeekRotaResponse represents the complete weekly rota
type WeekRotaResponse struct {
	WeekStart      string    `json:"week_start"`
	WeekEnd        string    `json:"week_end"`
	DayShiftTeam   int       `json:"day_shift_team"`
	NightShiftTeam int       `json:"night_shift_team"`
	Days           []DayRota `json:"days"`
}

// GetWeekRota godoc
// @Summary Get complete weekly rota view
// @Description Get the full duty rota for a specific week with all officers and their shifts
// @Tags rota
// @Produce json
// @Param week_start query string true "Week start date (Sunday, YYYY-MM-DD)"
// @Success 200 {object} WeekRotaResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /rota/week [get]
func GetWeekRota(c *gin.Context) {
	weekStartStr := c.Query("week_start")
	if weekStartStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "week_start is required"})
		return
	}

	weekStart, err := time.Parse("2006-01-02", weekStartStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use YYYY-MM-DD"})
		return
	}

	if weekStart.Weekday() != time.Sunday {
		c.JSON(http.StatusBadRequest, gin.H{"error": "week_start must be a Sunday"})
		return
	}

	// Get week rotation info
	var rotation models.WeekRotation
	if err := database.DB.Where("week_start = ?", weekStart).First(&rotation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No rota found for this week. Generate it first using POST /shifts/generate"})
		return
	}

	nightShiftTeam := 1
	if rotation.DayShiftTeam == 1 {
		nightShiftTeam = 2
	}

	weekEnd := weekStart.AddDate(0, 0, 6)

	// Get all shifts for the week
	var shifts []models.Shift
	database.DB.Preload("Officer").
		Where("date >= ? AND date <= ?", weekStart, weekEnd).
		Order("date ASC, shift_type ASC").
		Find(&shifts)

	// Organize shifts by day
	dayRotas := make([]DayRota, 7)
	for i := 0; i < 7; i++ {
		currentDate := weekStart.AddDate(0, 0, i)
		dayRotas[i] = DayRota{
			Date:       currentDate.Format("2006-01-02"),
			DayOfWeek:  currentDate.Weekday().String(),
			DayShift:   []OfficerDuty{},
			NightShift: []OfficerDuty{},
		}
	}

	// Populate shifts
	for _, shift := range shifts {
		dayIndex := int(shift.Date.Sub(weekStart).Hours() / 24)
		if dayIndex < 0 || dayIndex > 6 {
			continue
		}

		duty := OfficerDuty{
			Name:   shift.Officer.Name,
			Role:   string(shift.Officer.Role),
			Status: string(shift.Status),
		}

		if shift.ShiftType == models.ShiftDay {
			dayRotas[dayIndex].DayShift = append(dayRotas[dayIndex].DayShift, duty)
		} else {
			dayRotas[dayIndex].NightShift = append(dayRotas[dayIndex].NightShift, duty)
		}
	}

	response := WeekRotaResponse{
		WeekStart:      weekStart.Format("2006-01-02"),
		WeekEnd:        weekEnd.Format("2006-01-02"),
		DayShiftTeam:   rotation.DayShiftTeam,
		NightShiftTeam: nightShiftTeam,
		Days:           dayRotas,
	}

	c.JSON(http.StatusOK, response)
}
