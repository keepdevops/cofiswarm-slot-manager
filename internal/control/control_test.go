package control

import "testing"

func TestTickEvictsOnlyWhenPolicyFires(t *testing.T) {
	readings := []Reading{
		{EndpointID: "a", Usage: 0.50}, // nominal -> no action
		{EndpointID: "b", Usage: 0.65}, // proactive evict
		{EndpointID: "c", Usage: 0.90}, // full clear
	}
	evicted := map[string]int{}
	d := Deps{
		Snapshot: func() []Reading { return readings },
		Evaluate: func(usage float64) (bool, bool) {
			switch {
			case usage >= 0.75:
				return true, false // clear
			case usage >= 0.60:
				return false, true // evict
			default:
				return false, false
			}
		},
		Evict: func(r Reading) int { evicted[r.EndpointID]++; return 1 },
	}
	if fired := Tick(d); fired != 2 {
		t.Fatalf("fired = %d, want 2", fired)
	}
	if evicted["a"] != 0 || evicted["b"] != 1 || evicted["c"] != 1 {
		t.Fatalf("evicted = %v, want {b:1,c:1}", evicted)
	}
}

func TestTickNoActionWhenAllNominal(t *testing.T) {
	d := Deps{
		Snapshot: func() []Reading { return []Reading{{EndpointID: "a", Usage: 0.1}} },
		Evaluate: func(float64) (bool, bool) { return false, false },
		Evict:    func(Reading) int { t.Fatal("evict must not be called"); return 0 },
	}
	if fired := Tick(d); fired != 0 {
		t.Fatalf("fired = %d, want 0", fired)
	}
}
