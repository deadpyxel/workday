package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/deadpyxel/workday/internal/journal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ExportData represents the structure for JSON exports
type ExportData struct {
	GeneratedAt time.Time               `json:"generated_at"`
	DateRange   string                  `json:"date_range"`
	Entries     []journal.JournalEntry  `json:"entries"`
	Summary     ExportSummary           `json:"summary"`
}

type ExportSummary struct {
	TotalEntries    int           `json:"total_entries"`
	TotalWorkTime   time.Duration `json:"total_work_time"`
	TotalBreakTime  time.Duration `json:"total_break_time"`
	TotalBreaks     int           `json:"total_breaks"`
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export workday data in various formats",
	Long:  "Export workday data to JSON or CSV formats with filtering options",
}

var exportBreaksCmd = &cobra.Command{
	Use:   "breaks",
	Short: "Export break data",
	Long:  "Export break data to JSON or CSV format",
	RunE:  exportBreaks,
}

var exportTimesheetCmd = &cobra.Command{
	Use:   "timesheet",
	Short: "Export timesheet data",
	Long:  "Export complete timesheet data including work time and breaks",
	RunE:  exportTimesheet,
}

func exportBreaks(cmd *cobra.Command, args []string) error {
	format, _ := cmd.Flags().GetString("format")
	output, _ := cmd.Flags().GetString("output")
	dateFilter, _ := cmd.Flags().GetString("date")
	last, _ := cmd.Flags().GetInt("last")

	journalPath := viper.GetString("journalPath")
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	// Filter entries based on flags
	filteredEntries, dateRange, err := filterEntries(entries, dateFilter, last)
	if err != nil {
		return err
	}

	// Extract breaks from filtered entries
	var allBreaks []BreakExportData
	for _, entry := range filteredEntries {
		for i, br := range entry.Breaks {
			breakData := BreakExportData{
				Date:      entry.StartTime.Format("2006-01-02"),
				BreakID:   i + 1,
				StartTime: br.StartTime.Format("15:04:05"),
				EndTime:   "",
				Duration:  "",
				Reason:    br.Reason,
			}
			
			if !br.EndTime.IsZero() {
				breakData.EndTime = br.EndTime.Format("15:04:05")
				breakData.Duration = br.Duration().String()
			}
			
			allBreaks = append(allBreaks, breakData)
		}
	}

	return exportData(allBreaks, format, output, "breaks", dateRange)
}

func exportTimesheet(cmd *cobra.Command, args []string) error {
	format, _ := cmd.Flags().GetString("format")
	output, _ := cmd.Flags().GetString("output")
	dateFilter, _ := cmd.Flags().GetString("date")
	last, _ := cmd.Flags().GetInt("last")

	journalPath := viper.GetString("journalPath")
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	// Filter entries based on flags
	filteredEntries, dateRange, err := filterEntries(entries, dateFilter, last)
	if err != nil {
		return err
	}

	// Create export data
	var timesheetData []TimesheetExportData
	var totalWorkTime, totalBreakTime time.Duration
	var totalBreaks int

	for _, entry := range filteredEntries {
		workTime := entry.TotalWorkTime()
		totalWorkTime += workTime

		var breakTime time.Duration
		for _, br := range entry.Breaks {
			if !br.EndTime.IsZero() {
				breakTime += br.Duration()
			}
		}
		totalBreakTime += breakTime
		totalBreaks += len(entry.Breaks)

		timesheetEntry := TimesheetExportData{
			Date:          entry.StartTime.Format("2006-01-02"),
			StartTime:     entry.StartTime.Format("15:04:05"),
			EndTime:       "",
			WorkTime:      workTime.String(),
			BreakTime:     breakTime.String(),
			NumberBreaks:  len(entry.Breaks),
			Notes:         len(entry.Notes),
		}

		if !entry.EndTime.IsZero() {
			timesheetEntry.EndTime = entry.EndTime.Format("15:04:05")
		}

		timesheetData = append(timesheetData, timesheetEntry)
	}

	// Add summary if JSON format
	if format == "json" {
		exportDataStruct := ExportData{
			GeneratedAt: time.Now(),
			DateRange:   dateRange,
			Entries:     filteredEntries,
			Summary: ExportSummary{
				TotalEntries:   len(filteredEntries),
				TotalWorkTime:  totalWorkTime,
				TotalBreakTime: totalBreakTime,
				TotalBreaks:    totalBreaks,
			},
		}
		return exportData(exportDataStruct, format, output, "timesheet", dateRange)
	}

	return exportData(timesheetData, format, output, "timesheet", dateRange)
}

