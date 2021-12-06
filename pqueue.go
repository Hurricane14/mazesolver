package main

import (
	"bytes"
	"fmt"
	"image"
)

type PriorityQueue []image.Point

func (pq *PriorityQueue) Len() int {
	return len(*pq)
}

func (pq *PriorityQueue) Less(i, j int) bool {
	lp, rp := (*pq)[i], (*pq)[j]
	fi := float64(global[lp]) + heuristics(lp)
	fj := float64(global[rp]) + heuristics(rp)
	return fi < fj
}

func (pq *PriorityQueue) Swap(i, j int) {
	a := *pq
	a[i], a[j] = a[j], a[i]
}

func (pq *PriorityQueue) Push(ip interface{}) {
	p := ip.(image.Point)
	*pq = append(*pq, p)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	val := old[0]
	*pq = old[1:]
	return val
}

func (pq *PriorityQueue) String() string {
	var buffer bytes.Buffer
	buffer.WriteRune('[')
	for _, p := range *pq {
		dist := float64(global[p]) + heuristics(p)
		buffer.WriteString(fmt.Sprintf("\t%v: %f\n", p, dist))
	}
	buffer.WriteRune(']')
	return buffer.String()
}
