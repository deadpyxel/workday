package cmd

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/deadpyxel/workday/internal/journal"
	"github.com/deadpyxel/workday/internal/styles"
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


	var content strings.Builder

	// Title
	dateStr := m.date.Format("Monday, January 2, 2006")
	content.WriteString(styles.TitleStyle.Render(fmt.Sprintf("Workday Report - %s", dateStr)))
	content.WriteString("\n\n")

	// Work Hours Section
	content.WriteString(styles.SectionStyle.Render("⏰ Work Hours"))
	content.WriteString("\n")

	startTime := m.entry.StartTime.Format("15:04")
	content.WriteString(styles.LabelStyle.Render("Start:") + " " + styles.ValueStyle.Render(startTime))
	content.WriteString("\n")

	endTime := "Ongoing"
	if !m.entry.EndTime.IsZero() {
		endTime = m.entry.EndTime.Format("15:04")
	}
	content.WriteString(styles.LabelStyle.Render("End:") + " " + styles.ValueStyle.Render(endTime))
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
		content.WriteString(styles.LabelStyle.Render("Duration:") + " " + styles.ValueStyle.Render(fmt.Sprintf("%dh %dm", hours, minutes)))
	} else {
		content.WriteString(styles.LabelStyle.Render("Duration:") + " " + styles.ValueStyle.Render("In progress"))
	}
	content.WriteString("\n")

	// Breaks Section
	if len(m.entry.Breaks) > 0 {
		content.WriteString("\n")
		content.WriteString(styles.SectionStyle.Render("☕ Breaks"))
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

			content.WriteString(styles.BreakStyle.Render(breakText))
			content.WriteString("\n")
		}
	}

	// Notes Section
	if len(m.entry.Notes) > 0 {
		content.WriteString("\n")
		content.WriteString(styles.SectionStyle.Render("📝 Notes"))
		content.WriteString("\n")

		for i, note := range m.entry.Notes {
			noteText := fmt.Sprintf("%d. %s", i+1, note.Contents)
			if len(note.Tags) > 0 {
				noteText += fmt.Sprintf(" [%s]", strings.Join(note.Tags, ", "))
			}
			content.WriteString(styles.NoteStyle.Render(noteText))
			content.WriteString("\n")
		}
	}

	// Help
	content.WriteString("\n")
	content.WriteString(styles.HelpStyle.Render("Press 'q' or 'esc' to quit"))

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
