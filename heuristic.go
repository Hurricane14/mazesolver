package main

import (
	"math"
)

type HeuristicFunc func(Point, Point) float64

func euclidian(dest, p Point) float64 {
	dx := float64(dest.X - p.X)
	dy := float64(dest.Y - p.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

func manhattan(dest, p Point) float64 {
	dx := math.Abs(float64(dest.X - p.X))
	dy := math.Abs(float64(dest.Y - p.Y))
	return dx + dy
}
