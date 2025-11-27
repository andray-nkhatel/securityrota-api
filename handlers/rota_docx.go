package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"securityrota-api/database"
	"securityrota-api/models"

	"github.com/gin-gonic/gin"
	"github.com/unidoc/unioffice/document"
)

// GetWeekRotaDOCX godoc
// @Summary Download weekly rota as DOCX
// @Description Generate and download a DOCX of the weekly duty rota
// @Tags rota
// @Produce application/vnd.openxmlformats-officedocument.wordprocessingml.document
// @Param week_start query string true "Week start date (Sunday, YYYY-MM-DD)"
// @Success 200 {file} file
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /rota/week/docx [get]
func GetWeekRotaDOCX(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{"error": "No rota found for this week"})
		return
	}

	weekEnd := weekStart.AddDate(0, 0, 6)

	// Get all shifts for the week
	var shifts []models.Shift
	database.DB.Preload("Officer").
		Where("date >= ? AND date <= ?", weekStart, weekEnd).
		Order("date ASC, shift_type ASC, officer_id ASC").
		Find(&shifts)

	// Organize shifts by day
	type dayData struct {
		date       time.Time
		dayShift   []string
		nightShift []string
		leave      []string
	}
	days := make([]dayData, 7)
	for i := 0; i < 7; i++ {
		days[i] = dayData{date: weekStart.AddDate(0, 0, i)}
	}

	for _, shift := range shifts {
		dayIndex := int(shift.Date.Sub(weekStart).Hours() / 24)
		if dayIndex < 0 || dayIndex > 6 {
			continue
		}

		name := shift.Officer.Name
		name = strings.TrimPrefix(name, "Officer ")
		name = strings.TrimPrefix(name, "Sgt. ")
		name = strings.ToUpper(name)

		if shift.Status == models.StatusOffDuty {
			days[dayIndex].leave = append(days[dayIndex].leave, name)
		} else if shift.ShiftType == models.ShiftDay {
			days[dayIndex].dayShift = append(days[dayIndex].dayShift, name)
		} else {
			days[dayIndex].nightShift = append(days[dayIndex].nightShift, name)
		}
	}

	// Create a new document
	doc := document.New()

	// Title
	para := doc.AddParagraph()
	run := para.AddRun()
	run.AddText("Security Officer Duty Rota")
	run.Properties().SetBold(true)
	run.Properties().SetSize(28)

	para = doc.AddParagraph()
	run = para.AddRun()
	run.AddText(fmt.Sprintf("Week: %s to %s", weekStart.Format("Mon 02 Jan 2006"), weekEnd.Format("Mon 02 Jan 2006")))
	run.Properties().SetSize(22)

	para = doc.AddParagraph()
	nightShiftTeam := 1
	if rotation.DayShiftTeam == 1 {
		nightShiftTeam = 2
	}
	run = para.AddRun()
	run.AddText(fmt.Sprintf("Day Shift: Team %d | Night Shift: Team %d", rotation.DayShiftTeam, nightShiftTeam))
	run.Properties().SetSize(18)

	doc.AddParagraph()

	// Create table
	table := doc.AddTable()

	// Header row
	row := table.AddRow()
	cell := row.AddCell()
	para = cell.AddParagraph()
	run = para.AddRun()
	run.AddText("SHIFT TYPE")
	run.Properties().SetBold(true)

	dayNames := []string{"SUNDAY", "MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY"}
	for i, day := range dayNames {
		cell = row.AddCell()
		para = cell.AddParagraph()
		run = para.AddRun()
		run.AddText(day)
		run.Properties().SetBold(true)
		para.AddRun().AddBreak()
		para.AddRun().AddText(days[i].date.Format("02/01/06"))
	}

	// DAY SHIFT row
	row = table.AddRow()
	cell = row.AddCell()
	para = cell.AddParagraph()
	run = para.AddRun()
	run.AddText("DAY SHIFT")
	run.Properties().SetBold(true)

	for _, d := range days {
		cell = row.AddCell()
		for i, name := range d.dayShift {
			if i > 0 {
				cell.AddParagraph()
			}
			para = cell.AddParagraph()
			para.AddRun().AddText(name)
		}
	}

	// NIGHT SHIFT row
	row = table.AddRow()
	cell = row.AddCell()
	para = cell.AddParagraph()
	run = para.AddRun()
	run.AddText("NIGHT SHIFT")
	run.Properties().SetBold(true)

	for _, d := range days {
		cell = row.AddCell()
		for i, name := range d.nightShift {
			if i > 0 {
				cell.AddParagraph()
			}
			para = cell.AddParagraph()
			para.AddRun().AddText(name)
		}
	}

	// DAY-OFF row
	row = table.AddRow()
	cell = row.AddCell()
	para = cell.AddParagraph()
	run = para.AddRun()
	run.AddText("DAY-OFF")
	run.Properties().SetBold(true)

	for _, d := range days {
		cell = row.AddCell()
		for i, name := range d.leave {
			if i > 0 {
				cell.AddParagraph()
			}
			para = cell.AddParagraph()
			para.AddRun().AddText(name)
		}
	}

	// Generate document bytes
	docBytes, err := doc.WriteToBytes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate DOCX"})
		return
	}

	// Output
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	filename := fmt.Sprintf("rota_%s.docx", weekStartStr)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docBytes)
}
