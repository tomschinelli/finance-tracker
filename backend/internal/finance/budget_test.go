package finance

import "testing"

func TestMonthlyBufferCents(t *testing.T) {
	cases := []struct {
		name    string
		entries map[int]int64
		want    int64
	}{
		{"empty", map[int]int64{}, 0},
		{"monthly entries are ignored", map[int]int64{1: 500000}, 0},
		{"quarterly-equivalent (interval 3)", map[int]int64{3: 9000}, 3000},
		{"yearly-equivalent (interval 12)", map[int]int64{12: 12000}, 1000},
		{"custom interval (every 2 months)", map[int]int64{2: 5000}, 2500},
		{"multiple intervals, integer division truncates", map[int]int64{3: 1000, 12: 1000}, 1000/3 + 1000/12},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := MonthlyBufferCents(c.entries); got != c.want {
				t.Fatalf("MonthlyBufferCents(%v) = %d, want %d", c.entries, got, c.want)
			}
		})
	}
}
