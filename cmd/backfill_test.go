package cmd

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/deadpyxel/workday/internal/journal"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
)

func TestParseBackfillArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		want      parsedBackfill
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "missing start",
			args:      []string{"end:17:30"},
			wantErr:   true,
			errSubstr: "start and end are required",
		},
		{
			name:      "missing end",
			args:      []string{"start:09:00"},
			wantErr:   true,
			errSubstr: "start and end are required",
		},
		{
			name:      "duplicate start",
			args:      []string{"start:09:00", "start:10:00", "end:17:30"},
			wantErr:   true,
			errSubstr: "duplicate field 'start'",
		},
		{
			name:      "duplicate end",
			args:      []string{"start:09:00", "end:17:30", "end:18:00"},
			wantErr:   true,
			errSubstr: "duplicate field 'end'",
		},
		{
			name:      "unknown field",
			args:      []string{"start:09:00", "end:17:30", "lunch:12:00"},
			wantErr:   true,
			errSubstr: "unknown field 'lunch'. Available: start, end, break, note",
		},
		{
			name:      "token without colon",
			args:      []string{"start:09:00", "end:17:30", "garbage"},
			wantErr:   true,
			errSubstr: "invalid argument format 'garbage'. Use field:value",
		},
		{
			name:      "empty note value",
			args:      []string{"start:09:00", "end:17:30", "note:   "},
			wantErr:   true,
			errSubstr: "note value cannot be empty",
		},
		{
			name:      "empty break reason",
			args:      []string{"start:09:00", "end:17:30", "break:12:00-13:00:   "},
			wantErr:   true,
			errSubstr: "invalid break",
		},
		{
			name:      "malformed break range no dash",
			args:      []string{"start:09:00", "end:17:30", "break:12:00"},
			wantErr:   true,
			errSubstr: "invalid break",
		},
		{
			name:      "malformed break missing reason",
			args:      []string{"start:09:00", "end:17:30", "break:12:00-13:00"},
			wantErr:   true,
			errSubstr: "invalid break",
		},
		{
			name: "happy minimal",
			args: []string{"start:09:00", "end:17:30"},
			want: parsedBackfill{startStr: "09:00", endStr: "17:30"},
		},
		{
			name: "happy with breaks",
			args: []string{"start:09:00", "end:17:30", "break:12:00-13:00:lunch", "break:15:00-15:15:coffee"},
			want: parsedBackfill{
				startStr: "09:00",
				endStr:   "17:30",
				breakSpecs: []breakSpec{
					{startStr: "12:00", endStr: "13:00", reason: "lunch"},
					{startStr: "15:00", endStr: "15:15", reason: "coffee"},
				},
			},
		},
		{
			name: "happy with notes",
			args: []string{"start:09:00", "end:17:30", "note:Reviewed PRs", "note:Release notes #done"},
			want: parsedBackfill{
				startStr:  "09:00",
				endStr:    "17:30",
				noteSpecs: []string{"Reviewed PRs", "Release notes #done"},
			},
		},
		{
			name: "happy mixed order",
			args: []string{"note:first", "break:12:00-13:00:lunch", "end:17:30", "start:09:00"},
			want: parsedBackfill{
				startStr:   "09:00",
				endStr:     "17:30",
				breakSpecs: []breakSpec{{startStr: "12:00", endStr: "13:00", reason: "lunch"}},
				noteSpecs:  []string{"first"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBackfillArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseBackfillArgs(%v) expected error, got nil", tt.args)
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("parseBackfillArgs(%v) error = %q, want substring %q", tt.args, err.Error(), tt.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseBackfillArgs(%v) unexpected error: %v", tt.args, err)
			}
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(parsedBackfill{}, breakSpec{})); diff != "" {
				t.Errorf("parseBackfillArgs(%v) mismatch (-want +got):\n%s", tt.args, diff)
			}
		})
	}
}

// writeTempJournal creates a temp journal file seeded with the given entries
// and returns its path. t.Cleanup removes it.
func writeTempJournal(t *testing.T, entries []journal.JournalEntry) string {
	t.Helper()
	f, err := os.CreateTemp("", "journal_backfill_*.json")
	if err != nil {
		t.Fatal(err)
	}
	name := f.Name()
	f.Close()
	t.Cleanup(func() { os.Remove(name) })
	if err := journal.SaveEntries(entries, name); err != nil {
		t.Fatalf("failed to seed temp journal: %v", err)
	}
	return name
}

// setBackfillViper configures viper for backfill tests, restoring prior state
// via t.Cleanup. Matches the root_test.go save/restore pattern.
func setBackfillViper(t *testing.T, journalPath, minWork, lunch, maxWork string) {
	t.Helper()
	original := viper.AllSettings()
	t.Cleanup(func() {
		viper.Reset()
		for k, v := range original {
			viper.Set(k, v)
		}
	})
	viper.Set("journalPath", journalPath)
	viper.Set("minWorkTime", minWork)
	viper.Set("lunchTime", lunch)
	viper.Set("maxWorkTime", maxWork)
}

