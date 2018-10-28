package main

import (
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl" // OR: github.com/go-gl/gl/v2.1/gl
	"github.com/go-gl/glfw/v3.2/glfw"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	EXIT_TOTAL_TIME    = 0
	EXIT_NO_LIFE       = 1
	EXIT_STABLE_LIFE   = 2
	vertexShaderSource = `
		#version 410
		in vec3 vp;
		void main() {
			gl_Position = vec4(vp, 1.0);
		}
	` + "\x00"

	fragmentShaderSourceRed = `
		#version 410
		out vec4 frag_colour;
		void main() {
			frag_colour = vec4(1.0, 0.0, 0, 1.0);
		}
	` + "\x00"

	fragmentShaderSourceYellow = `
                #version 410
                out vec4 frag_colour;
                void main() {
                        frag_colour = vec4(0.9, 0.9, 0, 1.0);
                }
        ` + "\x00"

	fragmentShaderSourceGreen = `
		#version 410
		out vec4 frag_colour;
		void main() {
			frag_colour = vec4(0, 0.8, 0, 1.0);
		}
	` + "\x00"

	fragmentShaderSourceBlue = `
		#version 410
		out vec4 frag_colour;
		void main() {
			frag_colour = vec4(0, 0.2, 0.9, 0.8);
		}
	` + "\x00"
)

var (
	prog                 uint32
	vertexShader         uint32
	fragmentShaderRed    uint32
	fragmentShaderYellow uint32
	fragmentShaderGreen  uint32
	fragmentShaderBlue   uint32

	width   = 500 // TODO  add key binding to change ...
	height  = 500 // TODO
	rows    = 100 // TODO
	columns = 100 // TODO

	threshold    float64 // TODO
	fps          int     // TODO add key binding to change
	alivePercent = 0.0
	fmtChosen    = 0 // TODO add flag
	timeStart    = time.Now()
	timeAtEnd    = "5s"
	timeToRun    = "10s"
	timeTotal    = "0s"
	square       = []float32{
		-0.5, 0.5, 0,
		-0.5, -0.5, 0,
		0.5, -0.5, 0,

		-0.5, 0.5, 0,
		0.5, 0.5, 0,
		0.5, -0.5, 0,
	}
)

type cell struct {
	drawable uint32

	alive     bool
	aliveNext bool
	color     uint32

	x int
	y int
}

func main() {
	cliArgs := os.Args
	args := len(cliArgs)
	var err error
	// [1] fps [2] threshold
	switch args {
	case 3:
		fps, err = strconv.Atoi(cliArgs[1])
		if fps > 60 || fps < 0 {
			fps = 1
		}
		if err != nil {
			fmt.Println("TODO: Oops")
		}
		threshold, err = strconv.ParseFloat(cliArgs[2], 64)
		if threshold > 1.0 || threshold < 0.01 {
			threshold = 0.25
		}
		if err != nil {
			fmt.Println("TODO: Oops")
		}
	case 2:
		fps, err = strconv.Atoi(cliArgs[1])
		if fps > 60 || fps < 0 {
			fps = 1
		}
		if err != nil {
			fmt.Println("TODO: Oops")
		}
		if threshold > 1.0 || threshold < 0.01 {
			threshold = 0.25
		}
	default:
		fps = 1
		threshold = 0.025
		fmt.Println("Usage: go-gl-conway [fps] [%initial_life]")
	}
	fmt.Println("Using: fps =", fps, "and threshold =", threshold)
	runtime.LockOSThread()

	window := initGlfw()
	defer glfw.Terminate()
	program := initOpenGL()
	prog = program
	cells := makeCells()
	cellsTotal := float64(len(cells) * 100.0)
	rounds := 0
	aliveTotal := threshold
	aliveTotalLast := 0.0
	aliveTotalRepeated := 0
	timeDuration, _ := time.ParseDuration(timeToRun)
	for !window.ShouldClose() {
		t := time.Now()
		totalTime := time.Since(timeStart)
		if aliveTotal == 0 {
			fmt.Println("Life has died out completely.")
			os.Exit(1)
		}
		if aliveTotal == aliveTotalLast { // make checkAlive function TODO
			aliveTotalRepeated += 1
			if aliveTotalRepeated > 1 {
				fmt.Println("Initial chance of life", fmt.Sprintf("% 5.2f%%",
					threshold), "has stabilized at", aliveTotal,
					"lives after", rounds, "rounds")
				timeToSleep, _ := time.ParseDuration(timeAtEnd) // TODO ERROR
				fmt.Println("Sleeping for", fmt.Sprintf("%s", timeToSleep))
				time.Sleep(timeToSleep) // TODO -t timeout
				os.Exit(2)
			}
		} else {
			aliveTotalRepeated = 0
		}
		if totalTime > timeDuration {
			fmt.Println("Life has stopped running after", fmt.Sprintf("%v", timeToRun),
				"according to the timeToRun parameter")
			os.Exit(0)
		}
		aliveTotalLast = aliveTotal
		aliveTotal = 0
		for x := range cells {
			for _, c := range cells[x] {
				c.checkState(cells)
				if c.alive {
					aliveTotal += 1.0
				}
			}
		}
		alivePercent = aliveTotal / cellsTotal * 100
		alivePercentString := fmt.Sprintf("% 9.2f%%", alivePercent)
		rounds += 1
		switch fmtChosen {
		case 1:
			fmt.Println(" life with", aliveTotal, "cells alive and", cellsTotal, "total cells after", rounds, "rounds")
		case 2:
			fmt.Printf("%v,%v,%v,%5.2f\n", rounds, aliveTotal, cellsTotal, alivePercent)
		case 3:
			fmt.Println(rounds, aliveTotal, cellsTotal, alivePercentString)
		default:
			fmt.Println("Round:", fmt.Sprintf("% 7.0f", float64(rounds)),
				"         Alive:", alivePercentString)

		}
		draw(cells, window, program)

		time.Sleep(time.Second/time.Duration(fps) - time.Since(t))
	}
}

