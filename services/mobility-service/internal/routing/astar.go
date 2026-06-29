package routing

import (
	"container/heap"
	"fmt"
	"math"
)

type Graph struct {
	Nodes map[string]*Node
	Edges map[string][]*Edge
}

type Node struct {
	ID  string
	Lat float64
	Lng float64
	H3  string
}

type Edge struct {
	To        *Node
	WeightS   float64
	DistanceM int
}

type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Route struct {
	Nodes     []*Node  `json:"-"`
	Waypoints []LatLng `json:"waypoints"`
	DistanceM int      `json:"distance_m"`
	ETAS      int      `json:"eta_s"`
}

func (g *Graph) AStar(from, to *Node) (*Route, error) {
	if g == nil || from == nil || to == nil {
		return nil, fmt.Errorf("graph and nodes are required")
	}
	if from.ID == to.ID {
		return &Route{
			Nodes:     []*Node{from},
			Waypoints: []LatLng{{Lat: from.Lat, Lng: from.Lng}},
		}, nil
	}

	openSet := &priorityQueue{}
	heap.Init(openSet)
	heap.Push(openSet, &queueItem{id: from.ID, score: 0})

	cameFrom := map[string]string{}
	gScore := map[string]float64{from.ID: 0}
	fScore := map[string]float64{from.ID: heuristic(from, to)}

	for openSet.Len() > 0 {
		current := heap.Pop(openSet).(*queueItem).id
		if current == to.ID {
			return g.reconstructRoute(cameFrom, current)
		}

		for _, edge := range g.Edges[current] {
			tentative := gScore[current] + edge.WeightS
			if prev, ok := gScore[edge.To.ID]; ok && tentative >= prev {
				continue
			}
			cameFrom[edge.To.ID] = current
			gScore[edge.To.ID] = tentative
			fScore[edge.To.ID] = tentative + heuristic(edge.To, to)
			heap.Push(openSet, &queueItem{id: edge.To.ID, score: fScore[edge.To.ID]})
		}
	}

	return nil, fmt.Errorf("route not found")
}

func (g *Graph) reconstructRoute(cameFrom map[string]string, current string) (*Route, error) {
	path := []string{current}
	for {
		prev, ok := cameFrom[current]
		if !ok {
			break
		}
		path = append([]string{prev}, path...)
		current = prev
	}

	nodes := make([]*Node, 0, len(path))
	waypoints := make([]LatLng, 0, len(path))
	distance := 0
	eta := 0.0

	for idx, id := range path {
		node, ok := g.Nodes[id]
		if !ok {
			return nil, fmt.Errorf("node %s missing", id)
		}
		nodes = append(nodes, node)
		waypoints = append(waypoints, LatLng{Lat: node.Lat, Lng: node.Lng})

		if idx == 0 {
			continue
		}
		prevID := path[idx-1]
		for _, edge := range g.Edges[prevID] {
			if edge.To.ID == id {
				distance += edge.DistanceM
				eta += edge.WeightS
				break
			}
		}
	}

	return &Route{
		Nodes:     nodes,
		Waypoints: waypoints,
		DistanceM: distance,
		ETAS:      int(math.Round(eta)),
	}, nil
}

func heuristic(current, goal *Node) float64 {
	const assumedSpeedMS = 10.0
	return haversineMeters(current.Lat, current.Lng, goal.Lat, goal.Lng) / assumedSpeedMS
}

func haversineMeters(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusM = 6371000
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusM * c
}

type queueItem struct {
	id    string
	score float64
	index int
}

type priorityQueue []*queueItem

func (pq priorityQueue) Len() int           { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].score < pq[j].score }
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}
func (pq *priorityQueue) Push(x any) {
	item := x.(*queueItem)
	item.index = len(*pq)
	*pq = append(*pq, item)
}
func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[:n-1]
	return item
}
