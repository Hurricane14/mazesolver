package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	gifpkg "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"

	"mazesolver/pqueue"
)

var (
	allowedDiagonals bool
	writeAsGIF       bool
	useDijkstra      bool
	heuristics       = HeuristicFunc(manhattan)
)

const (
	BlackIndex uint8 = iota
	WhiteIndex
	RedIndex
	BlueIndex
	GreenIndex
	OrangeIndex
)

var palette = color.Palette{
	color.RGBA{0, 0, 0, 255},
	color.RGBA{255, 255, 255, 255},
	color.RGBA{255, 0, 0, 255},
	color.RGBA{0, 0, 255, 255},
	color.RGBA{0, 255, 0, 255},
	color.RGBA{255, 215, 0, 255},
}

const frameDelay = 5

var (
	img        image.Image
	gif        gifpkg.GIF
	MaxX, MaxY int
)

type Point image.Point

var (
	adjsBuff   [8]Point
	orig, dest Point
	global     = map[Point]int{}
	cameFrom   = map[Point]Point{}
	passed     = map[Point]bool{}
)

func wallAt(x, y int) bool {
	r, g, b, _ := img.At(x, y).RGBA()
	return r == 0 && g == 0 && b == 0
}

func adjacents(p Point) []Point {
	adjs := adjsBuff[:0]
	px, py := p.X, p.Y
	neighbours := []Point{
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

func createFrame() *image.Paletted {
	frame := image.NewPaletted(image.Rect(0, 0, MaxX, MaxY), palette)
	for y := 0; y < MaxY; y++ {
		for x := 0; x < MaxX; x++ {
			ind := WhiteIndex
			if wallAt(x, y) {
				ind = BlackIndex
			}
			if passed[Point(image.Pt(x, y))] {
				ind = BlueIndex
			}
			frame.SetColorIndex(x, y, ind)
		}
	}
	frame.SetColorIndex(orig.X, orig.Y, OrangeIndex)
	frame.SetColorIndex(dest.X, dest.Y, GreenIndex)
	return frame
}

func appendFrame() {
	frame := createFrame()
	gif.Image = append(gif.Image, frame)
	gif.Delay = append(gif.Delay, frameDelay)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s *input-file*\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.BoolVar(&allowedDiagonals, "d", false, "Allow diagonals")
	flag.BoolVar(&writeAsGIF, "s", false, "Write search as GIF")
	flag.BoolVar(&useDijkstra, "D", false, "Use Dijkstra's algorithm instead of A*")
	flag.Var(&heuristics, "H", "Heuristic function to use [manhattan|euclidian]")
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	filename := flag.Arg(0)

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	img, _, err = image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	err = file.Close()
	if err != nil {
		log.Fatal(err)
	}

	bounds := img.Bounds().Max
	MaxX, MaxY = bounds.X, bounds.Y
	for x := 0; x < MaxX; x++ {
		if !wallAt(x, 0) {
			orig = Point(image.Pt(x, 0))
			break
		}
	}
	for x := 0; x < MaxX; x++ {
		if !wallAt(x, bounds.Y-1) {
			dest = Point(image.Pt(x, bounds.Y-1))
			break
		}
	}

	gif = gifpkg.GIF{LoopCount: -1}
	pq := pqueue.New(func(p1, p2 Point) bool {
		fi, fj := float64(global[p1]), float64(global[p2])
		if !useDijkstra {
			fi += heuristics(p1)
			fj += heuristics(p2)
		}
		return fi < fj
	})
	pq.Push(Point(orig))
	global[orig] = 0
	for !pq.Empty() {
		p, _ := pq.Pop()
		passed[p] = true
		if p == dest {
			break
		}
		pdist := global[p]
		for _, adj := range adjacents(p) {
			adjDist, ok := global[adj]
			if ok && adjDist <= (pdist+1) {
				continue
			}
			cameFrom[adj] = p
			global[adj] = pdist + 1
			pq.Push(adj)
		}
		if writeAsGIF {
			appendFrame()
		}
	}
	appendFrame()
	frame := gif.Image[len(gif.Image)-1]
	for p := cameFrom[dest]; p != orig; p = cameFrom[p] {
		frame.SetColorIndex(p.X, p.Y, RedIndex)
	}

	file, err = os.Create("out.gif")
	if err != nil {
		log.Fatal(err)
	}
	err = gifpkg.EncodeAll(file, &gif)
	if err != nil {
		log.Fatal(err)
	}
	err = file.Close()
	if err != nil {
		log.Fatal(err)
	}
}
