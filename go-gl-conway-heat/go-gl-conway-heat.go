package main

import (
	"flag"
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl" // OR: github.com/go-gl/gl/v2.1/gl
	"github.com/go-gl/glfw/v3.2/glfw"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	EXIT_NO_LIFE       = 1
	EXIT_STABLE_LIFE   = 2
	EXIT_TOTAL_TIME    = 3
	EXIT_TOTAL_ROUNDS  = 4
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
	program              uint32
	vertexShader         uint32
	fragmentShaderRed    uint32
	fragmentShaderYellow uint32
	fragmentShaderGreen  uint32
	fragmentShaderBlue   uint32

	grid   = 100      // TODO
	width  = 5 * grid // TODO
	height = width    // TODO

	seed                = time.Now().UnixNano()
	probability         = 0.15
	probability_default = 0.15
	fps                 = 5
	fps_default         = 5
	alivePercent        = 0.0
	report              = 0
	timeStart           = time.Now()
	timeDelay           = "5s"
	timeExpire          = "0d0h0m0s"
	timeTotal           = "0s"
	timeToSleep         time.Duration
	timeDuration        time.Duration
	showNext            = true
	showColor           = true
	maxTurns            = 0
	square              = []float32{
		-0.5, 0.5, 0,
		-0.5, -0.5, 0,
		0.5, -0.5, 0,

		-0.5, 0.5, 0,
		0.5, 0.5, 0,
		0.5, -0.5, 0,
	}
)

func parseFlags() {
	flag.BoolVar(&showColor, "c", showColor, "Same as -color.")
	flag.BoolVar(&showColor, "color", showColor,
		"If true, the number of neighbors a live cell is colored red > 3, yellow = 3, green = 2, and blue < 2. If false, then live cells will appear white.")
	flag.StringVar(&timeDelay, "d", timeDelay, "Same as -delay.")
	flag.StringVar(&timeDelay, "delay", timeDelay,
		"Sets the amount of time to delay at the end of the game.")
	flag.StringVar(&timeExpire, "e", timeExpire, "Same as -expire.")
	flag.StringVar(&timeExpire, "expire", timeExpire,
		"Sets the amount of time to run the game. When -expire is a zero duration, it removes any time constraint.")
	flag.IntVar(&fps, "f", fps, "Same as -fps.")
	flag.IntVar(&fps, "fps", fps,
		"Sets the frames-per-second, used set the speed of the simulation.")
	flag.IntVar(&grid, "g", grid, "Same as -grid.")
	flag.IntVar(&grid, "grid", grid,
		"Sets both the number of rows and columns for the game grid.")
	flag.BoolVar(&showNext, "n", showNext, "Same as -next.")
	flag.BoolVar(&showNext, "next", showNext,
		"Boolean to determine if next alive cell is shown as a purple color. ")
	flag.Float64Var(&probability, "p", probability, "Same as -probability.")
	flag.Float64Var(&probability, "probability", probability,
		"A percentage between 0 and 1 used in conjunction with the -seed to determine if a cell starts alive. For example, 0.15 means each cell has a 15% probability of starting alive.")
	flag.IntVar(&report, "r", report, "Same as -report.")
	flag.IntVar(&report, "report", report,
		"Sets the output report. 1: detailed, 2: comma separated, 3: space separated, 4: round number and alive percentage. The default is no output.")
	flag.Int64Var(&seed, "s", seed, "Same as -seed.")
	flag.Int64Var(&seed, "seed", seed,
		"Sets the starting seed of the game, used to randomize the initial state.")
	flag.IntVar(&maxTurns, "t", maxTurns, "Same as -turns")
	flag.IntVar(&maxTurns, "turns", maxTurns,
		"Integer for how many turns to execute. When -turns is zero, it removes any constraint on the number of turns.")
	flag.Parse()
	if fps > 60 || fps < 0 {
		fps = fps_default
	}
	if probability > 1.0 || probability < 0.01 {
		probability = probability_default
	}
	//probability = float64(grid) / 100.0  * probability
	width = 5 * grid // TODO
	height = width   // TODO
	timeDuration, _ = time.ParseDuration(timeExpire)
	timeToSleep, _ = time.ParseDuration(timeDelay) // TODO ERROR
	fmt.Println("Using following values:")
	fmt.Println("color", showColor)
	fmt.Println("delay", timeDelay)
	fmt.Println("expire", timeExpire)
	fmt.Println("fps", fps)
	fmt.Println("grid", grid)
	fmt.Println("next", showNext)
	fmt.Println("probability", probability)
	fmt.Println("report", report)
	fmt.Println("seed", seed)
	fmt.Println("turns", maxTurns)
}

type cell struct {
	drawable uint32

	alive     bool
	aliveNext bool
	color     uint32

	x int
	y int
}

