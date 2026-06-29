package dispatch

import "testing"

func TestScoreDriver(t *testing.T) {
	score := ScoreDriver(DriverCandidate{
		DriverID:   "driver-1",
		DistanceM:  1000,
		Rating:     4.8,
		TotalTrips: 250,
	}, 5000)

	if score <= 0 {
		t.Fatalf("expected positive score, got %f", score)
	}
	if score > 100 {
		t.Fatalf("expected score <= 100, got %f", score)
	}
}

func TestRankDrivers(t *testing.T) {
	ranked := RankDrivers([]DriverCandidate{
		{DriverID: "far", DistanceM: 4000, Rating: 4.1, TotalTrips: 400},
		{DriverID: "best", DistanceM: 500, Rating: 4.9, TotalTrips: 900},
		{DriverID: "mid", DistanceM: 900, Rating: 4.6, TotalTrips: 300},
	}, 5000)

	if len(ranked) != 3 {
		t.Fatalf("expected 3 ranked drivers, got %d", len(ranked))
	}
	if ranked[0].DriverID != "best" {
		t.Fatalf("expected best driver first, got %q", ranked[0].DriverID)
	}
}
