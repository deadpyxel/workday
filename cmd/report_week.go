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

	// Define styles consistent with other report commands
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color("86")).
		MarginBottom(1).
		PaddingBottom(1)

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginTop(1).
		MarginBottom(1)

	entryStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		Width(12)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	noteStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("159")).
		PaddingLeft(2).
		MarginBottom(1)

	breakStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		PaddingLeft(2)

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

	// Daily Entries Section
	content.WriteString(sectionStyle.Render("📋 Daily Entries"))
	content.WriteString("\n")

	// Format each entry with proper structure
	for _, entry := range m.entries {
		var entryContent strings.Builder

		// Date header
		date := entry.StartTime.Format("Mon, Jan 2")
		entryContent.WriteString(labelStyle.Render("Date:") + " " + valueStyle.Render(date))
		entryContent.WriteString("\n")

		// Work hours
		startTime := entry.StartTime.Format("15:04")
		entryContent.WriteString(labelStyle.Render("Start:") + " " + valueStyle.Render(startTime))
		entryContent.WriteString("\n")

		endTime := "Ongoing"
		if !entry.EndTime.IsZero() {
			endTime = entry.EndTime.Format("15:04")
		}
		entryContent.WriteString(labelStyle.Render("End:") + " " + valueStyle.Render(endTime))
		entryContent.WriteString("\n")

		// Duration calculation
		if !entry.EndTime.IsZero() {
			duration := entry.EndTime.Sub(entry.StartTime)

			// Subtract break time
			for _, br := range entry.Breaks {
				if !br.EndTime.IsZero() {
					duration -= br.EndTime.Sub(br.StartTime)
				}
			}

			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60
			entryContent.WriteString(labelStyle.Render("Duration:") + " " + valueStyle.Render(fmt.Sprintf("%dh %dm", hours, minutes)))
		} else {
			entryContent.WriteString(labelStyle.Render("Duration:") + " " + valueStyle.Render("In progress"))
		}
		entryContent.WriteString("\n")

		// Breaks (if any)
		if len(entry.Breaks) > 0 {
			entryContent.WriteString("\n")
			entryContent.WriteString(labelStyle.Render("Breaks:"))
			entryContent.WriteString("\n")
			for i, br := range entry.Breaks {
				startTime := br.StartTime.Format("15:04")
				endTime := "Ongoing"
				if !br.EndTime.IsZero() {
					endTime = br.EndTime.Format("15:04")
				}

				breakText := fmt.Sprintf("%d. %s - %s", i+1, startTime, endTime)
				if br.Reason != "" {
					breakText += fmt.Sprintf(" (%s)", br.Reason)
				}

				entryContent.WriteString(breakStyle.Render(breakText))
				entryContent.WriteString("\n")
			}
		}

		// Notes (if any)
		if len(entry.Notes) > 0 {
			entryContent.WriteString("\n")
			entryContent.WriteString(labelStyle.Render("Notes:"))
			entryContent.WriteString("\n")
			for i, note := range entry.Notes {
				noteText := fmt.Sprintf("%d. %s", i+1, note.Contents)
				if len(note.Tags) > 0 {
					noteText += fmt.Sprintf(" [%s]", strings.Join(note.Tags, ", "))
				}
				entryContent.WriteString(noteStyle.Render(noteText))
				entryContent.WriteString("\n")
			}
		}

		content.WriteString(entryStyle.Render(entryContent.String()))
		content.WriteString("\n")
	}

	// Summary Section
	content.WriteString(summaryStyle.Render(fmt.Sprintf("📊 Summary: %d entries • Total work time: %v",
		len(m.entries),
		m.totalWorkTime)))
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
