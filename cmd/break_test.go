package cmd

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/deadpyxel/workday/internal/journal"
)

// bootstrapBreakJournal writes the given entries to a temp file and returns its path.
func bootstrapBreakJournal(t *testing.T, entries []journal.JournalEntry) string {
	t.Helper()
	f, err := os.CreateTemp("", "journal_break_")
	if err != nil {
		t.Fatal(err)
	}
	name := f.Name()
	f.Close()
	t.Cleanup(func() { os.Remove(name) })

	if err := journal.SaveEntries(entries, name); err != nil {
		t.Fatalf("failed to bootstrap journal file: %v", err)
	}
	return name
}

// loadBreakJournal reloads entries from filename for assertions.
func loadBreakJournal(t *testing.T, filename string) []journal.JournalEntry {
	t.Helper()
	entries, err := journal.LoadEntries(filename)
	if err != nil {
		t.Fatalf("failed to reload journal file: %v", err)
	}
	return entries
}

func TestAddBreakToJournal(t *testing.T) {
	// Fixed "now" for the unset-date path: 2026-06-03 10:00 local.
	now := time.Date(2026, 6, 3, 10, 0, 0, 0, time.Local)
	todayID := now.Format("20060102")

	t.Run("date set, entry exists: break appended to that day anchored to that date", func(t *testing.T) {
		targetID := "20240527"
		targetStart := time.Date(2024, 5, 27, 9, 0, 0, 0, time.Local)
		entries := []journal.JournalEntry{
			{ID: targetID, StartTime: targetStart},
		}
		journalPath := bootstrapBreakJournal(t, entries)

		entry, idx, err := addBreakToJournal(journalPath, "2024-05-27", now,
			[]string{"start:12:00", "end:13:00", "reason:lunch"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if idx != 0 {
			t.Errorf("expected entry index 0, got %d", idx)
		}
		if len(entry.Breaks) != 1 {
			t.Fatalf("expected 1 break, got %d", len(entry.Breaks))
		}

		br := entry.Breaks[0]
		// Assert the DATE components anchor to the target day (2024-05-27), not today.
		if br.StartTime.Year() != 2024 || br.StartTime.Month() != time.May || br.StartTime.Day() != 27 {
			t.Errorf("StartTime date = %v, want 2024-05-27", br.StartTime)
		}
		if br.EndTime.Year() != 2024 || br.EndTime.Month() != time.May || br.EndTime.Day() != 27 {
			t.Errorf("EndTime date = %v, want 2024-05-27", br.EndTime)
		}
		// Assert the time-of-day components.
		if br.StartTime.Hour() != 12 || br.StartTime.Minute() != 0 {
			t.Errorf("StartTime time = %v, want 12:00", br.StartTime)
		}
		if br.EndTime.Hour() != 13 || br.EndTime.Minute() != 0 {
			t.Errorf("EndTime time = %v, want 13:00", br.EndTime)
		}
		if br.Reason != "lunch" {
			t.Errorf("Reason = %q, want %q", br.Reason, "lunch")
		}

		// Persisted state must match: reload and verify the break landed on the target day.
		reloaded := loadBreakJournal(t, journalPath)
		target, ridx := journal.FetchEntryByID(targetID, reloaded)
		if ridx == -1 {
			t.Fatalf("target entry %s not found after save", targetID)
		}
		if len(target.Breaks) != 1 {
			t.Fatalf("expected 1 persisted break, got %d", len(target.Breaks))
		}
		if target.Breaks[0].StartTime.Day() != 27 || target.Breaks[0].StartTime.Month() != time.May {
			t.Errorf("persisted break date = %v, want 2024-05-27", target.Breaks[0].StartTime)
		}
	})

	t.Run("date set, entry missing: refused with backfill message", func(t *testing.T) {
		entries := []journal.JournalEntry{
			{ID: todayID, StartTime: now},
		}
		journalPath := bootstrapBreakJournal(t, entries)

		_, _, err := addBreakToJournal(journalPath, "2024-05-27", now,
			[]string{"start:12:00", "end:13:00", "reason:lunch"})
		if err == nil {
			t.Fatalf("expected error for missing entry, got nil")
		}
		want := "no entry found for 2024-05-27; use 'workday backfill' to create it"
		if err.Error() != want {
			t.Errorf("error = %q, want %q", err.Error(), want)
		}
	})

	t.Run("date set, invalid format: refused with format message", func(t *testing.T) {
		entries := []journal.JournalEntry{
			{ID: todayID, StartTime: now},
		}
		journalPath := bootstrapBreakJournal(t, entries)

		_, _, err := addBreakToJournal(journalPath, "2024/05/27", now,
			[]string{"start:12:00", "end:13:00", "reason:lunch"})
		if err == nil {
			t.Fatalf("expected error for invalid date format, got nil")
		}
		if !strings.Contains(err.Error(), "invalid date format. Use YYYY-MM-DD") {
			t.Errorf("error = %q, want it to contain %q", err.Error(), "invalid date format. Use YYYY-MM-DD")
		}
	})

	t.Run("date unset: targets today as before", func(t *testing.T) {
		entries := []journal.JournalEntry{
			{ID: todayID, StartTime: now},
		}
		journalPath := bootstrapBreakJournal(t, entries)

		entry, _, err := addBreakToJournal(journalPath, "", now,
			[]string{"start:12:00", "end:13:00", "reason:lunch"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entry.Breaks) != 1 {
			t.Fatalf("expected 1 break, got %d", len(entry.Breaks))
		}
		br := entry.Breaks[0]
		// Anchored to today's date (now), not the zero-value 0001 year.
		if br.StartTime.Year() != now.Year() || br.StartTime.Month() != now.Month() || br.StartTime.Day() != now.Day() {
			t.Errorf("StartTime date = %v, want %v", br.StartTime, now.Format("2006-01-02"))
		}
		if br.StartTime.Hour() != 12 || br.EndTime.Hour() != 13 {
			t.Errorf("break times = %v-%v, want 12:00-13:00", br.StartTime, br.EndTime)
		}

		// Verify persisted to today's entry.
		reloaded := loadBreakJournal(t, journalPath)
		today, ridx := journal.FetchEntryByID(todayID, reloaded)
		if ridx == -1 {
			t.Fatalf("today entry %s not found after save", todayID)
		}
		if len(today.Breaks) != 1 {
			t.Fatalf("expected 1 persisted break, got %d", len(today.Breaks))
		}
	})

	t.Run("date unset, today entry missing: refused", func(t *testing.T) {
		journalPath := bootstrapBreakJournal(t, []journal.JournalEntry{})

		_, _, err := addBreakToJournal(journalPath, "", now,
			[]string{"start:12:00", "end:13:00", "reason:lunch"})
		if err == nil {
			t.Fatalf("expected error for missing today entry, got nil")
		}
	})

	t.Run("overlap rejected on target day", func(t *testing.T) {
		targetID := "20240527"
		targetStart := time.Date(2024, 5, 27, 9, 0, 0, 0, time.Local)
		existing := journal.Break{
			StartTime: time.Date(2024, 5, 27, 12, 0, 0, 0, time.Local),
			EndTime:   time.Date(2024, 5, 27, 13, 0, 0, 0, time.Local),
			Reason:    "lunch",
		}
		entries := []journal.JournalEntry{
			{ID: targetID, StartTime: targetStart, Breaks: []journal.Break{existing}},
		}
		journalPath := bootstrapBreakJournal(t, entries)

		_, _, err := addBreakToJournal(journalPath, "2024-05-27", now,
			[]string{"start:12:30", "end:13:30", "reason:coffee"})
		if err == nil {
			t.Fatalf("expected overlap error, got nil")
		}

		// No write on validation failure: reload and confirm still 1 break.
		reloaded := loadBreakJournal(t, journalPath)
		target, _ := journal.FetchEntryByID(targetID, reloaded)
		if len(target.Breaks) != 1 {
			t.Errorf("expected break count unchanged at 1, got %d", len(target.Breaks))
		}
	})

	// Regression for the UTC-vs-local anchoring bug: a break added via --date
	// must be anchored in the same timezone as locally-stored breaks. We force
	// a non-UTC local zone so the bug is reproducible regardless of host TZ.
	// Pre-fix, time.Parse anchored the new break to UTC, shifting it by the
	// local offset and falsely overlapping the existing local break.
	t.Run("non-overlapping break accepted across non-UTC local zone", func(t *testing.T) {
		restore := time.Local
		time.Local = time.FixedZone("UTC-3", -3*60*60)
		t.Cleanup(func() { time.Local = restore })

		targetID := "20260623"
		// Existing break stored in local time, as break start would write it.
		existing := journal.Break{
			StartTime: time.Date(2026, 6, 23, 9, 21, 0, 0, time.Local),
			EndTime:   time.Date(2026, 6, 23, 9, 58, 0, 0, time.Local),
			Reason:    "morning",
		}
		entries := []journal.JournalEntry{
			{ID: targetID, StartTime: time.Date(2026, 6, 23, 9, 0, 0, 0, time.Local), Breaks: []journal.Break{existing}},
		}
		journalPath := bootstrapBreakJournal(t, entries)

		entry, _, err := addBreakToJournal(journalPath, "2026-06-23", now,
			[]string{"start:11:56", "end:13:00", "reason:lunch"})
		if err != nil {
			t.Fatalf("unexpected error adding non-overlapping break: %v", err)
		}
		if len(entry.Breaks) != 2 {
			t.Fatalf("expected 2 breaks after add, got %d", len(entry.Breaks))
		}

		// The new break must keep its local clock time (11:56-13:00 local),
		// not be shifted into another zone.
		added := entry.Breaks[1]
		if added.StartTime.Hour() != 11 || added.StartTime.Minute() != 56 {
			t.Errorf("StartTime = %v, want 11:56 local", added.StartTime)
		}
		if added.EndTime.Hour() != 13 || added.EndTime.Minute() != 0 {
			t.Errorf("EndTime = %v, want 13:00 local", added.EndTime)
		}
		if _, off := added.StartTime.Zone(); off != -3*60*60 {
			t.Errorf("StartTime offset = %d, want -10800 (local)", off)
		}
	})
}
