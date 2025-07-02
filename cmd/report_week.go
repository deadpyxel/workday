package cmd

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
	entries  []journal.JournalEntry
	week     time.Time
	width    int
	height   int
	quitting bool
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

	var content strings.Builder

	// Preserve original output format
	for _, entry := range m.entries {
		content.WriteString(fmt.Sprintf("%s\n---\n", entry.String()))
	}

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

	model := reportWeekModel{
		entries: currentWeek,
		week:    now,
	}

	p := tea.NewProgram(&model)
	_, err = p.Run()
	return err
}

func init() {
	reportCmd.AddCommand(reportWeekCmd)
}
