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

// breakSpec is the parsed, still-string form of a single break: argument.
type breakSpec struct {
	startStr, endStr, reason string
}

// parsedBackfill holds the result of parsing the non-date backfill arguments.
// Times are kept as strings here; conversion to time.Time happens in runBackfill
// once the date anchor is known.
type parsedBackfill struct {
	startStr   string
	endStr     string
	breakSpecs []breakSpec
	noteSpecs  []string
}

// parseBackfillArgs walks the tokens that follow the positional <date> argument
// and extracts start, end, breaks and notes. Order is irrelevant. Each token is
// split on its FIRST ':' to determine the field (lowercased); the remainder is
// the field value.
func parseBackfillArgs(args []string) (parsedBackfill, error) {
	var result parsedBackfill
	startSeen := false
	endSeen := false

	for _, arg := range args {
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) != 2 {
			return parsedBackfill{}, fmt.Errorf("invalid argument format '%s'. Use field:value", arg)
		}
		field := strings.ToLower(strings.TrimSpace(parts[0]))
		value := parts[1]

		switch field {
		case "start":
			if startSeen {
				return parsedBackfill{}, fmt.Errorf("duplicate field 'start'")
			}
			startSeen = true
			result.startStr = strings.TrimSpace(value)
		case "end":
			if endSeen {
				return parsedBackfill{}, fmt.Errorf("duplicate field 'end'")
			}
			endSeen = true
			result.endStr = strings.TrimSpace(value)
		case "break":
			spec, err := parseBreakSpec(value)
			if err != nil {
				return parsedBackfill{}, err
			}
			result.breakSpecs = append(result.breakSpecs, spec)
		case "note":
			text := strings.TrimSpace(value)
			if text == "" {
				return parsedBackfill{}, fmt.Errorf("note value cannot be empty")
			}
			result.noteSpecs = append(result.noteSpecs, text)
		default:
			return parsedBackfill{}, fmt.Errorf("unknown field '%s'. Available: start, end, break, note", field)
		}
	}

	if !startSeen || !endSeen {
		return parsedBackfill{}, fmt.Errorf("start and end are required. Usage: workday backfill <date> start:HH:MM end:HH:MM [...]")
	}

	return result, nil
}

// parseBreakSpec parses the value portion of a break: token, i.e.
// "HH:MM-HH:MM:reason". It splits on the first '-' to isolate the start time,
// then splits the remainder after the end time's HH:MM to isolate the reason.
// The reason is trimmed and must be non-empty.
func parseBreakSpec(value string) (breakSpec, error) {
	malformed := fmt.Errorf("invalid break '%s'. Use HH:MM-HH:MM:reason", value)

	dashIdx := strings.Index(value, "-")
	if dashIdx == -1 {
		return breakSpec{}, malformed
	}
	startStr := strings.TrimSpace(value[:dashIdx])
	rest := value[dashIdx+1:]

	endStr, reason, err := splitBreakEndAndReason(rest)
	if err != nil {
		return breakSpec{}, malformed
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return breakSpec{}, malformed
	}
	return breakSpec{startStr: strings.TrimSpace(startStr), endStr: strings.TrimSpace(endStr), reason: reason}, nil
}

// splitBreakEndAndReason takes "HH:MM:reason" and returns ("HH:MM", "reason").
// The end time HH:MM itself contains one ':', so the reason follows the SECOND ':'.
func splitBreakEndAndReason(rest string) (string, string, error) {
	first := strings.Index(rest, ":")
	if first == -1 {
		return "", "", fmt.Errorf("malformed")
	}
	second := strings.Index(rest[first+1:], ":")
	if second == -1 {
		return "", "", fmt.Errorf("malformed")
	}
	splitAt := first + 1 + second
	return rest[:splitAt], rest[splitAt+1:], nil
}

var backfillCmd = &cobra.Command{
	Use:   "backfill <date> start:HH:MM end:HH:MM [break:HH:MM-HH:MM:reason ...] [note:text ...]",
	Short: "Create a complete past day's entry in one shot",
	Long: `Backfill creates a full journal entry for a past day you forgot to log:
start, end, breaks and notes, all from CLI arguments. It refuses to touch a day
that already has an entry (use 'workday edit' or 'workday break add --date' for that).

The date is YYYYMMDD. Times are HH:MM. Order of arguments after the date is irrelevant.

Examples:
  workday backfill 20240527 start:09:00 end:17:30
  workday backfill 20240527 start:09:00 end:17:30 break:12:00-13:00:lunch
  workday backfill 20240527 start:09:00 end:17:30 \
    break:12:00-13:00:lunch break:15:00-15:15:coffee \
    note:"Reviewed PRs" note:"Wrapped up release notes"`,
	Args: cobra.MinimumNArgs(3),
	RunE: runBackfill,
}

