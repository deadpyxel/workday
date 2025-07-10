package cmd

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/deadpyxel/workday/internal/journal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// reportCmd represents the report command
var reportWeekCmd = &cobra.Command{
	Use:   "week",
	Short: "Generates a report for the current week",
	Long: `The week command generates a report for the current week.
It loads the existing journal entries from the file and fetches the entries for the current week.
If there are no entries for the current week, it returns an error.
Otherwise, it prints out the entries.`,
	RunE: reportWeek,
}

type reportWeekModel struct {
	entries       []journal.JournalEntry
	week          time.Time
	totalWorkTime time.Duration
	width         int
	height        int
	quitting      bool
}

func (m reportWeekModel) Init() tea.Cmd {
	return nil
}

func (m reportWeekModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m reportWeekModel) View() string {
	if m.quitting {
		return ""
	}

	// Define styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color("86")).
		MarginBottom(1).
		PaddingBottom(1)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Align(lipgloss.Center)

	cellStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Align(lipgloss.Center)

	summaryStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("120")).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(lipgloss.Color("120")).
		PaddingTop(1).
		MarginTop(2)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	var content strings.Builder

	// Title
	weekStart := m.week.AddDate(0, 0, -int(m.week.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 6)
	weekStr := fmt.Sprintf("%s - %s", 
		weekStart.Format("Jan 2"), 
		weekEnd.Format("Jan 2, 2006"))
	content.WriteString(titleStyle.Render(fmt.Sprintf("📅 Weekly Report - %s", weekStr)))
	content.WriteString("\n\n")

	// Create table data
	headers := []string{"Date", "Start", "End", "Duration", "Breaks"}
	rows := [][]string{}

	// Generate all days of the week
	for i := 0; i < 7; i++ {
		currentDay := weekStart.AddDate(0, 0, i)
		
		// Find entry for this day
		var entry *journal.JournalEntry
		for _, e := range m.entries {
			if e.StartTime.Format("2006-01-02") == currentDay.Format("2006-01-02") {
				entry = &e
				break
			}
		}

		if entry == nil {
			// No entry for this day
			rows = append(rows, []string{
				currentDay.Format("Mon, Jan 2"),
				"--",
				"--",
				"--",
				"--",
			})
		} else {
			// Format entry data
			date := entry.StartTime.Format("Mon, Jan 2")
			startTime := entry.StartTime.Format("15:04")
			
			endTime := "Ongoing"
			duration := "In progress"
			if !entry.EndTime.IsZero() {
				endTime = entry.EndTime.Format("15:04")
				
				// Calculate work duration
				workDuration := entry.EndTime.Sub(entry.StartTime)
				
				// Subtract break time
				var totalBreakTime time.Duration
				for _, br := range entry.Breaks {
					if !br.EndTime.IsZero() {
						totalBreakTime += br.EndTime.Sub(br.StartTime)
					}
				}
				
				workDuration -= totalBreakTime
				hours := int(workDuration.Hours())
				minutes := int(workDuration.Minutes()) % 60
				duration = fmt.Sprintf("%dh %dm", hours, minutes)
			}
			
			// Format breaks
			breakInfo := "--"
			if len(entry.Breaks) > 0 {
				var breakTimes []string
				for _, br := range entry.Breaks {
					if !br.EndTime.IsZero() {
						breakTime := fmt.Sprintf("%s-%s", br.StartTime.Format("15:04"), br.EndTime.Format("15:04"))
						breakTimes = append(breakTimes, breakTime)
					} else {
						breakTime := fmt.Sprintf("%s-ongoing", br.StartTime.Format("15:04"))
						breakTimes = append(breakTimes, breakTime)
					}
				}
				if len(breakTimes) > 0 {
					breakInfo = strings.Join(breakTimes, ", ")
				}
			}
			
			rows = append(rows, []string{date, startTime, endTime, duration, breakInfo})
		}
	}

	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Add padding to column widths
	for i := range colWidths {
		colWidths[i] += 2
	}

	// Create table
	var table strings.Builder
	
	// Top border
	table.WriteString("┌")
	for i, width := range colWidths {
		table.WriteString(strings.Repeat("─", width))
		if i < len(colWidths)-1 {
			table.WriteString("┬")
		}
	}
	table.WriteString("┐\n")

	// Header row
	table.WriteString("│")
	for i, header := range headers {
		cell := headerStyle.Width(colWidths[i]).Render(header)
		table.WriteString(cell)
		table.WriteString("│")
	}
	table.WriteString("\n")

	// Header separator
	table.WriteString("├")
	for i, width := range colWidths {
		table.WriteString(strings.Repeat("─", width))
		if i < len(colWidths)-1 {
			table.WriteString("┼")
		}
	}
	table.WriteString("┤\n")

	// Data rows
	for _, row := range rows {
		table.WriteString("│")
		for i, cell := range row {
			styledCell := cellStyle.Width(colWidths[i]).Render(cell)
			table.WriteString(styledCell)
			table.WriteString("│")
		}
		table.WriteString("\n")
	}

	// Bottom border
	table.WriteString("└")
	for i, width := range colWidths {
		table.WriteString(strings.Repeat("─", width))
		if i < len(colWidths)-1 {
			table.WriteString("┴")
		}
	}
	table.WriteString("┘\n")

	content.WriteString(table.String())

	// Summary Section
	workDays := 0
	for _, entry := range m.entries {
		if !entry.EndTime.IsZero() {
			workDays++
		}
	}
	
	content.WriteString(summaryStyle.Render(fmt.Sprintf("📊 Total work time: %v across %d days",
		m.totalWorkTime, workDays)))
	content.WriteString("\n")

	// Help
	content.WriteString(helpStyle.Render("Press 'q' or 'esc' to quit"))

	return content.String()
}

// reportWeek reports the workday entries for the current week.
// It first loads the existing journal entries from the file.
// If there are no entries for the current week, it returns an error.
// Otherwise, it displays the entries using Bubble Tea.
func reportWeek(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	journalEntries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}
	now := time.Now()
	currentWeek, err := journal.FetchEntriesByWeekDate(journalEntries, now)
	if err != nil {
		return err
	}

	lunchTime, err := time.ParseDuration(viper.GetString("lunchTime"))
	if err != nil {
		return err
	}

	var totalWorkTime time.Duration
	for _, entry := range currentWeek {
		if !entry.EndTime.IsZero() {
			dayWorkTime := entry.EndTime.Sub(entry.StartTime)
			var totalBreakTime time.Duration

			// Calculate total breaktime
			for _, br := range entry.Breaks {
				if !br.EndTime.IsZero() {
					totalBreakTime += br.EndTime.Sub(br.StartTime)
				}
			}

			// If no breaks were recorded, subtract default lunchTime
			if len(entry.Breaks) == 0 {
				totalBreakTime = lunchTime
			}

			dayWorkTime -= totalBreakTime
			totalWorkTime += dayWorkTime
		}
	}

	model := reportWeekModel{
		entries:       currentWeek,
		week:          now,
		totalWorkTime: totalWorkTime,
	}

	p := tea.NewProgram(&model)
	_, err = p.Run()
	return err
}

func init() {
	reportCmd.AddCommand(reportWeekCmd)
}
