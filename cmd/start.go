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

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a new workday entry",
	Long: `The start command is used to begin a new workday entry in the journal.

It creates a new JournalEntry with the current date and time as the start time,
appends it to the existing journal entries, and saves the updated journal entries to the file.
After running this command, you can begin adding notes to the new workday entry.`,
	RunE: startWorkDay,
}

type startModel struct {
	startTime       time.Time
	isNewEntry      bool
	previousEndTime *time.Time
	width           int
	height          int
	quitting        bool
}

func (m startModel) Init() tea.Cmd {
	return nil
}

func (m startModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m startModel) View() string {
	if m.quitting {
		return ""
	}

	// Define styles consistent with other commands
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

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	var content strings.Builder

	// Title
	dateStr := m.startTime.Format("Monday, January 2, 2006")
	if m.isNewEntry {
		content.WriteString(titleStyle.Render(fmt.Sprintf("🚀 Workday Started - %s", dateStr)))
	} else {
		content.WriteString(titleStyle.Render(fmt.Sprintf("🔄 Workday Restarted - %s", dateStr)))
	}
	content.WriteString("\n\n")

	// Start Details Section
	content.WriteString(sectionStyle.Render("⏰ Start Details"))
	content.WriteString("\n")

	startTime := m.startTime.Format("15:04")
	content.WriteString(labelStyle.Render("Time:") + " " + valueStyle.Render(startTime))
	content.WriteString("\n")

	if m.isNewEntry {
		content.WriteString(labelStyle.Render("Status:") + " " + successStyle.Render("New workday entry created"))
	} else {
		content.WriteString(labelStyle.Render("Status:") + " " + infoStyle.Render("Existing entry overwritten"))
	}
	content.WriteString("\n")

	// Previous Entry Info (if applicable)
	if m.previousEndTime != nil {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("📝 Previous Entry"))
		content.WriteString("\n")
		
		prevEndTime := m.previousEndTime.Format("15:04 on Jan 2")
		content.WriteString(labelStyle.Render("End Time:") + " " + valueStyle.Render(fmt.Sprintf("Updated to %s", prevEndTime)))
		content.WriteString("\n")
	}

	// Next Steps
	content.WriteString("\n")
	content.WriteString(sectionStyle.Render("📋 Next Steps"))
	content.WriteString("\n")
	content.WriteString(infoStyle.Render("• Add notes with: workday note \"Your note here\""))
	content.WriteString("\n")
	content.WriteString(infoStyle.Render("• Take breaks with: workday break start \"lunch\""))
	content.WriteString("\n")
	content.WriteString(infoStyle.Render("• End your day with: workday end"))
	content.WriteString("\n")

	// Help
	content.WriteString(helpStyle.Render("Press 'q' or 'esc' to quit"))

	return content.String()
}

// startWorkDay starts a new workday entry in the journal.
// It first loads the existing journal entries from the file.
// If the last entry on the journal was not closed properly (misisng EndTime) it will offer to update that first
// If there is already an entry for the current day, it asks the user if they want to override it.
// If the user agrees, it overwrites the existing entry with a new one.
// If the user does not agree, it does nothing and returns nil.
// If there is no entry for the current day, it creates a new JournalEntry with the current date and time as the start time,
// appends the new entry to the journal entries, and saves the updated journal entries back to the file.
// It then prints a message indicating that a new JournalEntry has been added for the current day.
func startWorkDay(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	now := time.Now()
	currentDayId := now.Format("20060102")
	var lastEntry *journal.JournalEntry
	var previousEndTime *time.Time

	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].ID[:8] != currentDayId {
			lastEntry = &entries[i]
			break
		}
	}
	// If the last entry has no EndTime set (considering it is not the entry for the current day)
	if lastEntry != nil && lastEntry.EndTime.IsZero() {
		fmt.Println("Warning: Last entry has no EndTime set.")
		fmt.Printf("Do you want to set the EndTime for the last entry? (y/N): ")
		userInput, err := getUserInput()
		if err != nil {
			return err
		}
		if userInput == "y" {
			fmt.Printf("Please type the EndTime in HH:MM format: ")
			endTimeStr, err := getUserInput()
			if err != nil {
				return err
			}
			endTime, err := time.Parse("15:04", endTimeStr)
			if err != nil {
				return err
			}
			finalEndTime := time.Date(
				lastEntry.StartTime.Year(),
				lastEntry.StartTime.Month(),
				lastEntry.StartTime.Day(),
				endTime.Hour(),
				endTime.Minute(),
				0,
				0,
				lastEntry.StartTime.Location())
			lastEntry.EndTime = finalEndTime
			previousEndTime = &finalEndTime

			err = journal.SaveEntries(entries, journalPath)
			if err != nil {
				return err
			}

			fmt.Println("Endtime set for the last entry.")
		}
	}

	dateStr := now.Format("2006-01-02")
	_, idx := journal.FetchEntryByID(currentDayId, entries)
	isNewEntry := idx == -1

	if idx != -1 {
		fmt.Printf("There is already an entry for %s. Do you want to override it? (y/N): ", dateStr)
		userInput, err := getUserInput()
		if err != nil {
			return err
		}
		if userInput != "y" {
			fmt.Println("No changes made...")
			return nil
		}
		entries[idx] = *journal.NewJournalEntry()
	} else {
		newEntry := journal.NewJournalEntry()
		entries = append(entries, *newEntry)
	}

	err = journal.SaveEntries(entries, journalPath)
	if err != nil {
		return err
	}

	// Create and run the Bubble Tea program for styled confirmation
	model := startModel{
		startTime:       now,
		isNewEntry:      isNewEntry,
		previousEndTime: previousEndTime,
	}

	p := tea.NewProgram(&model)
	_, err = p.Run()
	return err
}

func getUserInput() (string, error) {

	var userInput string
	_, err := fmt.Scanln(&userInput)
	if err != nil {
		return "", err
	}
	return strings.ToLower(userInput), nil
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
