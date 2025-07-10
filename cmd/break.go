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

var breakReason string

var breakCmd = &cobra.Command{
	Use:   "break",
	Short: "Manages work break entries",
	Long:  "The break command allows you to start and stop tracking work breaks.",
}

var breakStartCmd = &cobra.Command{
	Use:   "start [reason]",
	Short: "starts a new work break",
	Long:  "Starts a new work break, recording the start time and reason",
	Args:  cobra.MinimumNArgs(1), // Reason is mandatory
	RunE:  startBreak,
}

func startBreak(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	now := time.Now()
	currentDayId := now.Format("20060102")
	entry, idx := journal.FetchEntryByID(currentDayId, entries)
	if idx == -1 {
		return journal.EntryNotFoundError(currentDayId)
	}

	if len(args) > 0 {
		breakReason = args[0]
	}

	newBreak := journal.Break{
		StartTime: now,
		Reason:    breakReason,
	}

	entry.Breaks = append(entry.Breaks, newBreak)
	entries[idx] = *entry // Update the entry in the slice

	err = journal.SaveEntries(entries, journalPath)
	if err != nil {
		return err
	}

	// Calculate daily break statistics
	totalDayBreaks := len(entry.Breaks)
	var totalBreakTime time.Duration
	for _, br := range entry.Breaks {
		if !br.EndTime.IsZero() {
			totalBreakTime += br.EndTime.Sub(br.StartTime)
		}
	}

	// Create and run the Bubble Tea program for styled confirmation
	model := breakModel{
		isStarting:     true,
		breakTime:      now,
		reason:         breakReason,
		totalDayBreaks: totalDayBreaks,
		totalBreakTime: totalBreakTime,
	}

	p := tea.NewProgram(&model)
	_, err = p.Run()
	return err
}

var breakStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stops a current work break",
	Long:  "Stops the current work break, recording the end time.",
	RunE:  stopBreak,
}

type breakModel struct {
	isStarting      bool
	breakTime       time.Time
	reason          string
	duration        *time.Duration
	totalDayBreaks  int
	totalBreakTime  time.Duration
	width           int
	height          int
	quitting        bool
}

func (m breakModel) Init() tea.Cmd {
	return nil
}

func (m breakModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m breakModel) View() string {
	if m.quitting {
		return ""
	}


	var content strings.Builder

	// Title based on action
	if m.isStarting {
		content.WriteString(styles.TitleStyle.Render("☕ Break Started"))
	} else {
		content.WriteString(styles.TitleStyle.Render("✅ Break Ended"))
	}
	content.WriteString("\n\n")

	// Break Details Section
	content.WriteString(styles.SectionStyle.Render("🕐 Break Details"))
	content.WriteString("\n")

	breakTime := m.breakTime.Format("15:04")
	if m.isStarting {
		content.WriteString(styles.LabelStyle.Render("Started:") + " " + styles.ValueStyle.Render(breakTime))
	} else {
		content.WriteString(styles.LabelStyle.Render("Ended:") + " " + styles.ValueStyle.Render(breakTime))
	}
	content.WriteString("\n")

	if m.reason != "" {
		content.WriteString(styles.LabelStyle.Render("Reason:") + " " + styles.ValueStyle.Render(m.reason))
		content.WriteString("\n")
	}

	if m.duration != nil {
		hours := int(m.duration.Hours())
		minutes := int(m.duration.Minutes()) % 60
		durationStr := fmt.Sprintf("%dh %dm", hours, minutes)
		if hours == 0 {
			durationStr = fmt.Sprintf("%dm", minutes)
		}
		content.WriteString(styles.LabelStyle.Render("Duration:") + " " + styles.SuccessStyle.Render(durationStr))
		content.WriteString("\n")
	}

	// Daily Summary Section
	content.WriteString("\n")
	content.WriteString(styles.SectionStyle.Render("📊 Today's Summary"))
	content.WriteString("\n")

	content.WriteString(styles.LabelStyle.Render("Breaks:") + " " + styles.ValueStyle.Render(fmt.Sprintf("%d taken", m.totalDayBreaks)))
	content.WriteString("\n")

	if m.totalBreakTime > 0 {
		totalHours := int(m.totalBreakTime.Hours())
		totalMinutes := int(m.totalBreakTime.Minutes()) % 60
		totalTimeStr := fmt.Sprintf("%dh %dm", totalHours, totalMinutes)
		if totalHours == 0 {
			totalTimeStr = fmt.Sprintf("%dm", totalMinutes)
		}
		content.WriteString(styles.LabelStyle.Render("Total Time:") + " " + styles.ValueStyle.Render(totalTimeStr))
		content.WriteString("\n")
	}

	// Action guidance
	content.WriteString("\n")
	if m.isStarting {
		content.WriteString(styles.InfoStyle.Render("💡 Stop your break with: workday break stop"))
	} else {
		content.WriteString(styles.SuccessStyle.Render("🎯 Ready to continue working!"))
	}
	content.WriteString("\n")

	// Help
	content.WriteString(styles.HelpStyle.Render("Press 'q' or 'esc' to quit"))

	return content.String()
}

func stopBreak(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	now := time.Now()
	currentDayId := now.Format("20060102")
	entry, idx := journal.FetchEntryByID(currentDayId, entries)
	if idx == -1 {
		return fmt.Errorf("No entry found for the current day.")
	}

	if len(entry.Breaks) == 0 {
		return fmt.Errorf("No break started for the current day.")
	}

	lastBreak := &entry.Breaks[len(entry.Breaks)-1] // Get the last break

	if !lastBreak.EndTime.IsZero() {
		return fmt.Errorf("Last break was already stopped.")
	}
	lastBreak.EndTime = now
	entries[idx] = *entry

	err = journal.SaveEntries(entries, journalPath)
	if err != nil {
		return err
	}

	// Calculate break duration and daily statistics
	breakDuration := now.Sub(lastBreak.StartTime)
	totalDayBreaks := len(entry.Breaks)
	var totalBreakTime time.Duration
	for _, br := range entry.Breaks {
		if !br.EndTime.IsZero() {
			totalBreakTime += br.EndTime.Sub(br.StartTime)
		}
	}

	// Create and run the Bubble Tea program for styled confirmation
	model := breakModel{
		isStarting:     false,
		breakTime:      now,
		reason:         lastBreak.Reason,
		duration:       &breakDuration,
		totalDayBreaks: totalDayBreaks,
		totalBreakTime: totalBreakTime,
	}

	p := tea.NewProgram(&model)
	_, err = p.Run()
	return err
}

func init() {
	rootCmd.AddCommand(breakCmd)
	breakCmd.AddCommand(breakStartCmd)
	breakCmd.AddCommand(breakStopCmd)

	// Add flag to the `start` subcommand
	breakStartCmd.Flags().StringVarP(&breakReason, "reason", "r", "", "Reason for the break")
}
