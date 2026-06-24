package cmd

import (
	"fmt"
	"strconv"
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
	isStarting     bool
	breakTime      time.Time
	reason         string
	duration       *time.Duration
	totalDayBreaks int
	totalBreakTime time.Duration
	width          int
	height         int
	quitting       bool
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

// Break list command
var breakListCmd = &cobra.Command{
	Use:   "list [date]",
	Short: "List breaks for a specific date",
	Long:  "Lists all breaks for today or a specific date (format: YYYY-MM-DD)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  listBreaks,
}

func listBreaks(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	// Determine target date
	var targetDate time.Time
	var entryId string

	if len(args) > 0 {
		targetDate, err = time.Parse("2006-01-02", args[0])
		if err != nil {
			return fmt.Errorf("invalid date format. Use YYYY-MM-DD")
		}
		entryId = targetDate.Format("20060102")
	} else {
		targetDate = time.Now()
		entryId = targetDate.Format("20060102")
	}

	entry, idx := journal.FetchEntryByID(entryId, entries)
	if idx == -1 {
		return journal.EntryNotFoundError(entryId)
	}

	if len(entry.Breaks) == 0 {
		fmt.Printf("No breaks found for %s\n", targetDate.Format("2006-01-02"))
		return nil
	}

	// Create TUI model for break list
	model := breakListModel{
		entry:      entry,
		targetDate: targetDate,
		breaks:     entry.Breaks,
	}

	p := tea.NewProgram(&model)
	_, err = p.Run()
	return err
}

// Break list TUI model
type breakListModel struct {
	entry      *journal.JournalEntry
	targetDate time.Time
	breaks     []journal.Break
	width      int
	height     int
	quitting   bool
}

func (m breakListModel) Init() tea.Cmd {
	return nil
}

func (m breakListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m breakListModel) View() string {
	if m.quitting {
		return ""
	}

	var content strings.Builder

	// Title
	content.WriteString(styles.TitleStyle.Render(fmt.Sprintf("☕ Breaks for %s", m.targetDate.Format("Monday, January 2, 2006"))))
	content.WriteString("\n\n")

	// Break list table
	if len(m.breaks) == 0 {
		content.WriteString(styles.InfoStyle.Render("No breaks recorded for this date"))
		content.WriteString("\n")
	} else {
		// Table header
		content.WriteString(styles.HeaderStyle.Render("ID") + "  ")
		content.WriteString(styles.HeaderStyle.Render("Start Time") + "  ")
		content.WriteString(styles.HeaderStyle.Render("End Time") + "    ")
		content.WriteString(styles.HeaderStyle.Render("Duration") + "  ")
		content.WriteString(styles.HeaderStyle.Render("Reason"))
		content.WriteString("\n")
		content.WriteString(strings.Repeat("─", 80) + "\n")

		// Break rows
		var totalDuration time.Duration
		for i, br := range m.breaks {
			id := fmt.Sprintf("%d", i+1)
			startTime := br.StartTime.Format("15:04")
			endTime := "ongoing"
			duration := "N/A"

			if !br.EndTime.IsZero() {
				endTime = br.EndTime.Format("15:04")
				dur := br.Duration()
				totalDuration += dur
				if dur.Hours() >= 1 {
					duration = fmt.Sprintf("%dh %dm", int(dur.Hours()), int(dur.Minutes())%60)
				} else {
					duration = fmt.Sprintf("%dm", int(dur.Minutes()))
				}
			}

			content.WriteString(styles.CellStyle.Render(fmt.Sprintf("%-2s", id)) + "  ")
			content.WriteString(styles.CellStyle.Render(fmt.Sprintf("%-10s", startTime)) + "  ")
			content.WriteString(styles.CellStyle.Render(fmt.Sprintf("%-10s", endTime)) + "  ")
			content.WriteString(styles.CellStyle.Render(fmt.Sprintf("%-8s", duration)) + "  ")
			content.WriteString(styles.ValueStyle.Render(br.Reason))
			content.WriteString("\n")
		}

		// Summary
		if totalDuration > 0 {
			content.WriteString("\n")
			content.WriteString(styles.SummaryStyle.Render("Summary"))
			content.WriteString("\n")

			totalHours := int(totalDuration.Hours())
			totalMinutes := int(totalDuration.Minutes()) % 60
			totalTimeStr := fmt.Sprintf("%dh %dm", totalHours, totalMinutes)
			if totalHours == 0 {
				totalTimeStr = fmt.Sprintf("%dm", totalMinutes)
			}

			content.WriteString(styles.LabelStyle.Render("Total breaks:") + " " + styles.ValueStyle.Render(fmt.Sprintf("%d", len(m.breaks))))
			content.WriteString("\n")
			content.WriteString(styles.LabelStyle.Render("Total time:") + " " + styles.SuccessStyle.Render(totalTimeStr))
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")
	content.WriteString(styles.InfoStyle.Render("💡 Use 'workday break modify <id> ...' to edit breaks"))
	content.WriteString("\n")
	content.WriteString(styles.HelpStyle.Render("Press 'q' or 'esc' to quit"))

	return content.String()
}

// Break modify command
var breakModifyCmd = &cobra.Command{
	Use:   "modify <id> [field:value...]",
	Short: "Modify a break entry",
	Long: `Modify a break entry using task-style syntax.
	
Examples:
  workday break modify 1 reason:"Doctor appointment"
  workday break modify 2 start:14:30 end:15:15
  workday break modify 3 duration:45m
  workday break modify 1 --date 2024-07-29 reason:"Updated reason"`,
	Args: cobra.MinimumNArgs(2),
	RunE: modifyBreak,
}

var breakDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a break entry",
	Long:  "Delete a specific break entry by ID",
	Args:  cobra.ExactArgs(1),
	RunE:  deleteBreak,
}

var breakAddCmd = &cobra.Command{
	Use:   "add start:HH:MM end:HH:MM reason:text",
	Short: "Add a completed break with explicit start and end times",
	Long: `Add a completed break entry with explicit start and end times.
Useful for retroactively logging breaks you forgot to track in real-time.

Examples:
  workday break add start:12:00 end:13:00 reason:lunch
  workday break add start:15:00 end:15:15 reason:break
  workday break add start:09:30 end:10:00 reason:drive
  workday break add --date 2024-05-27 start:12:00 end:13:00 reason:lunch`,
	Args: cobra.MinimumNArgs(2),
	RunE: addBreak,
}

func modifyBreak(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	// Parse break ID
	breakID := args[0]
	breakIndex, err := parseBreakID(breakID)
	if err != nil {
		return err
	}

	// Get target date
	dateFlag, _ := cmd.Flags().GetString("date")
	var targetDate time.Time
	var entryId string

	if dateFlag != "" {
		targetDate, err = time.Parse("2006-01-02", dateFlag)
		if err != nil {
			return fmt.Errorf("invalid date format. Use YYYY-MM-DD")
		}
		entryId = targetDate.Format("20060102")
	} else {
		targetDate = time.Now()
		entryId = targetDate.Format("20060102")
	}

	entry, idx := journal.FetchEntryByID(entryId, entries)
	if idx == -1 {
		return journal.EntryNotFoundError(entryId)
	}

	if breakIndex >= len(entry.Breaks) {
		return fmt.Errorf("break ID %s not found. Use 'workday break list' to see available breaks", breakID)
	}

	// Parse modifications
	modifications := args[1:]
	originalBreak := entry.Breaks[breakIndex]

	for _, mod := range modifications {
		err = applyBreakModification(&entry.Breaks[breakIndex], mod)
		if err != nil {
			return fmt.Errorf("error applying modification '%s': %v", mod, err)
		}
	}

	// Validate the modified break
	if result := journal.ValidateBreak(entry.Breaks[breakIndex]); !result.IsValid {
		return fmt.Errorf("invalid break modification: %v", result.Error)
	}

	// Save changes
	entries[idx] = *entry
	err = journal.SaveEntries(entries, journalPath)
	if err != nil {
		return err
	}

	// Show confirmation
	fmt.Printf("✅ Break %s modified successfully\n", breakID)
	fmt.Printf("Original: %s %s-%s (%s)\n",
		originalBreak.StartTime.Format("15:04"),
		originalBreak.EndTime.Format("15:04"),
		originalBreak.Reason,
		originalBreak.Duration())

	newBreak := entry.Breaks[breakIndex]
	endTime := "ongoing"
	duration := "N/A"
	if !newBreak.EndTime.IsZero() {
		endTime = newBreak.EndTime.Format("15:04")
		duration = newBreak.Duration().String()
	}

	fmt.Printf("Updated:  %s %s-%s (%s)\n",
		newBreak.StartTime.Format("15:04"),
		endTime,
		newBreak.Reason,
		duration)

	return nil
}

func deleteBreak(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	// Parse break ID
	breakID := args[0]
	breakIndex, err := parseBreakID(breakID)
	if err != nil {
		return err
	}

	// Get target date
	dateFlag, _ := cmd.Flags().GetString("date")
	var targetDate time.Time
	var entryId string

	if dateFlag != "" {
		targetDate, err = time.Parse("2006-01-02", dateFlag)
		if err != nil {
			return fmt.Errorf("invalid date format. Use YYYY-MM-DD")
		}
		entryId = targetDate.Format("20060102")
	} else {
		targetDate = time.Now()
		entryId = targetDate.Format("20060102")
	}

	entry, idx := journal.FetchEntryByID(entryId, entries)
	if idx == -1 {
		return journal.EntryNotFoundError(entryId)
	}

	if breakIndex >= len(entry.Breaks) {
		return fmt.Errorf("break ID %s not found. Use 'workday break list' to see available breaks", breakID)
	}

	// Confirm deletion
	deletedBreak := entry.Breaks[breakIndex]
	fmt.Printf("Delete break: %s %s-%s (%s)? [y/N]: ",
		deletedBreak.StartTime.Format("15:04"),
		deletedBreak.EndTime.Format("15:04"),
		deletedBreak.Reason,
		deletedBreak.Duration())

	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("Deletion cancelled")
		return nil
	}

	// Remove break from slice
	entry.Breaks = append(entry.Breaks[:breakIndex], entry.Breaks[breakIndex+1:]...)

	// Save changes
	entries[idx] = *entry
	err = journal.SaveEntries(entries, journalPath)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Break %s deleted successfully\n", breakID)
	return nil
}

// addBreakToJournal loads the journal, resolves the target day (today when
// dateFlag is empty, otherwise the YYYY-MM-DD date), builds a completed break
// from the field:value args anchored to the target date, validates it, and
// persists the change. It returns the updated entry and its slice index so the
// caller can render a confirmation. All times-of-day are anchored to the
// resolved target date rather than to now.
func addBreakToJournal(journalPath string, dateFlag string, now time.Time, args []string) (*journal.JournalEntry, int, error) {
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return nil, -1, err
	}

	// Resolve the target date.
	var targetDate time.Time
	var entryId string
	if dateFlag != "" {
		// Parse in the local zone so the break's times-of-day are anchored to
		// the same timezone that real-time commands (break start) use. Plain
		// time.Parse defaults to UTC, which would shift the break by the local
		// UTC offset and cause false overlaps against locally-stored breaks.
		targetDate, err = time.ParseInLocation("2006-01-02", dateFlag, time.Local)
		if err != nil {
			return nil, -1, fmt.Errorf("invalid date format. Use YYYY-MM-DD")
		}
		entryId = targetDate.Format("20060102")
	} else {
		targetDate = now
		entryId = targetDate.Format("20060102")
	}

	entry, idx := journal.FetchEntryByID(entryId, entries)
	if idx == -1 {
		if dateFlag != "" {
			return nil, -1, fmt.Errorf("no entry found for %s; use 'workday backfill' to create it", dateFlag)
		}
		return nil, -1, journal.EntryNotFoundError(entryId)
	}

	// Parse field:value arguments
	var startStr, endStr, reason string
	for _, arg := range args {
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) != 2 {
			return nil, -1, fmt.Errorf("invalid argument format '%s'. Use field:value (e.g., start:12:00)", arg)
		}
		field := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch field {
		case "start":
			startStr = value
		case "end":
			endStr = value
		case "reason":
			reason = value
		default:
			return nil, -1, fmt.Errorf("unknown field '%s'. Available fields: start, end, reason", field)
		}
	}

	// Validate required fields
	if startStr == "" {
		return nil, -1, fmt.Errorf("start time is required. Usage: workday break add start:HH:MM end:HH:MM reason:text")
	}
	if endStr == "" {
		return nil, -1, fmt.Errorf("end time is required. Usage: workday break add start:HH:MM end:HH:MM reason:text")
	}
	if strings.TrimSpace(reason) == "" {
		return nil, -1, fmt.Errorf("reason is required. Usage: workday break add start:HH:MM end:HH:MM reason:text")
	}

	// Parse start time
	startTime, err := time.Parse("15:04", startStr)
	if err != nil {
		return nil, -1, fmt.Errorf("invalid start time format '%s'. Use HH:MM", startStr)
	}

	// Parse end time
	endTime, err := time.Parse("15:04", endStr)
	if err != nil {
		return nil, -1, fmt.Errorf("invalid end time format '%s'. Use HH:MM", endStr)
	}

	// Construct the break using the TARGET date with the specified times
	newBreak := journal.Break{
		StartTime: time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(),
			startTime.Hour(), startTime.Minute(), 0, 0, targetDate.Location()),
		EndTime: time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(),
			endTime.Hour(), endTime.Minute(), 0, 0, targetDate.Location()),
		Reason: reason,
	}

	// Validate the break (checks start not zero, end after start, reason non-empty)
	if result := journal.ValidateBreak(newBreak); !result.IsValid {
		return nil, -1, fmt.Errorf("invalid break: %v", result.Error)
	}

	// Validate no overlap with existing breaks
	if result := journal.ValidateBreakOverlap(newBreak, entry.Breaks); !result.IsValid {
		return nil, -1, fmt.Errorf("cannot add break: %v", result.Error)
	}

	// Append the break and save
	entry.Breaks = append(entry.Breaks, newBreak)
	entries[idx] = *entry

	if err := journal.SaveEntries(entries, journalPath); err != nil {
		return nil, -1, err
	}

	return entry, idx, nil
}