// backfillAndSave performs all of the backfill work that does NOT involve the
// terminal UI: it loads the journal, refuses on a date collision, parses the
// remaining arguments, builds the entry via journal.NewBackfilledEntry, saves
// it, then runs policy validation (validateEntry from end.go). On a policy
// failure it appends a "Validation Error: ..." note and saves again.
//
// It returns the constructed entry, the policy validation error (nil if the day
// passed policy), and a hard error. A non-nil hard error means nothing was
// written (collision, parse failure, structural validation failure, or I/O
// failure). A non-nil validation error with a nil hard error means the entry WAS
// saved with a warning note.
func backfillAndSave(args []string) (*journal.JournalEntry, error, error) {
	journalPath := viper.GetString("journalPath")
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load journal: %v", err)
	}

	dateAnchor, err := time.Parse("20060102", args[0])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid date '%s'. Use YYYYMMDD (e.g., 20240527)", args[0])
	}
	entryID := dateAnchor.Format("20060102")

	if _, idx := journal.FetchEntryByID(entryID, entries); idx != -1 {
		return nil, nil, fmt.Errorf("entry for %s already exists. Use 'workday edit %s' or 'workday break add --date %s' to modify it.", entryID, entryID, dateAnchor.Format("2006-01-02"))
	}

	parsed, err := parseBackfillArgs(args[1:])
	if err != nil {
		return nil, nil, err
	}

	start, err := anchorTime(dateAnchor, parsed.startStr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid start time '%s'. Use HH:MM", parsed.startStr)
	}
	end, err := anchorTime(dateAnchor, parsed.endStr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid end time '%s'. Use HH:MM", parsed.endStr)
	}

	breaks := make([]journal.Break, 0, len(parsed.breakSpecs))
	for _, bs := range parsed.breakSpecs {
		bStart, err := anchorTime(dateAnchor, bs.startStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid break time '%s'. Use HH:MM", bs.startStr)
		}
		bEnd, err := anchorTime(dateAnchor, bs.endStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid break time '%s'. Use HH:MM", bs.endStr)
		}
		breaks = append(breaks, journal.Break{StartTime: bStart, EndTime: bEnd, Reason: bs.reason})
	}

	notes := make([]journal.Note, 0, len(parsed.noteSpecs))
	for _, text := range parsed.noteSpecs {
		notes = append(notes, journal.Note{Contents: text})
	}

	entry, err := journal.NewBackfilledEntry(dateAnchor, start, end, breaks, notes)
	if err != nil {
		return nil, nil, err
	}

	entries = append(entries, *entry)
	if err := journal.SaveEntries(entries, journalPath); err != nil {
		return nil, nil, fmt.Errorf("failed to save journal entries: %v", err)
	}

	// Policy validation mirrors end.go: on failure, save anyway with a warning note.
	savedIdx := len(entries) - 1
	validationErr := validateEntry(&entries[savedIdx])
	if validationErr != nil {
		validationNote := journal.Note{Contents: fmt.Sprintf("Validation Error: %s", validationErr)}
		entries[savedIdx].AddNote(validationNote)
		if err := journal.SaveEntries(entries, journalPath); err != nil {
			return nil, nil, fmt.Errorf("failed to save journal entries: %v", err)
		}
	}

	return &entries[savedIdx], validationErr, nil
}

// anchorTime parses an HH:MM string and stamps it onto the given date, in the
// date's location. Mirrors the time-anchoring pattern in cmd/break.go:addBreak.
func anchorTime(date time.Time, hhmm string) (time.Time, error) {
	parsed, err := time.Parse("15:04", hhmm)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(date.Year(), date.Month(), date.Day(),
		parsed.Hour(), parsed.Minute(), 0, 0, date.Location()), nil
}

// runBackfill is the cobra RunE handler. It performs the persistence work via
// backfillAndSave, then launches the endModel confirmation TUI.
func runBackfill(cmd *cobra.Command, args []string) error {
	entry, validationErr, err := backfillAndSave(args)
	if err != nil {
		return err
	}

	model := endModel{
		entry:         entry,
		date:          entry.StartTime,
		totalWorkTime: entry.TotalWorkTime(),
		validationErr: validationErr,
	}

	p := tea.NewProgram(&model)
	_, err = p.Run()
	return err
}

func init() {
	rootCmd.AddCommand(backfillCmd)
}
