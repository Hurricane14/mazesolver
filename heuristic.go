package main

import (
	"errors"
	"math"
	"strings"
)

type HeuristicFunc func(Point) float64

func euclidian(p Point) float64 {
	dx := float64(dest.X - p.X)
	dy := float64(dest.Y - p.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

func manhattan(p Point) float64 {
	dx := math.Abs(float64(dest.X - p.X))
	dy := math.Abs(float64(dest.Y - p.Y))
	return dx + dy
}

func (h *HeuristicFunc) String() string {
	return ""
}

func (hf *HeuristicFunc) Set(s string) error {
	s = strings.ToLower(s)
	if strings.HasPrefix("manhattan", s) {
		*hf = manhattan
	} else if strings.HasPrefix("euclidian", s) {
		*hf = euclidian
	} else {
		return errors.New("couldn't parse heuristic function")
	}
	return nil
}
