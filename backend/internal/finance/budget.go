package finance

// MonthlyBufferCents estimates the amount that should be set aside each
// month to smooth irregular (non-monthly) recurring expenses into a steady
// monthly transfer, independent of which months they actually fall due in.
// expenseCentsByIntervalMonths sums active expense amounts keyed by their
// recurring interval in months; entries for interval 1 (monthly) are
// ignored since those aren't smoothed. Integer division is intentional:
// this is an estimate, not an accounting figure.
func MonthlyBufferCents(expenseCentsByIntervalMonths map[int]int64) int64 {
	var buffer int64
	for interval, total := range expenseCentsByIntervalMonths {
		if interval <= 1 {
			continue
		}
		buffer += total / int64(interval)
	}
	return buffer
}
