package cmd

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/deadpyxel/workday/internal/journal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit [date]",
	Args:  cobra.MaximumNArgs(1),
	Short: "Interactive editor for journal entries",
	Long: `The edit command provides an interactive TUI for editing journal entries.

You can specify a date in YYYYMMDD format, or if no date is provided, 
it will edit today's entry. The TUI allows you to edit notes, times, 
and other entry details in a form-like interface.

Examples:
  workday edit           # Edit today's entry
  workday edit 20231201  # Edit entry for December 1, 2023`,
	RunE: runEditTUI,
}

type editModel struct {
	entry       *journal.JournalEntry
	entryIndex  int
	entries     []journal.JournalEntry
	journalPath string

	// Form fields
	inputs  []textinput.Model
	focused int

	// UI state
	width    int
	height   int
	quitting bool
	saved    bool
}

const (
	inputStartTime = iota
	inputEndTime
	inputNotes
)

func (m editModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m editModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "ctrl+s":
			if err := m.saveEntry(); err != nil {
				// Handle error - in a real app you might show an error message
				return m, tea.Quit
			}
			m.saved = true
			m.quitting = true
			return m, tea.Quit

		case "tab", "down":
			m.nextInput()

		case "shift+tab", "up":
			m.prevInput()

		case "enter":
			if m.focused == inputNotes {
				// Allow multiline input for notes
				m.inputs[m.focused].SetValue(m.inputs[m.focused].Value() + "\n")
			} else {
				m.nextInput()
			}
		}
	}

	// Update focused input
	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m editModel) View() string {
	if m.quitting {
		if m.saved {
			return "Entry saved successfully!\n"
		}
		return "Edit cancelled.\n"
	}

	var s string

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		PaddingBottom(1)

	s += headerStyle.Render(fmt.Sprintf("Editing Journal Entry: %s", m.entry.ID))
	s += "\n\n"

	// Form fields
	fieldStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		PaddingRight(2)

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Width(12)

	// Start Time
	s += fieldStyle.Render(
		labelStyle.Render("Start Time:")+" "+m.inputs[inputStartTime].View(),
	) + "\n\n"

	// End Time
	s += fieldStyle.Render(
		labelStyle.Render("End Time:")+" "+m.inputs[inputEndTime].View(),
	) + "\n\n"

	// Notes
	s += fieldStyle.Render(
		labelStyle.Render("Notes:")+"\n"+m.inputs[inputNotes].View(),
	) + "\n\n"

	// Instructions
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		PaddingTop(1)

	s += helpStyle.Render(
		"Tab/↑↓: Navigate • Ctrl+S: Save • Esc: Cancel",
	)

	return s
}

func (m *editModel) nextInput() {
	m.inputs[m.focused].Blur()
	m.focused = (m.focused + 1) % len(m.inputs)
	m.inputs[m.focused].Focus()
}

func (m *editModel) prevInput() {
	m.inputs[m.focused].Blur()
	m.focused--
	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
	m.inputs[m.focused].Focus()
}

func (m *editModel) saveEntry() error {
	// Parse and update start time
	if startTimeStr := m.inputs[inputStartTime].Value(); startTimeStr != "" {
		if startTime, err := time.Parse("15:04", startTimeStr); err == nil {
			// Keep the original date, update time
			originalDate := m.entry.StartTime
			m.entry.StartTime = time.Date(
				originalDate.Year(), originalDate.Month(), originalDate.Day(),
				startTime.Hour(), startTime.Minute(), 0, 0, originalDate.Location(),
			)
		}
	}

	// Parse and update end time
	if endTimeStr := m.inputs[inputEndTime].Value(); endTimeStr != "" {
		if endTime, err := time.Parse("15:04", endTimeStr); err == nil {
			// Keep the original date, update time
			originalDate := m.entry.StartTime
			m.entry.EndTime = time.Date(
				originalDate.Year(), originalDate.Month(), originalDate.Day(),
				endTime.Hour(), endTime.Minute(), 0, 0, originalDate.Location(),
			)
		}
	}

	// Update notes
	notesText := m.inputs[inputNotes].Value()
	if notesText != "" {
		// Clear existing notes and add new ones
		m.entry.Notes = []journal.Note{}
		// Split by lines and create notes
		lines := splitLines(notesText)
		for _, line := range lines {
			if line != "" {
				m.entry.Notes = append(m.entry.Notes, journal.Note{Contents: line})
			}
		}
	}

	// Update the entry in the slice
	m.entries[m.entryIndex] = *m.entry

	// Save to file
	return journal.SaveEntries(m.entries, m.journalPath)
}

func splitLines(text string) []string {
	var lines []string
	current := ""
	for _, char := range text {
		if char == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func runEditTUI(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return fmt.Errorf("failed to load journal: %w", err)
	}

	// Determine which entry to edit
	var targetDate string
	if len(args) > 0 {
		targetDate = args[0]
	} else {
		targetDate = time.Now().Format("20060102")
	}

	entry, idx := journal.FetchEntryByID(targetDate, entries)
	if idx == -1 {
		return fmt.Errorf("no entry found for date: %s", targetDate)
	}

	// Create text inputs
	inputs := make([]textinput.Model, 3)

	// Start time input
	inputs[inputStartTime] = textinput.New()
	inputs[inputStartTime].Placeholder = "09:00"
	inputs[inputStartTime].SetValue(entry.StartTime.Format("15:04"))
	inputs[inputStartTime].Width = 20
	inputs[inputStartTime].Focus()

	// End time input
	inputs[inputEndTime] = textinput.New()
	inputs[inputEndTime].Placeholder = "17:30"
	if !entry.EndTime.IsZero() {
		inputs[inputEndTime].SetValue(entry.EndTime.Format("15:04"))
	}
	inputs[inputEndTime].Width = 20

	// Notes input
	inputs[inputNotes] = textinput.New()
	inputs[inputNotes].Placeholder = "Enter your notes..."
	inputs[inputNotes].Width = 60
	// Combine all notes into a single text field
	var notesText string
	for i, note := range entry.Notes {
		if i > 0 {
			notesText += "\n"
		}
		notesText += note.Contents
	}
	inputs[inputNotes].SetValue(notesText)

	model := editModel{
		entry:       entry,
		entryIndex:  idx,
		entries:     entries,
		journalPath: journalPath,
		inputs:      inputs,
		focused:     0,
	}

	p := tea.NewProgram(&model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func init() {
	rootCmd.AddCommand(editCmd)
}

