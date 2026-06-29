package tracking

import (
	"math"
	"testing"
)

func TestKalmanFilterSmoothsTowardSignal(t *testing.T) {
	filter := NewKalmanFilter(6.5000, 3.3000)
	samples := [][2]float64{
		{6.5005, 3.3004},
		{6.4998, 3.3002},
		{6.5003, 3.3001},
	}

	lat, lng := 0.0, 0.0
	for _, sample := range samples {
		lat, lng = filter.Update(sample[0], sample[1], 1)
	}

	if math.Abs(lat-6.5003) > 0.001 {
		t.Fatalf("expected smoothed lat near signal, got %f", lat)
	}
	if math.Abs(lng-3.3001) > 0.001 {
		t.Fatalf("expected smoothed lng near signal, got %f", lng)
	}
}
