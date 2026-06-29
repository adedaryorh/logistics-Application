package h3

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func LatLngToCell(lat, lng float64, resolution int) string {
	scale := math.Pow10(max(resolution-4, 0))
	latBucket := int(math.Round(lat * scale))
	lngBucket := int(math.Round(lng * scale))
	return fmt.Sprintf("r%d:%d:%d", resolution, latBucket, lngBucket)
}

func CellToLatLng(cell string) (lat, lng float64) {
	resolution, latBucket, lngBucket, ok := parseCell(cell)
	if !ok {
		return 0, 0
	}
	scale := math.Pow10(max(resolution-4, 0))
	return float64(latBucket) / scale, float64(lngBucket) / scale
}

func KRingCells(cell string, k int) []string {
	resolution, latBucket, lngBucket, ok := parseCell(cell)
	if !ok {
		return nil
	}
	cells := make([]string, 0, (2*k+1)*(2*k+1))
	for dLat := -k; dLat <= k; dLat++ {
		for dLng := -k; dLng <= k; dLng++ {
			cells = append(cells, fmt.Sprintf("r%d:%d:%d", resolution, latBucket+dLat, lngBucket+dLng))
		}
	}
	return cells
}

func IsInGeofenceCells(cell string, geofenceCells []string) bool {
	for _, geofenceCell := range geofenceCells {
		if geofenceCell == cell {
			return true
		}
	}
	return false
}

func DistanceBetweenCells(a, b string) int {
	_, aLat, aLng, okA := parseCell(a)
	_, bLat, bLng, okB := parseCell(b)
	if !okA || !okB {
		return 0
	}
	return abs(aLat-bLat) + abs(aLng-bLng)
}

func parseCell(cell string) (resolution, latBucket, lngBucket int, ok bool) {
	parts := strings.Split(cell, ":")
	if len(parts) != 3 || !strings.HasPrefix(parts[0], "r") {
		return 0, 0, 0, false
	}
	resolution, err := strconv.Atoi(strings.TrimPrefix(parts[0], "r"))
	if err != nil {
		return 0, 0, 0, false
	}
	latBucket, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, false
	}
	lngBucket, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, false
	}
	return resolution, latBucket, lngBucket, true
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
