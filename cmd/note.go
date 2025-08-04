package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/deadpyxel/workday/internal/journal"
	"github.com/deadpyxel/workday/internal/styles"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// noteCmd represents the note command
var noteCmd = &cobra.Command{
	Use:   "note [note]",
	Args:  cobra.MaximumNArgs(1),
	Short: "Adds a note to the current workday entry",
	Long: `The note command is used to add a note to the current workday entry.

Usage modes:
  workday note                    # Interactive TUI mode for note entry
  workday note "Your note here"   # Quick mode - directly add note

The interactive mode allows you to write notes with hashtag support for automatic
tag extraction (e.g., "Meeting done #progress #team"). Tags can also be specified
using the --tags flag in quick mode.`,
	RunE: handleNoteCommand,
}

type noteModel struct {
	textInput textinput.Model
	entry     *journal.JournalEntry
	entryIdx  int
	entries   []journal.JournalEntry

	journalPath string
	width       int
	height      int
	quitting    bool
	saved       bool
	err         error
}

func (m noteModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m noteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "ctrl+s", "enter":
			if strings.TrimSpace(m.textInput.Value()) == "" {
				return m, nil
			}

			// Create note with tag parsing
			note := journal.Note{Contents: m.textInput.Value()}
			note.ParseContent()

			// Validate note
			if result := journal.ValidateNote(note); !result.IsValid {
				m.err = result.Error
				return m, nil
			}

			// Add note to entry
			m.entries[m.entryIdx].Notes = append(m.entries[m.entryIdx].Notes, note)

			// Save entries
			if err := journal.SaveEntries(m.entries, m.journalPath); err != nil {
				m.err = err
				return m, nil
			}

			m.saved = true
			m.quitting = true
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m noteModel) View() string {
	if m.quitting {
		if m.saved {
			return styles.SuccessStyle.Render("✅ Note saved successfully!")
		}
		return ""
	}

	var content strings.Builder

	// Title
	content.WriteString(styles.TitleStyle.Render("📝 Add Note"))
	content.WriteString("\n\n")

	// Instructions
	content.WriteString(styles.SectionStyle.Render("💡 Instructions"))
	content.WriteString("\n")
	content.WriteString(styles.InfoBlueStyle.Render("• Use hashtags for automatic tagging: #progress #meeting"))
	content.WriteString("\n")
	content.WriteString(styles.InfoBlueStyle.Render("• Press Enter or Ctrl+S to save"))
	content.WriteString("\n")
	content.WriteString(styles.InfoBlueStyle.Render("• Press Esc or Ctrl+C to cancel"))
	content.WriteString("\n\n")

	// Input field
	content.WriteString(styles.SectionStyle.Render("✏️  Note Content"))
	content.WriteString("\n")
	content.WriteString(m.textInput.View())
	content.WriteString("\n\n")

	// Error display
	if m.err != nil {
		content.WriteString(styles.ErrorStyle.Render(fmt.Sprintf("❌ Error: %s", m.err.Error())))
		content.WriteString("\n\n")
	}

	// Help
	content.WriteString(styles.HelpStyle.Render("Enter/Ctrl+S: save • Esc/Ctrl+C: cancel"))

	return content.String()
}

// handleNoteCommand handles both interactive and quick note modes
func handleNoteCommand(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		// Interactive mode
		return runInteractiveNoteEntry()
	} else {
		// Quick mode
		return addNoteToCurrentDay(cmd, args)
	}
}

// runInteractiveNoteEntry starts the TUI for interactive note entry
func runInteractiveNoteEntry() error {
	journalPath := viper.GetString("journalPath")
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	// Find current day entry
	_, idx, err := journal.FindCurrentDayEntry(entries)
	if err != nil {
		fmt.Println("Please run `workday start` first to create a new entry.")
		return err
	}

	// Create text input
	ti := textinput.New()
	ti.Placeholder = "Enter your note here... (use #tags for automatic tagging)"
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 80

	// Create model
	model := noteModel{
		textInput:   ti,
		entry:       &entries[idx],
		entryIdx:    idx,
		entries:     entries,
		journalPath: journalPath,
	}

	// Run TUI
	p := tea.NewProgram(&model)
	_, err = p.Run()
	return err
}

// addNoteToCurrentDay adds a note to the current workday entry.
// It first loads the existing journal entries from the file.
// If there is no entry for the current day, it prints an error message and returns an error.
// Otherwise, it adds the note to the current entry and saves the updated journal entries back to the file.
func addNoteToCurrentDay(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	// Find current day entry using validation helper
	_, idx, err := journal.FindCurrentDayEntry(entries)
	if err != nil {
		fmt.Println("Please run `workday start` first to create a new entry.")
		return err
	}

	// Create note with content and manual tags
	note := journal.Note{Contents: args[0]}

	// Parse hashtags from content
	note.ParseContent()

	// Add manual tags from flag if provided
	if tags != "" {
		tagList := strings.Split(tags, ",")
		// Clean up tags and add them
		for _, tag := range tagList {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				// Check for duplicates
				exists := false
				for _, existingTag := range note.Tags {
					if existingTag == tag {
						exists = true
						break
					}
				}
				if !exists {
					note.Tags = append(note.Tags, tag)
				}
			}
		}
	}

	// Validate note
	if result := journal.ValidateNote(note); !result.IsValid {
		return result.Error
	}

	entries[idx].Notes = append(entries[idx].Notes, note)
	err = journal.SaveEntries(entries, journalPath)
	if err != nil {
		return err
	}
	fmt.Println("Successfully added new note to current day.")
	return nil
}

var tags string

func init() {
	noteCmd.Flags().StringVarP(&tags, "tags", "t", "", "Tags for the note")
	rootCmd.AddCommand(noteCmd)
}