func addBreak(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	dateFlag, _ := cmd.Flags().GetString("date")

	entry, _, err := addBreakToJournal(journalPath, dateFlag, time.Now(), args)
	if err != nil {
		return err
	}

	// The last appended break is the one we just added.
	newBreak := entry.Breaks[len(entry.Breaks)-1]

	// Calculate daily break statistics
	totalDayBreaks := len(entry.Breaks)
	var totalBreakTime time.Duration
	for _, br := range entry.Breaks {
		if !br.EndTime.IsZero() {
			totalBreakTime += br.EndTime.Sub(br.StartTime)
		}
	}

	// Display TUI confirmation
	breakDuration := newBreak.EndTime.Sub(newBreak.StartTime)
	model := breakModel{
		isStarting:     false,
		breakTime:      newBreak.EndTime,
		reason:         newBreak.Reason,
		duration:       &breakDuration,
		totalDayBreaks: totalDayBreaks,
		totalBreakTime: totalBreakTime,
	}

	p := tea.NewProgram(&model)
	_, err = p.Run()
	return err
}

func parseBreakID(id string) (int, error) {
	breakIndex, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("invalid break ID '%s'. Expected a number", id)
	}
	if breakIndex < 1 {
		return 0, fmt.Errorf("break ID must be >= 1")
	}
	return breakIndex - 1, nil // Convert to 0-based index
}

