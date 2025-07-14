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

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Shows current workday status and expected end time",
	Long: `The status command shows your current workday progress including:
- Current work time
- Expected end time based on minimum work requirements
- Time remaining until expected end
- Lunch break status
- Quick validation overview

This helps you plan your day and maintain work-life balance.`,
	RunE: showWorkdayStatus,
}

type statusModel struct {
	entry            *journal.JournalEntry
	date             time.Time
	expectedEndTime  time.Time
	timeRemaining    time.Duration
	currentWorkTime  time.Duration
	hasLunchBreak    bool
	lunchBreakNeeded bool
	width            int
	height           int
	quitting         bool
}

func (m statusModel) Init() tea.Cmd {
	return nil
}

func (m statusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m statusModel) View() string {
	if m.quitting {
		return ""
	}

	var content strings.Builder

	// Title
	dateStr := m.date.Format("Monday, January 2, 2006")
	content.WriteString(styles.TitleStyle.Render(fmt.Sprintf("📊 Workday Status - %s", dateStr)))
	content.WriteString("\n\n")

	// Current Time Section
	content.WriteString(styles.SectionStyle.Render("🕐 Current Progress"))
	content.WriteString("\n")

	startTime := m.entry.StartTime.Format("15:04")
	content.WriteString(styles.LabelStyle.Render("Started:") + " " + styles.ValueStyle.Render(startTime))
	content.WriteString("\n")

	currentTime := time.Now().Format("15:04")
	content.WriteString(styles.LabelStyle.Render("Current:") + " " + styles.ValueStyle.Render(currentTime))
	content.WriteString("\n")

	// Work Time
	hours := int(m.currentWorkTime.Hours())
	minutes := int(m.currentWorkTime.Minutes()) % 60
	workTimeStr := fmt.Sprintf("%dh %dm", hours, minutes)
	content.WriteString(styles.LabelStyle.Render("Work Time:") + " " + styles.ValueStyle.Render(workTimeStr))
	content.WriteString("\n")

	// Expected End Time Section
	content.WriteString("\n")
	content.WriteString(styles.SectionStyle.Render("🎯 Expected End Time"))
	content.WriteString("\n")

	expectedEndStr := m.expectedEndTime.Format("15:04")
	content.WriteString(styles.LabelStyle.Render("Expected End:") + " " + styles.ValueStyle.Render(expectedEndStr))
	content.WriteString("\n")

	// Time Remaining
	if m.timeRemaining > 0 {
		remainingHours := int(m.timeRemaining.Hours())
		remainingMinutes := int(m.timeRemaining.Minutes()) % 60
		remainingStr := fmt.Sprintf("%dh %dm", remainingHours, remainingMinutes)
		content.WriteString(styles.LabelStyle.Render("Time Remaining:") + " " + styles.ValueStyle.Render(remainingStr))
	} else {
		content.WriteString(styles.LabelStyle.Render("Time Remaining:") + " " + styles.SuccessStyle.Render("You can finish!"))
	}
	content.WriteString("\n")

	// Lunch Break Status
	content.WriteString("\n")
	content.WriteString(styles.SectionStyle.Render("🍽️ Lunch Break"))
	content.WriteString("\n")

	if m.hasLunchBreak {
		content.WriteString(styles.SuccessStyle.Render("✅ Lunch break completed"))
	} else {
		content.WriteString(styles.ErrorStyle.Render("⚠️  Lunch break needed (1h minimum)"))
	}
	content.WriteString("\n")

	// Help
	content.WriteString("\n")
	content.WriteString(styles.HelpStyle.Render("Press 'q' or 'esc' to quit"))

	return content.String()
}

func showWorkdayStatus(cmd *cobra.Command, args []string) error {
	// Load configuration
	journalPath := viper.GetString("journalPath")
	minWorkTime, err := time.ParseDuration(viper.GetString("minWorkTime"))
	if err != nil {
		return fmt.Errorf("invalid minimum work time format in config: %v", err)
	}
	lunchTime, err := time.ParseDuration(viper.GetString("lunchTime"))
	if err != nil {
		return fmt.Errorf("invalid lunch time format in config: %v", err)
	}

	// Load journal entries
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	// Find current day entry
	now := time.Now()
	currentDayId := now.Format("20060102")
	entry, _ := journal.FetchEntryByID(currentDayId, entries)
	if entry == nil {
		return fmt.Errorf("no entry found for today. Start your workday first with 'workday start'")
	}

	// Calculate expected end time
	expectedEndTime, timeRemaining, currentWorkTime := calculateExpectedEndTime(entry, minWorkTime, lunchTime, now)

	// Check lunch break status
	hasLunchBreak := false
	for _, br := range entry.Breaks {
		if br.Duration() >= lunchTime {
			hasLunchBreak = true
			break
		}
	}

	// Create and run the status model
	model := statusModel{
		entry:            entry,
		date:             now,
		expectedEndTime:  expectedEndTime,
		timeRemaining:    timeRemaining,
		currentWorkTime:  currentWorkTime,
		hasLunchBreak:    hasLunchBreak,
		lunchBreakNeeded: !hasLunchBreak,
	}

	p := tea.NewProgram(&model)
	_, err = p.Run()
	return err
}

// calculateExpectedEndTime calculates when the workday should end based on minimum work requirements
func calculateExpectedEndTime(entry *journal.JournalEntry, minWorkTime, lunchTime time.Duration, now time.Time) (time.Time, time.Duration, time.Duration) {
	// Calculate current work time (excluding breaks)
	currentWorkTime := now.Sub(entry.StartTime)
	
	// Subtract completed breaks and handle ongoing breaks
	var totalBreakTime time.Duration
	var ongoingBreakStart time.Time
	for _, br := range entry.Breaks {
		if !br.EndTime.IsZero() {
			// Completed break
			totalBreakTime += br.Duration()
		} else {
			// Ongoing break - don't count the time since break started
			ongoingBreakStart = br.StartTime
		}
	}
	
	// If there's an ongoing break, calculate work time up to the break start
	if !ongoingBreakStart.IsZero() {
		currentWorkTime = ongoingBreakStart.Sub(entry.StartTime)
	}
	
	// Subtract completed breaks from current work time
	currentWorkTime -= totalBreakTime

	// Check if we need to account for lunch break
	hasLunchBreak := false
	for _, br := range entry.Breaks {
		if br.Duration() >= lunchTime {
			hasLunchBreak = true
			break
		}
	}

	// Calculate expected end time
	expectedEndTime := entry.StartTime.Add(minWorkTime)
	expectedEndTime = expectedEndTime.Add(totalBreakTime)
	
	// Add lunch break time if not taken yet
	if !hasLunchBreak {
		expectedEndTime = expectedEndTime.Add(lunchTime)
	}

	// Calculate time remaining
	timeRemaining := expectedEndTime.Sub(now)
	if timeRemaining < 0 {
		timeRemaining = 0
	}

	return expectedEndTime, timeRemaining, currentWorkTime
}

func init() {
	rootCmd.AddCommand(statusCmd)
}