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

var reportDate string

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Reports the workday entry for the current day",
	Long: `The report command is used to print out the workday entry for the current day.

It loads the existing jorunal entries from the file and fetches the entry for the current day.
If there is no entry for the current day, it returns an error.
Otherwise, it prints out the entry.`,
	RunE: reportWorkDay,
}

type reportModel struct {
	entry    *journal.JournalEntry
	date     time.Time
	width    int
	height   int
	quitting bool
}

func (m reportModel) Init() tea.Cmd {
	return nil
}

func (m reportModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m reportModel) View() string {
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

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		Width(12)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	breakStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		PaddingLeft(2)

	noteStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("159")).
		PaddingLeft(2).
		MarginBottom(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	var content strings.Builder

	// Title
	dateStr := m.date.Format("Monday, January 2, 2006")
	content.WriteString(titleStyle.Render(fmt.Sprintf("Workday Report - %s", dateStr)))
	content.WriteString("\n\n")

	// Work Hours Section
	content.WriteString(sectionStyle.Render("⏰ Work Hours"))
	content.WriteString("\n")

	startTime := m.entry.StartTime.Format("15:04")
	content.WriteString(labelStyle.Render("Start:") + " " + valueStyle.Render(startTime))
	content.WriteString("\n")

	endTime := "Ongoing"
	if !m.entry.EndTime.IsZero() {
		endTime = m.entry.EndTime.Format("15:04")
	}
	content.WriteString(labelStyle.Render("End:") + " " + valueStyle.Render(endTime))
	content.WriteString("\n")

	// Calculate work duration
	var duration time.Duration
	if !m.entry.EndTime.IsZero() {
		duration = m.entry.EndTime.Sub(m.entry.StartTime)

		// Subtract break time
		for _, br := range m.entry.Breaks {
			if !br.EndTime.IsZero() {
				duration -= br.EndTime.Sub(br.StartTime)
			}
		}

		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		content.WriteString(labelStyle.Render("Duration:") + " " + valueStyle.Render(fmt.Sprintf("%dh %dm", hours, minutes)))
	} else {
		content.WriteString(labelStyle.Render("Duration:") + " " + valueStyle.Render("In progress"))
	}
	content.WriteString("\n")

	// Breaks Section
	if len(m.entry.Breaks) > 0 {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("☕ Breaks"))
		content.WriteString("\n")

		for i, br := range m.entry.Breaks {
			startTime := br.StartTime.Format("15:04")
			endTime := "Ongoing"
			if !br.EndTime.IsZero() {
				endTime = br.EndTime.Format("15:04")
			}

			breakText := fmt.Sprintf("%d. %s - %s", i+1, startTime, endTime)
			if br.Reason != "" {
				breakText += fmt.Sprintf(" (%s)", br.Reason)
			}

			content.WriteString(breakStyle.Render(breakText))
			content.WriteString("\n")
		}
	}

	// Notes Section
	if len(m.entry.Notes) > 0 {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("📝 Notes"))
		content.WriteString("\n")

		for i, note := range m.entry.Notes {
			noteText := fmt.Sprintf("%d. %s", i+1, note.Contents)
			if len(note.Tags) > 0 {
				noteText += fmt.Sprintf(" [%s]", strings.Join(note.Tags, ", "))
			}
			content.WriteString(noteStyle.Render(noteText))
			content.WriteString("\n")
		}
	}

	// Help
	content.WriteString("\n")
	content.WriteString(helpStyle.Render("Press 'q' or 'esc' to quit"))

	return content.String()
}

func reportWorkDay(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	journalEntries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	var tgtEntry *journal.JournalEntry
	tgtDay := time.Now()

	if reportDate != "" {
		tgtDay, err = time.Parse("2006-01-02", reportDate)
		if err != nil {
			return fmt.Errorf("invalid date format, Use YYYY-MM-DD: %v", err)
		}
	}

	tgtDayID := tgtDay.Format("20060102")
	tgtEntry, _ = journal.FetchEntryByID(tgtDayID, journalEntries)
	if tgtEntry == nil {
		return fmt.Errorf("Could not find any entry for the date: %s", reportDate)
	}

	model := reportModel{
		entry: tgtEntry,
		date:  tgtDay,
	}

	p := tea.NewProgram(&model)
	_, err = p.Run()
	return err
}

func init() {
	rootCmd.AddCommand(reportCmd)

	// Specify date for report
	reportCmd.Flags().StringVarP(&reportDate, "date", "d", "", "Specify the date in YYYY-MM-DD format")
}
