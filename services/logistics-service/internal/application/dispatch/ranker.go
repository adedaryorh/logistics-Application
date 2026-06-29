package dispatch

import (
	"math"
	"sort"
)

type DriverCandidate struct {
	DriverID   string
	DistanceM  float64
	Rating     float64
	TotalTrips int
	H3Cell     string
}

func ScoreDriver(d DriverCandidate, maxDistanceM float64) float64 {
	if maxDistanceM <= 0 {
		maxDistanceM = 1
	}

	ratingComponent := 0.4 * clamp(d.Rating/5, 0, 1) * 100
	proximityComponent := 0.4 * clamp(1-(d.DistanceM/maxDistanceM), 0, 1) * 100
	experienceComponent := 0.2 * clamp(float64(d.TotalTrips)/1000, 0, 1) * 100

	return ratingComponent + proximityComponent + experienceComponent
}

func RankDrivers(candidates []DriverCandidate, maxDistanceM float64) []DriverCandidate {
	ranked := make([]DriverCandidate, len(candidates))
	copy(ranked, candidates)

	sort.SliceStable(ranked, func(i, j int) bool {
		left := ScoreDriver(ranked[i], maxDistanceM)
		right := ScoreDriver(ranked[j], maxDistanceM)
		if left == right {
			return ranked[i].DistanceM < ranked[j].DistanceM
		}
		return left > right
	})

	return ranked
}

func clamp(value, minValue, maxValue float64) float64 {
	return math.Min(math.Max(value, minValue), maxValue)
}
