package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	gifpkg "image/gif"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"

	"mazesolver/pqueue"
)

type Point image.Point

func main() {

	const (
		BlackIndex uint8 = iota
		WhiteIndex
		RedIndex
		BlueIndex
		GreenIndex
		OrangeIndex

		frameDelay = 5
	)

	var (
		allowedDiagonals bool
		writeAsGIF       bool
		useDijkstra      bool
		heuristics       = HeuristicFunc(manhattan)

		palette = color.Palette{
			color.RGBA{0, 0, 0, 255},
			color.RGBA{255, 255, 255, 255},
			color.RGBA{255, 0, 0, 255},
			color.RGBA{0, 0, 255, 255},
			color.RGBA{0, 255, 0, 255},
			color.RGBA{255, 215, 0, 255},
		}

		input      image.Image
		frame      *image.Paletted
		gif        gifpkg.GIF
		MaxX, MaxY int

		adjsBuff   [8]Point
		orig, dest Point
		distances  = map[Point]int{}
		cameFrom   = map[Point]Point{}
		visited    = map[Point]bool{}
	)

	decodeInput := func(filename string) (image.Image, error) {
		file, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		input, _, err = image.Decode(file)
		if err != nil {
			return nil, err
		}
		err = file.Close()
		if err != nil {
			return nil, err
		}
		return input, nil
	}

	encodeSolution := func() error {
		var filename string
		if writeAsGIF {
			filename = "output.gif"
		} else {
			filename = "output.png"
		}

		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		if writeAsGIF {
			err = gifpkg.EncodeAll(file, &gif)
		} else {
			err = png.Encode(file, frame)
		}
		if err != nil {
			return err
		}
		return file.Close()
	}

	wallAt := func(x, y int) bool {
		r, g, b, _ := input.At(x, y).RGBA()
		return r == 0 && g == 0 && b == 0
	}

	adjacents := func(p Point) []Point {
		buff := adjsBuff[:0]
		px, py := p.X, p.Y
		adjacents := []Point{
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
			adjacents = adjacents[:4]
		}

		for _, adj := range adjacents {
			if adj.X < 0 || adj.X == MaxX {
				continue
			}
			if adj.Y < 0 || adj.Y == MaxY {
				continue
			}
			if wallAt(adj.X, adj.Y) {
				continue
			}
			buff = append(buff, adj)
		}
		return buff
	}

	appendFrame := func() {
		new := image.NewPaletted(frame.Bounds(), frame.Palette)
		copy(new.Pix, frame.Pix)
		gif.Image = append(gif.Image, new)
		gif.Delay = append(gif.Delay, frameDelay)
	}

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
	ifilename := flag.Arg(0)

	input, err := decodeInput(ifilename)
	if err != nil {
		log.Fatal(err)
	}

	bounds := input.Bounds().Max
	MaxX, MaxY = bounds.X, bounds.Y
	for x := 0; x < MaxX; x++ {
		if !wallAt(x, 0) {
			orig = Point(image.Pt(x, 0))
			break
		}
	}
	for x := 0; x < MaxX; x++ {
		if !wallAt(x, MaxY-1) {
			dest = Point(image.Pt(x, MaxY-1))
			break
		}
	}

	frame = image.NewPaletted(input.Bounds(), palette)
	for y := 0; y < MaxY; y++ {
		for x := 0; x < MaxX; x++ {
			frame.Set(x, y, input.At(x, y))
		}
	}

	gif = gifpkg.GIF{LoopCount: -1}
	pq := pqueue.New(func(p1, p2 Point) bool {
		d1, d2 := float64(distances[p1]), float64(distances[p2])
		if !useDijkstra {
			d1 += heuristics(dest, p1)
			d2 += heuristics(dest, p2)
		}
		return d1 < d2
	})

	distances[orig] = 0
	pq.Push(Point(orig))
	for !pq.Empty() {
		p, _ := pq.Pop()
		visited[p] = true
		frame.SetColorIndex(p.X, p.Y, BlueIndex)
		if writeAsGIF {
			appendFrame()
		}
		if p == dest {
			break
		}
		dist := distances[p]
		for _, adj := range adjacents(p) {
			adjDist, ok := distances[adj]
			if ok && adjDist <= (dist+1) {
				continue
			}
			cameFrom[adj] = p
			distances[adj] = dist + 1
			pq.Push(adj)
		}
	}
	for p := dest; p != orig; p = cameFrom[p] {
		frame.SetColorIndex(p.X, p.Y, RedIndex)
	}
	if writeAsGIF {
		appendFrame()
	}

	if err := encodeSolution(); err != nil {
		log.Fatal(err)
	}
}
