package main

import (
	"container/heap"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	gifpkg "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

var (
	allowedDiagonals bool
	includeSteps     bool
	heuristics       = HeuristicFunc(manhattan)
)

const (
	BlackIndex uint8 = iota
	WhiteIndex
	RedIndex
	BlueIndex
)

var palette = color.Palette{
	color.RGBA{0, 0, 0, 255},
	color.RGBA{255, 255, 255, 255},
	color.RGBA{255, 0, 0, 255},
	color.RGBA{0, 0, 255, 255},
}

var (
	in         image.Image
	gif        gifpkg.GIF
	MaxX, MaxY int
)

var (
	adjsBuff   [8]image.Point
	orig, dest image.Point
	global     = map[image.Point]int{}
	cameFrom   = map[image.Point]image.Point{}
	passed     = map[image.Point]struct{}{}
)

func exitOnError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(-1)
}

func wallAt(x, y int) bool {
	r, g, b, _ := in.At(x, y).RGBA()
	return r == 0 && g == 0 && b == 0
}

func adjacents(p image.Point) []image.Point {
	adjs := adjsBuff[:0]
	px, py := p.X, p.Y
	neighbours := []image.Point{
		{px - 1, py},
		{px + 1, py},
		{px, py - 1},
		{px, py + 1},
		{px - 1, py - 1},
		{px + 1, py - 1},
		{px + 1, py + 1},
		{px - 1, py + 1},
	}
	if !allowedDiagonals {
		neighbours = neighbours[:4]
	}

	for _, n := range neighbours {
		if n.X < 0 || n.X == MaxX {
			continue
		}
		if n.Y < 0 || n.Y == MaxY {
			continue
		}
		if wallAt(n.X, n.Y) {
			continue
		}
		adjs = append(adjs, n)
	}
	return adjs
}

func writeFrame() {
	frame := image.NewPaletted(image.Rect(0, 0, MaxX, MaxY), palette)
	for y := 0; y < MaxY; y++ {
		for x := 0; x < MaxX; x++ {
			ind := BlackIndex
			if !wallAt(x, y) {
				ind = WhiteIndex
			}
			_, ok := passed[image.Pt(x, y)]
			if ok {
				ind = BlueIndex
			}
			frame.SetColorIndex(x, y, ind)
		}
	}
	gif.Image = append(gif.Image, frame)
	gif.Delay = append(gif.Delay, 10)
}

func writePathFrame() {
	var p image.Point
	frame := image.NewPaletted(image.Rect(0, 0, MaxX, MaxY), palette)
	for y := 0; y < MaxY; y++ {
		for x := 0; x < MaxX; x++ {
			ind := BlackIndex
			if !wallAt(x, y) {
				ind = WhiteIndex
			}
			frame.SetColorIndex(x, y, ind)
		}
	}
	for p = dest; p != orig; p = cameFrom[p] {
		frame.SetColorIndex(p.X, p.Y, RedIndex)
	}
	frame.SetColorIndex(p.X, p.Y, RedIndex)
	gif.Image = append(gif.Image, frame)
	gif.Delay = append(gif.Delay, 10)
}

func main() {
	flag.BoolVar(&allowedDiagonals, "d", false, "Allow diagonals")
	flag.BoolVar(&includeSteps, "s", false, "Write search steps as gif frames")
	flag.Var(&heuristics, "h", "Heuristics function to use [manhattan|euclidian]")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		exitOnError(errors.New("filename not provided"))
	}

	file, err := os.Open(args[0])
	if err != nil {
		exitOnError(err)
	}
	in, _, err = image.Decode(file)
	if err != nil {
		exitOnError(err)
	}
	err = file.Close()
	if err != nil {
		exitOnError(err)
	}

	bounds := in.Bounds().Max
	MaxX, MaxY = bounds.X, bounds.Y
	for x := 0; x < MaxX; x++ {
		if !wallAt(x, 0) {
			orig = image.Pt(x, 0)
			break
		}
	}
	for x := 0; x < MaxX; x++ {
		if !wallAt(x, bounds.Y-1) {
			dest = image.Pt(x, bounds.Y-1)
			break
		}
	}

	gif = gifpkg.GIF{LoopCount: -1}

	var p image.Point
	pq := PriorityQueue{}
	pq.Push(orig)
	global[orig] = 0
	for pq.Len() != 0 {
		p = pq.Pop().(image.Point)
		passed[p] = struct{}{}
		if p == dest {
			break
		}
		gp := global[p]
		for _, adj := range adjacents(p) {
			g, ok := global[adj]
			if ok && g <= (gp+1) {
				continue
			}
			cameFrom[adj] = p
			global[adj] = gp + 1
			pq.Push(adj)
		}
		if includeSteps {
			writeFrame()
		}
		heap.Init(&pq)
	}
	writePathFrame()

	file, err = os.Create("out.gif")
	if err != nil {
		exitOnError(err)
	}
	err = gifpkg.EncodeAll(file, &gif)
	if err != nil {
		exitOnError(err)
	}
	err = file.Close()
	if err != nil {
		exitOnError(err)
	}
}
