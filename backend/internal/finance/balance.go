package finance

// AccountBalanceCents derives an account's current balance. It is never
// stored as a mutable counter; it is always recomputed from the ledger.
func AccountBalanceCents(startingBalanceCents, paidPaymentsSumCents, adjustmentsSumCents int64) int64 {
	return startingBalanceCents + paidPaymentsSumCents + adjustmentsSumCents
}
