// Package finance holds domain logic (occurrence generation, balance and
// budget calculation) independent of HTTP and SQL concerns.
package finance

import "time"

// ComputeOccurrenceDates returns every due date in [from, to] (inclusive) for
// a recurring definition with the given anchor date and frequency.
func ComputeOccurrenceDates(anchor time.Time, frequency string, from, to time.Time) []time.Time {
	step := monthsPerOccurrence(frequency)
	if step == 0 {
		return nil
	}
	anchor = truncateToDate(anchor)
	from = truncateToDate(from)
	to = truncateToDate(to)

	var dates []time.Time
	for k := 0; ; k++ {
		d := addMonthsClamped(anchor, step*k)
		if d.After(to) {
			break
		}
		if !d.Before(from) {
			dates = append(dates, d)
		}
	}
	return dates
}

func monthsPerOccurrence(frequency string) int {
	switch frequency {
	case "monthly":
		return 1
	case "quarterly":
		return 3
	case "yearly":
		return 12
	default:
		return 0
	}
}

func truncateToDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// addMonthsClamped adds the given number of months to t, clamping the
// day-of-month to the last day of the resulting month when it overflows
// (e.g. Jan 31 + 1 month -> Feb 28/29).
func addMonthsClamped(t time.Time, months int) time.Time {
	year, month, day := t.Date()
	total := int(month) - 1 + months
	y := year + total/12
	m := total % 12
	if m < 0 {
		m += 12
		y--
	}
	firstOfMonth := time.Date(y, time.Month(m+1), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstOfMonth.AddDate(0, 1, -1).Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(y, time.Month(m+1), day, 0, 0, 0, 0, time.UTC)
}
