package routing

import "testing"

func TestGraphAStar(t *testing.T) {
	a := &Node{ID: "a", Lat: 6.5, Lng: 3.3}
	b := &Node{ID: "b", Lat: 6.51, Lng: 3.31}
	c := &Node{ID: "c", Lat: 6.52, Lng: 3.32}

	graph := &Graph{
		Nodes: map[string]*Node{"a": a, "b": b, "c": c},
		Edges: map[string][]*Edge{
			"a": {{To: b, WeightS: 10, DistanceM: 1000}},
			"b": {{To: c, WeightS: 15, DistanceM: 1500}},
			"c": {},
		},
	}

	route, err := graph.AStar(a, c)
	if err != nil {
		t.Fatalf("AStar() error = %v", err)
	}
	if route.DistanceM != 2500 {
		t.Fatalf("expected route distance 2500, got %d", route.DistanceM)
	}
	if len(route.Waypoints) != 3 {
		t.Fatalf("expected 3 waypoints, got %d", len(route.Waypoints))
	}
}
