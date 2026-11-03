package finance

// MonthlySetAsideCents estimates the amount that should be set aside each
// month to cover active recurring expenses of every frequency. Integer
// division is intentional: this is an estimate, not an accounting figure.
func MonthlySetAsideCents(monthlyExpenseCents, quarterlyExpenseCents, yearlyExpenseCents int64) int64 {
	return monthlyExpenseCents + quarterlyExpenseCents/3 + yearlyExpenseCents/12
}
