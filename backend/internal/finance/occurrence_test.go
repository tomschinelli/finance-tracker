package finance

import (
	"testing"
	"time"
)

func date(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func dateStrings(ts []time.Time) []string {
	out := make([]string, len(ts))
	for i, t := range ts {
		out[i] = t.Format("2006-01-02")
	}
	return out
}

func TestComputeOccurrenceDates_Monthly(t *testing.T) {
	got := dateStrings(ComputeOccurrenceDates(date("2026-01-15"), 1, date("2026-01-01"), date("2026-03-31")))
	want := []string{"2026-01-15", "2026-02-15", "2026-03-15"}
	if !equalStrings(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestComputeOccurrenceDates_CustomInterval(t *testing.T) {
	// Every 2 months.
	got := dateStrings(ComputeOccurrenceDates(date("2026-01-15"), 2, date("2026-01-01"), date("2026-06-30")))
	want := []string{"2026-01-15", "2026-03-15", "2026-05-15"}
	if !equalStrings(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestComputeOccurrenceDates_MonthEndClamped(t *testing.T) {
	// Jan 31 + 1 month has no Feb 31, so it clamps to Feb 28 (2026 is not a leap year).
	got := dateStrings(ComputeOccurrenceDates(date("2026-01-31"), 1, date("2026-01-01"), date("2026-03-31")))
	want := []string{"2026-01-31", "2026-02-28", "2026-03-31"}
	if !equalStrings(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestComputeOccurrenceDates_YearlyAnchorBeforeRange(t *testing.T) {
	got := dateStrings(ComputeOccurrenceDates(date("2020-06-01"), 12, date("2026-01-01"), date("2026-12-31")))
	want := []string{"2026-06-01"}
	if !equalStrings(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestComputeOccurrenceDates_NonPositiveIntervalReturnsNil(t *testing.T) {
	got := ComputeOccurrenceDates(date("2026-01-01"), 0, date("2026-01-01"), date("2026-12-31"))
	if got != nil {
		t.Fatalf("got %v, want nil for a non-positive interval", got)
	}
}

func TestComputeOccurrenceDates_RangeEntirelyBeforeAnchor(t *testing.T) {
	got := ComputeOccurrenceDates(date("2026-06-01"), 1, date("2026-01-01"), date("2026-03-31"))
	if len(got) != 0 {
		t.Fatalf("got %v, want none (anchor is after the whole range)", got)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
