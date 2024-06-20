package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/deadpyxel/workday/internal/journal"
)

type EditNoteModel struct {
	choices  []journal.Note
	cursor   int
	selected map[int]struct{}
}

func NewEditNoteModel(entry *journal.JournalEntry) *EditNoteModel {
	return &EditNoteModel{choices: entry.Notes, selected: make(map[int]struct{})}
}

type EditNoteState struct {
	Note     *journal.Note
	NewNote  textinput.Model
	NewTags  textinput.Model
	Finished bool
}

func NewEditNoteState(note *journal.Note) *EditNoteState {
	contentInput := textinput.New()
	contentInput.SetValue(note.Contents)
	contentInput.Placeholder = "Enter note contents"

	tagsInput := textinput.New()
	tagsInput.SetValue(strings.Join(note.Tags, ","))
	tagsInput.Placeholder = "Enter tags, comma separated"

	return &EditNoteState{
		Note:    note,
		NewNote: contentInput,
		NewTags: tagsInput,
	}
}