func draw(cells [][]*cell, window *glfw.Window, program uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	for x := range cells {
		for _, c := range cells[x] {
			c.draw()
		}
	}

	glfw.PollEvents()
	window.SwapBuffers()
}

func makeCells() [][]*cell {
	rand.Seed(time.Now().UnixNano())

	cells := make([][]*cell, rows, rows)
	for x := 0; x < rows; x++ {
		for y := 0; y < columns; y++ {
			c := newCell(x, y)

			c.alive = rand.Float64() < threshold
			c.aliveNext = c.alive

			cells[x] = append(cells[x], c)
		}
	}

	return cells
}
func newCell(x, y int) *cell {
	points := make([]float32, len(square), len(square))
	copy(points, square)

	for i := 0; i < len(points); i++ {
		var position float32
		var size float32
		switch i % 3 {
		case 0:
			size = 1.0 / float32(columns)
			position = float32(x) * size
		case 1:
			size = 1.0 / float32(rows)
			position = float32(y) * size
		default:
			continue
		}

		if points[i] < 0 {
			points[i] = (position * 2) - 1
		} else {
			points[i] = ((position + size) * 2) - 1
		}
	}

	return &cell{
		drawable: makeVao(points),

		x: x,
		y: y,
	}
}

func (c *cell) draw() {
	if !c.alive {
		return
	}

	/*gl.AttachShader(prog, vertexShader)
	gl.BindVertexArray(c.drawable)
	gl.AttachShader(prog, c.color)
	gl.LinkProgram(prog)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
	gl.DetachShader(prog, fragmentShaderRed)
	*/
	gl.BindVertexArray(c.drawable)
	shaderColor := c.color
	switch shaderColor {
	case fragmentShaderRed:
		gl.AttachShader(prog, c.color)
		gl.LinkProgram(prog)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
		gl.DetachShader(prog, fragmentShaderRed)
	case fragmentShaderYellow:
		gl.AttachShader(prog, fragmentShaderYellow)
		gl.LinkProgram(prog)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
		gl.DetachShader(prog, fragmentShaderYellow)
	case fragmentShaderGreen:
		gl.AttachShader(prog, fragmentShaderGreen)
		gl.LinkProgram(prog)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
		gl.DetachShader(prog, fragmentShaderGreen)
	case fragmentShaderBlue:
		gl.AttachShader(prog, fragmentShaderBlue)
		gl.LinkProgram(prog)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
		gl.DetachShader(prog, fragmentShaderBlue)
	}
}

// checkState determines the state of the cell for the next tick of the game.
func (c *cell) checkState(cells [][]*cell) {
	c.alive = c.aliveNext
	// c.aliveNext = c.alive

	liveCount := c.liveNeighbors(cells)
	if c.alive {
		// 1. Any live cell with fewer than two live neighbours dies, as if caused by underpopulation.
		if liveCount < 2 {
			c.aliveNext = false
			c.color = fragmentShaderBlue
		}

		// 2. Any live cell with two or three live neighbours lives on to the next generation.
		if liveCount == 2 {
			c.aliveNext = true
			c.color = fragmentShaderGreen
		}

		if liveCount == 3 {
			c.aliveNext = true
			c.color = fragmentShaderYellow
		}

		// 3. Any live cell with more than three live neighbours dies, as if by overpopulation.
		if liveCount > 3 {
			c.aliveNext = false
			c.color = fragmentShaderRed
		}
	} else {
		// 4. Any dead cell with exactly three live neighbours becomes a live cell, as if by reproduction.
		if liveCount == 3 {
			c.aliveNext = true
		}
	}
}

// liveNeighbors returns the number of live neighbors for a cell.
func (c *cell) liveNeighbors(cells [][]*cell) int {
	var liveCount int
	add := func(x, y int) {
		// If we're at an edge, check the other side of the board.
		if x == len(cells) {
			x = 0
		} else if x == -1 {
			x = len(cells) - 1
		}
		if y == len(cells[x]) {
			y = 0
		} else if y == -1 {
			y = len(cells[x]) - 1
		}

		if cells[x][y].alive {
			liveCount++
		}
	}

	add(c.x-1, c.y)   // To the left
	add(c.x+1, c.y)   // To the right
	add(c.x, c.y+1)   // up
	add(c.x, c.y-1)   // down
	add(c.x-1, c.y+1) // top-left
	add(c.x+1, c.y+1) // top-right
	add(c.x-1, c.y-1) // bottom-left
	add(c.x+1, c.y-1) // bottom-right

	return liveCount
}

// initGlfw initializes glfw and returns a Window to use.
func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "Conway's Game of Life", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}

// initOpenGL initializes OpenGL and returns an intiialized program.
func initOpenGL() uint32 {
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	vertShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	vertexShader = vertShader
	fragmentShaderRed, err = compileShader(fragmentShaderSourceRed, gl.FRAGMENT_SHADER)
	fragmentShaderGreen, err = compileShader(fragmentShaderSourceGreen, gl.FRAGMENT_SHADER)
	fragmentShaderBlue, err = compileShader(fragmentShaderSourceBlue, gl.FRAGMENT_SHADER)
	fragmentShaderYellow, err = compileShader(fragmentShaderSourceYellow, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vertexShader)
	gl.LinkProgram(prog)
	return prog
}

// makeVao initializes and returns a vertex array from the points provided.
func makeVao(points []float32) uint32 {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	return vao
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}