func main() {
	parseFlags()
	runtime.LockOSThread()
	window := initGlfw()
	defer glfw.Terminate()
	initOpenGL()
	cells := makeCells()
	cellsTotal := float64(len(cells) * 100.0)
	turns := 0
	aliveTotal := probability
	aliveTotalLast := 0.0
	aliveTotalRepeated := 0
	for !window.ShouldClose() {
		t := time.Now()
		totalTime := time.Since(timeStart)
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
		turns += 1
		aliveTotalRepeated = checkTurn(aliveTotal, aliveTotalLast, aliveTotalRepeated, turns, totalTime)
		outputReport(aliveTotal, cellsTotal, turns)
		draw(cells, window)

		time.Sleep(time.Second/time.Duration(fps) - time.Since(t))
	}
}

func checkTurn(aliveTotal float64, aliveTotalLast float64, aliveTotalRepeated int, turns int, totalTime time.Duration) int {

	if aliveTotal == 0 {
		fmt.Println("Life has died out completely.")
		os.Exit(EXIT_NO_LIFE)
	}
	if aliveTotal == aliveTotalLast { // make checkAlive function TODO
		aliveTotalRepeated += 1
		if aliveTotalRepeated > 1 {
			fmt.Println("Initial probability of life", fmt.Sprintf("% 5.2f%%",
				probability), "has stabilized at", aliveTotal,
				"lives after", turns, "turns")
			fmt.Println("Delaying for", fmt.Sprintf("%s", timeToSleep))
			time.Sleep(timeToSleep) // TODO -t timeout
			os.Exit(EXIT_STABLE_LIFE)
		}
	} else {
		aliveTotalRepeated = 0
	}
	if timeDuration > time.Second && totalTime > timeDuration {
		fmt.Println("Life has stopped running after", fmt.Sprintf("%v", timeExpire),
			"according to the timeExpire parameter")
		os.Exit(EXIT_TOTAL_TIME)
	}

	if maxTurns > 0 && turns > maxTurns {

		fmt.Println("Life has stopped running after", fmt.Sprintf("%v", turns-1),
			"turns, according to the maxTurns parameter")
		os.Exit(EXIT_TOTAL_ROUNDS)
	}
	return aliveTotalRepeated
}

func outputReport(aliveTotal float64, cellsTotal float64, turns int) {
	alivePercent := aliveTotal / cellsTotal * 100
	alivePercentString := fmt.Sprintf("% 9.2f%%", alivePercent)
	switch report {
	case 1:
		fmt.Println(alivePercentString, " life with", aliveTotal,
			"cells alive and", cellsTotal, "total cells after", turns, "turns")
	case 2:
		fmt.Printf("%v,%v,%v,%5.2f\n", turns, aliveTotal, cellsTotal, alivePercent)
	case 3:
		fmt.Println(turns, aliveTotal, cellsTotal, alivePercentString)
	case 4:
		fmt.Println("Turn:", fmt.Sprintf("% 7.0f", float64(turns)),
			"         Alive:", alivePercentString)
	}

}

func draw(cells [][]*cell, window *glfw.Window) {
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
	if seed > 0 {
		seed = time.Now().UnixNano()
	}
	rand.Seed(seed)

	cells := make([][]*cell, grid, grid)
	for x := 0; x < grid; x++ {
		for y := 0; y < grid; y++ {
			c := newCell(x, y)

			c.alive = rand.Float64() < probability
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
			size = 5.0 / float32(width)
			position = float32(x) * size
		case 1:
			size = 5.0 / float32(height)
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
	if c.alive {
		if showColor {
			gl.BindVertexArray(c.drawable)
			gl.AttachShader(program, c.color)
			gl.LinkProgram(program)
			gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
			gl.DetachShader(program, c.color)

		} else {
			gl.BindVertexArray(c.drawable)
			gl.AttachShader(program, fragmentShaderBlue)
			gl.AttachShader(program, fragmentShaderRed)
			gl.AttachShader(program, fragmentShaderGreen)
			gl.LinkProgram(program)
			gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
			gl.DetachShader(program, fragmentShaderBlue)
			gl.DetachShader(program, fragmentShaderRed)
			gl.DetachShader(program, fragmentShaderGreen)
		}
	} else if showNext && c.aliveNext {
		gl.BindVertexArray(c.drawable)
		gl.AttachShader(program, fragmentShaderBlue)
		gl.AttachShader(program, fragmentShaderRed)
		gl.LinkProgram(program)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
		gl.DetachShader(program, c.color)
	} else {
		return
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

// initOpenGL initializes OpenGL and set global program.
func initOpenGL() {
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

	program = gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.LinkProgram(program)
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