type BreakExportData struct {
	Date      string `json:"date" csv:"Date"`
	BreakID   int    `json:"break_id" csv:"Break ID"`
	StartTime string `json:"start_time" csv:"Start Time"`
	EndTime   string `json:"end_time" csv:"End Time"`
	Duration  string `json:"duration" csv:"Duration"`
	Reason    string `json:"reason" csv:"Reason"`
}

type TimesheetExportData struct {
	Date         string `json:"date" csv:"Date"`
	StartTime    string `json:"start_time" csv:"Start Time"`
	EndTime      string `json:"end_time" csv:"End Time"`
	WorkTime     string `json:"work_time" csv:"Work Time"`
	BreakTime    string `json:"break_time" csv:"Break Time"`
	NumberBreaks int    `json:"number_breaks" csv:"Number of Breaks"`
	Notes        int    `json:"notes" csv:"Number of Notes"`
}

func filterEntries(entries []journal.JournalEntry, dateFilter string, last int) ([]journal.JournalEntry, string, error) {
	var filteredEntries []journal.JournalEntry
	var dateRange string

	if dateFilter != "" {
		// Parse specific date
		targetDate, err := time.Parse("2006-01-02", dateFilter)
		if err != nil {
			return nil, "", fmt.Errorf("invalid date format. Use YYYY-MM-DD")
		}
		
		targetId := targetDate.Format("20060102")
		for _, entry := range entries {
			if entry.ID == targetId {
				filteredEntries = append(filteredEntries, entry)
				break
			}
		}
		dateRange = dateFilter
	} else if last > 0 {
		// Get last N days
		now := time.Now()
		cutoff := now.AddDate(0, 0, -last)
		
		for _, entry := range entries {
			if entry.StartTime.After(cutoff) {
				filteredEntries = append(filteredEntries, entry)
			}
		}
		dateRange = fmt.Sprintf("Last %d days", last)
	} else {
		// All entries
		filteredEntries = entries
		dateRange = "All time"
	}

	return filteredEntries, dateRange, nil
}

func exportData(data interface{}, format, output, dataType, dateRange string) error {
	var filename string
	if output != "" {
		filename = output
	} else {
		timestamp := time.Now().Format("20060102_150405")
		filename = fmt.Sprintf("workday_%s_%s.%s", dataType, timestamp, format)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	switch format {
	case "json":
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(data)
	case "csv":
		err = writeCSV(file, data)
	default:
		return fmt.Errorf("unsupported format: %s. Use 'json' or 'csv'", format)
	}

	if err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}

	fmt.Printf("✅ Exported %s data (%s) to %s\n", dataType, dateRange, filename)
	return nil
}

func writeCSV(file *os.File, data interface{}) error {
	writer := csv.NewWriter(file)
	defer writer.Flush()

	switch v := data.(type) {
	case []BreakExportData:
		// Write header
		writer.Write([]string{"Date", "Break ID", "Start Time", "End Time", "Duration", "Reason"})
		
		// Write rows
		for _, break_ := range v {
			writer.Write([]string{
				break_.Date,
				strconv.Itoa(break_.BreakID),
				break_.StartTime,
				break_.EndTime,
				break_.Duration,
				break_.Reason,
			})
		}
	case []TimesheetExportData:
		// Write header
		writer.Write([]string{"Date", "Start Time", "End Time", "Work Time", "Break Time", "Number of Breaks", "Number of Notes"})
		
		// Write rows
		for _, entry := range v {
			writer.Write([]string{
				entry.Date,
				entry.StartTime,
				entry.EndTime,
				entry.WorkTime,
				entry.BreakTime,
				strconv.Itoa(entry.NumberBreaks),
				strconv.Itoa(entry.Notes),
			})
		}
	default:
		return fmt.Errorf("unsupported data type for CSV export")
	}

	return nil
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.AddCommand(exportBreaksCmd)
	exportCmd.AddCommand(exportTimesheetCmd)

	// Common flags for export commands
	for _, cmd := range []*cobra.Command{exportBreaksCmd, exportTimesheetCmd} {
		cmd.Flags().StringP("format", "f", "json", "Export format (json or csv)")
		cmd.Flags().StringP("output", "o", "", "Output filename (default: auto-generated)")
		cmd.Flags().StringP("date", "d", "", "Specific date to export (YYYY-MM-DD)")
		cmd.Flags().IntP("last", "l", 0, "Export last N days")
	}
}