package finance

import "testing"

func TestAccountBalanceCents(t *testing.T) {
	got := AccountBalanceCents(1000, -300, 50)
	if got != 750 {
		t.Fatalf("AccountBalanceCents(1000, -300, 50) = %d, want 750", got)
	}
}