func TestBackfillAndSaveNewDay(t *testing.T) {
	path := writeTempJournal(t, []journal.JournalEntry{})
	setBackfillViper(t, path, "8h", "1h", "10h")

	args := []string{"20240527", "start:09:00", "end:18:30", "break:12:00-13:00:lunch", "note:Reviewed PRs #done"}
	entry, validationErr, err := backfillAndSave(args)
	if err != nil {
		t.Fatalf("backfillAndSave returned error: %v", err)
	}
	if validationErr != nil {
		t.Fatalf("expected no policy validation error, got: %v", validationErr)
	}
	if entry == nil {
		t.Fatal("expected non-nil entry")
	}

	entries, err := journal.LoadEntries(path)
	if err != nil {
		t.Fatalf("LoadEntries failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry persisted, got %d", len(entries))
	}
	got := entries[0]
	if got.ID != "20240527" {
		t.Errorf("entry ID = %q, want %q", got.ID, "20240527")
	}
	if got.StartTime.Format("2006-01-02 15:04") != "2024-05-27 09:00" {
		t.Errorf("StartTime = %v, want 2024-05-27 09:00", got.StartTime)
	}
	if got.EndTime.Format("2006-01-02 15:04") != "2024-05-27 18:30" {
		t.Errorf("EndTime = %v, want 2024-05-27 18:30", got.EndTime)
	}
	if len(got.Breaks) != 1 {
		t.Fatalf("expected 1 break, got %d", len(got.Breaks))
	}
	if got.Breaks[0].Reason != "lunch" {
		t.Errorf("break reason = %q, want %q", got.Breaks[0].Reason, "lunch")
	}
	if got.Breaks[0].StartTime.Format("2006-01-02 15:04") != "2024-05-27 12:00" {
		t.Errorf("break start = %v, want 2024-05-27 12:00", got.Breaks[0].StartTime)
	}
	if len(got.Notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(got.Notes))
	}
	if got.Notes[0].Contents != "Reviewed PRs" {
		t.Errorf("note contents = %q, want %q (hashtag should be parsed out)", got.Notes[0].Contents, "Reviewed PRs")
	}
	if len(got.Notes[0].Tags) != 1 || got.Notes[0].Tags[0] != "done" {
		t.Errorf("note tags = %v, want [done]", got.Notes[0].Tags)
	}
}

func TestBackfillAndSaveExistingDayRefused(t *testing.T) {
	existing := journal.JournalEntry{
		ID:        "20240527",
		StartTime: time.Date(2024, 5, 27, 9, 0, 0, 0, time.Local),
		EndTime:   time.Date(2024, 5, 27, 17, 0, 0, 0, time.Local),
	}
	path := writeTempJournal(t, []journal.JournalEntry{existing})
	setBackfillViper(t, path, "8h", "1h", "10h")

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"20240527", "start:08:00", "end:18:00"}
	_, _, err = backfillAndSave(args)
	if err == nil {
		t.Fatal("expected refusal error for existing day, got nil")
	}
	if !strings.Contains(err.Error(), "entry for 20240527 already exists") {
		t.Errorf("error = %q, want collision message", err.Error())
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Errorf("journal file was modified on collision; before=%q after=%q", before, after)
	}
}

func TestBackfillAndSavePolicyViolation(t *testing.T) {
	path := writeTempJournal(t, []journal.JournalEntry{})
	// minWorkTime 8h, but the backfilled day is only ~2h with no lunch break:
	// policy validation must fail and a warning note must be appended.
	setBackfillViper(t, path, "8h", "1h", "10h")

	args := []string{"20240527", "start:09:00", "end:11:00"}
	entry, validationErr, err := backfillAndSave(args)
	if err != nil {
		t.Fatalf("backfillAndSave returned hard error: %v", err)
	}
	if validationErr == nil {
		t.Fatal("expected a policy validation error, got nil")
	}
	if entry == nil {
		t.Fatal("expected entry to be saved despite policy violation")
	}

	entries, err := journal.LoadEntries(path)
	if err != nil {
		t.Fatalf("LoadEntries failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry persisted, got %d", len(entries))
	}
	foundWarning := false
	for _, n := range entries[0].Notes {
		if strings.HasPrefix(n.Contents, "Validation Error:") {
			foundWarning = true
		}
	}
	if !foundWarning {
		t.Errorf("expected a 'Validation Error:' note appended, notes = %+v", entries[0].Notes)
	}
}

// Regression for the UTC-vs-local anchoring bug: backfill must anchor the
// entry and its breaks to the local zone, matching how real-time commands
// store times. We force a non-UTC local zone so the assertion is meaningful
// regardless of host TZ. Pre-fix, time.Parse anchored to UTC, leaving the
// persisted offset at +0000 instead of the local offset.
func TestBackfillAndSaveAnchorsLocalZone(t *testing.T) {
	restore := time.Local
	time.Local = time.FixedZone("UTC-3", -3*60*60)
	t.Cleanup(func() { time.Local = restore })

	path := writeTempJournal(t, []journal.JournalEntry{})
	setBackfillViper(t, path, "8h", "1h", "10h")

	args := []string{"20240527", "start:09:00", "end:18:30", "break:12:00-13:00:lunch"}
	if _, _, err := backfillAndSave(args); err != nil {
		t.Fatalf("backfillAndSave returned error: %v", err)
	}

	entries, err := journal.LoadEntries(path)
	if err != nil {
		t.Fatalf("LoadEntries failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry persisted, got %d", len(entries))
	}
	got := entries[0]

	if _, off := got.StartTime.Zone(); off != -3*60*60 {
		t.Errorf("StartTime offset = %d, want -10800 (local)", off)
	}
	if got.StartTime.Hour() != 9 || got.EndTime.Hour() != 18 || got.EndTime.Minute() != 30 {
		t.Errorf("entry times = %v-%v, want 09:00-18:30 local", got.StartTime, got.EndTime)
	}
	if len(got.Breaks) != 1 {
		t.Fatalf("expected 1 break, got %d", len(got.Breaks))
	}
	if _, off := got.Breaks[0].StartTime.Zone(); off != -3*60*60 {
		t.Errorf("break StartTime offset = %d, want -10800 (local)", off)
	}
	if got.Breaks[0].StartTime.Hour() != 12 || got.Breaks[0].EndTime.Hour() != 13 {
		t.Errorf("break times = %v-%v, want 12:00-13:00 local", got.Breaks[0].StartTime, got.Breaks[0].EndTime)
	}
}
