package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m EditNoteModel) Init() tea.Cmd {
	return nil
}

func (m EditNoteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m EditNoteModel) View() string {
	s := "Which Note do you want to edit?\n\n"
	for i, note := range m.choices {
		// Is the cursor pointing at this choice?
		cursor := " " // No cursor
		if m.cursor == i {
			cursor = ">" // Render cursor
		}

		// Is this choice selected?
		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		// Render the row
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, &note)
	}

	s += "\nPress q to quit.\n"

	return s
}

func (s *EditNoteState) Init() tea.Cmd {
	return textinput.Blink
}

func (s *EditNoteState) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return s, tea.Quit
		case "enter":
			s.Finished = true
			return s, tea.Quit
		}
	default:
		var cmd tea.Cmd
		s.NewNote, cmd = s.NewNote.Update(msg)
		// s.NewTags, cmd = s.NewTags.Update(msg)
		return s, cmd
	}
	return s, nil
}

func (s *EditNoteState) View() string {
	if s.Finished {
		return ""
	}
	return fmt.Sprintf("Editing note:\n\n%s\n\nTags: %s\n\nPress Enter to save, Ctrl+C to cancel.", s.NewNote.View(), s.NewTags.View())
}
