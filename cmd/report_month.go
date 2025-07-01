package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/deadpyxel/workday/internal/journal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// reportMonthCmd represents the report command for generating a report for the current month.
var reportMonthCmd = &cobra.Command{
	Use:   "month",
	Short: "Generates a report for the current month",
	Long: `The month command generates a report for the current month.
It loads the existing journal entries from the file and fetches the entries for the current month.
If there are no entries for the current month, it returns an error.
Otherwise, it prints out the entries.`,
	RunE: reportMonth,
}

type reportMonthModel struct {
	entries       []journal.JournalEntry
	month         time.Time
	totalWorkTime time.Duration
	width         int
	height        int
	quitting      bool
}

func (m reportMonthModel) Init() tea.Cmd {
	return nil
}

func (m reportMonthModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m reportMonthModel) View() string {
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
	monthStr := m.month.Format("January 2006")
	content.WriteString(titleStyle.Render(fmt.Sprintf("📊 Monthly Report - %s", monthStr)))
	content.WriteString("\n\n")

	// Daily Entries Section
	content.WriteString(sectionStyle.Render("📅 Daily Entries"))
	content.WriteString("\n")

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

		content.WriteString(entryStyle.Render(entryContent.String()))
		content.WriteString("\n")
	}

	// Summary Section
	content.WriteString(summaryStyle.Render(fmt.Sprintf("📈 Summary: %d entries • Total work time: %v",
		len(m.entries),
		m.totalWorkTime)))
	content.WriteString("\n")

	// Help
	content.WriteString(helpStyle.Render("Press 'q' or 'esc' to quit"))

	return content.String()
}

func reportMonth(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	// Check if the month flag has been set
	monthFlag, _ := cmd.Flags().GetString("month")
	monthFilter := time.Now()
	if monthFlag != "" {
		monthFilter, err = time.Parse("2006-01", monthFlag)
		if err != nil {
			return errors.New("invalid month format, expected YYYY-MM")
		}
	}

	currMonth, err := journal.FetchEntriesByMonthDate(entries, monthFilter)
	if err != nil {
		return err
	}

	if len(currMonth) == 0 {
		return fmt.Errorf("no entries found for %s", monthFilter.Format("January 2006"))
	}

	lunchTime, err := time.ParseDuration(viper.GetString("lunchTime"))
	if err != nil {
		return err
	}

	var totalWorkTime time.Duration
	for _, entry := range currMonth {
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

	model := reportMonthModel{
		entries:       currMonth,
		month:         monthFilter,
		totalWorkTime: totalWorkTime,
	}

	p := tea.NewProgram(&model)
	_, err = p.Run()
	return err
}

func init() {
	reportMonthCmd.Flags().StringP("month", "m", "", "Specify the month and year in the format YYYY-MM")
	reportCmd.AddCommand(reportMonthCmd)
}
