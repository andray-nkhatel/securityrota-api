package handlers

import (
	"net/http"
	"time"

	"securityrota-api/database"
	"securityrota-api/models"

	"github.com/gin-gonic/gin"
)

// GetShiftsInput represents query params for getting shifts
type GetShiftsInput struct {
	Date      string `form:"date"` // YYYY-MM-DD
	OfficerID uint   `form:"officer_id"`
	WeekStart string `form:"week_start"` // YYYY-MM-DD (Sunday)
}

// GetShifts godoc
// @Summary Get shifts
// @Description Get shifts with optional filters (date, officer_id, week_start)
// @Tags shifts
// @Produce json
// @Param date query string false "Date (YYYY-MM-DD)"
// @Param officer_id query int false "Officer ID"
// @Param week_start query string false "Week start date (Sunday, YYYY-MM-DD)"
// @Success 200 {array} models.Shift
// @Router /shifts [get]
func GetShifts(c *gin.Context) {
	var input GetShiftsInput
	c.ShouldBindQuery(&input)

	query := database.DB.Preload("Officer")

	if input.Date != "" {
		date, _ := time.Parse("2006-01-02", input.Date)
		query = query.Where("date = ?", date)
	}

	if input.OfficerID > 0 {
		query = query.Where("officer_id = ?", input.OfficerID)
	}

	if input.WeekStart != "" {
		weekStart, _ := time.Parse("2006-01-02", input.WeekStart)
		weekEnd := weekStart.AddDate(0, 0, 7)
		query = query.Where("date >= ? AND date < ?", weekStart, weekEnd)
	}

	var shifts []models.Shift
	query.Order("date ASC").Find(&shifts)
	c.JSON(http.StatusOK, shifts)
}

// GenerateWeekRotaInput represents input for generating a week's rota
type GenerateWeekRotaInput struct {
	WeekStart string `json:"week_start" binding:"required"` // YYYY-MM-DD (must be Sunday)
}

