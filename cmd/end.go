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

// endCmd represents the end command
var endCmd = &cobra.Command{
	Use:   "end",
	Short: "Marks the current workday as finished",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: markDayAsFinished,
}

type endModel struct {
	entry         *journal.JournalEntry
	date          time.Time
	totalWorkTime time.Duration
	validationErr error
	width         int
	height        int
	quitting      bool
}

func (m endModel) Init() tea.Cmd {
	return nil
}

func (m endModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m endModel) View() string {
	if m.quitting {
		return ""
	}

	// Define styles consistent with report commands
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

	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("120"))

	errorStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	var content strings.Builder

	// Title
	dateStr := m.date.Format("Monday, January 2, 2006")
	if m.validationErr != nil {
		content.WriteString(titleStyle.Render(fmt.Sprintf("⚠️  Workday Completed - %s", dateStr)))
	} else {
		content.WriteString(titleStyle.Render(fmt.Sprintf("✅ Workday Completed - %s", dateStr)))
	}
	content.WriteString("\n\n")

	// Work Summary Section
	content.WriteString(sectionStyle.Render("📊 Work Summary"))
	content.WriteString("\n")

	startTime := m.entry.StartTime.Format("15:04")
	content.WriteString(labelStyle.Render("Started:") + " " + valueStyle.Render(startTime))
	content.WriteString("\n")

	endTime := m.entry.EndTime.Format("15:04")
	content.WriteString(labelStyle.Render("Ended:") + " " + valueStyle.Render(endTime))
	content.WriteString("\n")

	hours := int(m.totalWorkTime.Hours())
	minutes := int(m.totalWorkTime.Minutes()) % 60
	durationStr := fmt.Sprintf("%dh %dm", hours, minutes)
	content.WriteString(labelStyle.Render("Duration:") + " " + valueStyle.Render(durationStr))
	content.WriteString("\n")

	// Validation Status
	content.WriteString("\n")
	content.WriteString(sectionStyle.Render("🔍 Validation"))
	content.WriteString("\n")

	if m.validationErr != nil {
		content.WriteString(errorStyle.Render(fmt.Sprintf("⚠️  %s", m.validationErr.Error())))
	} else {
		content.WriteString(successStyle.Render("✅ All validations passed"))
	}
	content.WriteString("\n")

	// Help
	content.WriteString(helpStyle.Render("Press 'q' or 'esc' to quit"))

	return content.String()
}

// markDayAsFinished marks the current day's JournalEntry as finished.
// It loads the journal entries from the file, finds the entry for the current day,
// and sets its EndTime to the current time.
// If no entry is found for the current day, it returns and error.
// After modifying the entry, it saves the updated entries back to the file.
func markDayAsFinished(cmd *cobra.Command, args []string) error {
	// Get current date
	now := time.Now()
	currentDayId := now.Format("20060102")
	dateStr := now.Format("2006-01-02")

	// Load entries
	journalPath := viper.GetString("journalPath")
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	// Get the index of the entry on the slice, -1 if not found
	entry, idx := journal.FetchEntryByID(currentDayId, entries)

	if idx == -1 {
		return fmt.Errorf("No entry found for the current day")
	}

	if !entry.EndTime.IsZero() {
		fmt.Printf("There is already an EndTime for %s. Do you want to override it? (y/N): ", dateStr)
		userInput, err := getUserInput()
		if err != nil {
			return err
		}
		if userInput != "y" {
			fmt.Println("No changes made...")
			return nil
		}
		entries[idx].EndDay()
		fmt.Printf("Data for %s overwrote. Saving...", dateStr)
	} else {
		entries[idx].EndDay()
	}

	validationErr := validateEntry(&entries[idx])
	if validationErr != nil {
		validationNote := journal.Note{Contents: fmt.Sprintf("Validation Error: %s", validationErr)}
		entries[idx].AddNote(validationNote)
	}

	err = journal.SaveEntries(entries, journalPath)
	if err != nil {
		return fmt.Errorf("Failed to save journal entries: %v\n", err)
	}

	// Calculate total work time for display
	totalWorkTime := entries[idx].TotalWorkTime()

	// Create and run the Bubble Tea program for styled summary
	model := endModel{
		entry:         &entries[idx],
		date:          now,
		totalWorkTime: totalWorkTime,
		validationErr: validationErr,
	}

	p := tea.NewProgram(&model)
	_, err = p.Run()
	return err
}

func init() {
	rootCmd.AddCommand(endCmd)
}

func validateEntry(entry *journal.JournalEntry) error {
	minWorkTime, err := time.ParseDuration(viper.GetString("minWorkTime"))
	if err != nil {
		return fmt.Errorf("invalid minimum work time format in config: %v", err)
	}
	lunchTime, err := time.ParseDuration(viper.GetString("lunchTime"))
	if err != nil {
		return fmt.Errorf("invalid lunch time format in config: %v", err)
	}
	maxWorkTime, err := time.ParseDuration(viper.GetString("maxWorkTime"))
	if err != nil {
		return fmt.Errorf("invalid maximum work time format in config: %v", err)
	}

	totalWorkTime := entry.TotalWorkTime()

	// Check if total work time (accounting for breaks) is less than minimum
	if totalWorkTime < minWorkTime {
		return fmt.Errorf("total work time (%s) is less than the minimum required (%s)", totalWorkTime.String(), minWorkTime.String())
	}
	// Check if total work time (accounting for breaks) above allowed maximum
	if totalWorkTime > maxWorkTime {
		return fmt.Errorf("total work time (%s) exceeds the maximum allowed (%s) by %s", totalWorkTime.String(), maxWorkTime.String(), totalWorkTime-maxWorkTime)
	}

	// Check if there's at least one break of `lunchtime` duration
	for _, br := range entry.Breaks {
		if br.Duration() >= lunchTime {
			return nil
		}
	}
	// If not, return an error
	return fmt.Errorf("did not find any breaks during the day that have at least (%s) for lunchtime break.", lunchTime)
}