func applyBreakModification(br *journal.Break, modification string) error {
	parts := strings.SplitN(modification, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid modification format. Use field:value")
	}

	field := strings.ToLower(strings.TrimSpace(parts[0]))
	value := strings.TrimSpace(parts[1])

	switch field {
	case "reason":
		br.Reason = value
	case "start":
		startTime, err := time.Parse("15:04", value)
		if err != nil {
			return fmt.Errorf("invalid start time format. Use HH:MM")
		}
		// Keep the same date, just change time
		br.StartTime = time.Date(br.StartTime.Year(), br.StartTime.Month(), br.StartTime.Day(),
			startTime.Hour(), startTime.Minute(), 0, 0, br.StartTime.Location())
	case "end":
		endTime, err := time.Parse("15:04", value)
		if err != nil {
			return fmt.Errorf("invalid end time format. Use HH:MM")
		}
		// Keep the same date, just change time
		br.EndTime = time.Date(br.StartTime.Year(), br.StartTime.Month(), br.StartTime.Day(),
			endTime.Hour(), endTime.Minute(), 0, 0, br.StartTime.Location())
	case "duration":
		duration, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid duration format. Use formats like '30m', '1h30m'")
		}
		br.EndTime = br.StartTime.Add(duration)
	default:
		return fmt.Errorf("unknown field '%s'. Available fields: reason, start, end, duration", field)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(breakCmd)
	breakCmd.AddCommand(breakStartCmd)
	breakCmd.AddCommand(breakStopCmd)
	breakCmd.AddCommand(breakListCmd)
	breakCmd.AddCommand(breakModifyCmd)
	breakCmd.AddCommand(breakDeleteCmd)
	breakCmd.AddCommand(breakAddCmd)

	// Add flags
	breakStartCmd.Flags().StringVarP(&breakReason, "reason", "r", "", "Reason for the break")
	breakModifyCmd.Flags().StringP("date", "d", "", "Target date (YYYY-MM-DD)")
	breakDeleteCmd.Flags().StringP("date", "d", "", "Target date (YYYY-MM-DD)")
	breakAddCmd.Flags().StringP("date", "d", "", "Target date (YYYY-MM-DD)")
}