// GenerateWeekRota godoc
// @Summary Generate rota for a week
// @Description Generate the complete shift rota for a given week starting on Sunday
// @Tags shifts
// @Accept json
// @Produce json
// @Param input body GenerateWeekRotaInput true "Week start date (must be Sunday)"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /shifts/generate [post]
func GenerateWeekRota(c *gin.Context) {
	var input GenerateWeekRotaInput
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

	// Check if rota already exists for this week
	var existingRotation models.WeekRotation
	if database.DB.Where("week_start = ?", weekStart).First(&existingRotation).Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rota already exists for this week"})
		return
	}

	// Get previous week's rotation to determine this week's teams
	prevWeekStart := weekStart.AddDate(0, 0, -7)
	var prevRotation models.WeekRotation
	dayShiftTeam := 1 // Default if no previous rotation
	if database.DB.Where("week_start = ?", prevWeekStart).First(&prevRotation).Error == nil {
		// Swap teams: previous day shift becomes night shift
		if prevRotation.DayShiftTeam == 1 {
			dayShiftTeam = 2
		} else {
			dayShiftTeam = 1
		}
	}

	nightShiftTeam := 1
	if dayShiftTeam == 1 {
		nightShiftTeam = 2
	}

	// Create week rotation record
	rotation := models.WeekRotation{
		WeekStart:    weekStart,
		DayShiftTeam: dayShiftTeam,
	}
	database.DB.Create(&rotation)

	// Get all officers
	var sergeant models.Officer
	var female1, female2 models.Officer
	var dayTeamOfficers, nightTeamOfficers []models.Officer

	database.DB.Where("role = ?", models.RoleSergeant).First(&sergeant)
	database.DB.Where("role = ?", models.RoleFemale).Find(&[]models.Officer{})

	// Get female officers (assuming we have exactly 2)
	var females []models.Officer
	database.DB.Where("role = ?", models.RoleFemale).Find(&females)
	if len(females) >= 2 {
		female1 = females[0] // Sunday off
		female2 = females[1] // Saturday off
	}

	database.DB.Where("role = ? AND team = ?", models.RoleRegular, dayShiftTeam).Find(&dayTeamOfficers)
	database.DB.Where("role = ? AND team = ?", models.RoleRegular, nightShiftTeam).Find(&nightTeamOfficers)

	var shifts []models.Shift

	// Generate shifts for each day of the week
	for dayOffset := 0; dayOffset < 7; dayOffset++ {
		currentDate := weekStart.AddDate(0, 0, dayOffset)
		weekday := currentDate.Weekday()

		// Sergeant: Day shift Sun-Fri, off Saturday
		if sergeant.ID > 0 {
			if weekday != time.Saturday {
				shifts = append(shifts, models.Shift{
					OfficerID: sergeant.ID,
					Date:      currentDate,
					ShiftType: models.ShiftDay,
					Status:    models.StatusOnDuty,
				})
			} else {
				shifts = append(shifts, models.Shift{
					OfficerID: sergeant.ID,
					Date:      currentDate,
					ShiftType: models.ShiftDay,
					Status:    models.StatusOffDuty,
				})
			}
		}

		// Female Officer 1: Day shift Mon-Sat, off Sunday
		if female1.ID > 0 {
			if weekday != time.Sunday {
				shifts = append(shifts, models.Shift{
					OfficerID: female1.ID,
					Date:      currentDate,
					ShiftType: models.ShiftDay,
					Status:    models.StatusOnDuty,
				})
			} else {
				shifts = append(shifts, models.Shift{
					OfficerID: female1.ID,
					Date:      currentDate,
					ShiftType: models.ShiftDay,
					Status:    models.StatusOffDuty,
				})
			}
		}

		// Female Officer 2: Day shift Sun-Fri, off Saturday
		if female2.ID > 0 {
			if weekday != time.Saturday {
				shifts = append(shifts, models.Shift{
					OfficerID: female2.ID,
					Date:      currentDate,
					ShiftType: models.ShiftDay,
					Status:    models.StatusOnDuty,
				})
			} else {
				shifts = append(shifts, models.Shift{
					OfficerID: female2.ID,
					Date:      currentDate,
					ShiftType: models.ShiftDay,
					Status:    models.StatusOffDuty,
				})
			}
		}

		// Handle Sunday special case
		if weekday == time.Sunday {
			// Sunday day shift: 2 officers from day team (who worked Sat day)
			for i, officer := range dayTeamOfficers {
				if i < 2 {
					shifts = append(shifts, models.Shift{
						OfficerID: officer.ID,
						Date:      currentDate,
						ShiftType: models.ShiftDay,
						Status:    models.StatusOnDuty,
					})
				} else {
					// Rest of day team is off on Sunday (they worked Sat night prev week)
					shifts = append(shifts, models.Shift{
						OfficerID: officer.ID,
						Date:      currentDate,
						ShiftType: models.ShiftDay,
						Status:    models.StatusOffDuty,
					})
				}
			}

			// Sunday night shift: all night team officers
			for _, officer := range nightTeamOfficers {
				shifts = append(shifts, models.Shift{
					OfficerID: officer.ID,
					Date:      currentDate,
					ShiftType: models.ShiftNight,
					Status:    models.StatusOnDuty,
				})
			}
		} else if weekday == time.Saturday {
			// Saturday: All day team on duty, night team on duty
			for _, officer := range dayTeamOfficers {
				shifts = append(shifts, models.Shift{
					OfficerID: officer.ID,
					Date:      currentDate,
					ShiftType: models.ShiftDay,
					Status:    models.StatusOnDuty,
				})
			}
			for _, officer := range nightTeamOfficers {
				shifts = append(shifts, models.Shift{
					OfficerID: officer.ID,
					Date:      currentDate,
					ShiftType: models.ShiftNight,
					Status:    models.StatusOnDuty,
				})
			}
		} else {
			// Monday-Friday
			// Day team all on duty
			for _, officer := range dayTeamOfficers {
				shifts = append(shifts, models.Shift{
					OfficerID: officer.ID,
					Date:      currentDate,
					ShiftType: models.ShiftDay,
					Status:    models.StatusOnDuty,
				})
			}

			// Night team: 2 officers off each day Mon-Thu (rotating)
			dayIndex := int(weekday) - 1 // Mon=0, Tue=1, Wed=2, Thu=3
			for i, officer := range nightTeamOfficers {
				status := models.StatusOnDuty
				// Simple rotation: 2 officers off based on day index
				if dayIndex >= 0 && dayIndex < 4 { // Mon-Thu
					offStart := (dayIndex * 2) % len(nightTeamOfficers)
					if len(nightTeamOfficers) > 0 {
						offEnd := (offStart + 2) % len(nightTeamOfficers)
						if offStart < offEnd {
							if i >= offStart && i < offEnd {
								status = models.StatusOffDuty
							}
						} else {
							if i >= offStart || i < offEnd {
								status = models.StatusOffDuty
							}
						}
					}
				}
				shifts = append(shifts, models.Shift{
					OfficerID: officer.ID,
					Date:      currentDate,
					ShiftType: models.ShiftNight,
					Status:    status,
				})
			}
		}
	}

	// Batch insert shifts
	if len(shifts) > 0 {
		database.DB.Create(&shifts)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":          "Rota generated successfully",
		"week_start":       input.WeekStart,
		"day_shift_team":   dayShiftTeam,
		"night_shift_team": nightShiftTeam,
		"shifts_created":   len(shifts),
	})
}

// GetWeekRotation godoc
// @Summary Get week rotation info
// @Description Get which team is on which shift for a given week
// @Tags shifts
// @Produce json
// @Param week_start query string true "Week start date (Sunday, YYYY-MM-DD)"
// @Success 200 {object} models.WeekRotation
// @Failure 404 {object} map[string]string
// @Router /shifts/rotation [get]
func GetWeekRotation(c *gin.Context) {
	weekStartStr := c.Query("week_start")
	if weekStartStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "week_start is required"})
		return
	}

	weekStart, err := time.Parse("2006-01-02", weekStartStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	var rotation models.WeekRotation
	if err := database.DB.Where("week_start = ?", weekStart).First(&rotation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rotation not found for this week"})
		return
	}

	c.JSON(http.StatusOK, rotation)
}
