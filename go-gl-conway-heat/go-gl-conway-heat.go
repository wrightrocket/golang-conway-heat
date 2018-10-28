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
	EXIT_TOTAL_ROUNDS    = 4
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

	grid   = 100      // TODO
	width  = 5 * grid // TODO
	height = width    // TODO

	seed           = time.Now().UnixNano()
	chance         = 0.15
	chance_default = 0.15
	fps            = 5
	fps_default    = 5
	alivePercent   = 0.0
	format         = 0 // TODO add flag
	timeStart      = time.Now()
	timedelay      = "5s"
	timerun        = "0d0h0m0s"
	timeTotal      = "0s"
	timeToSleep   time.Duration
	timeDuration  time.Duration
	shownext 	= true
	showcolor 	= true
	maxrounds = 0
	square         = []float32{
		-0.5, 0.5, 0,
		-0.5, -0.5, 0,
		0.5, -0.5, 0,

		-0.5, 0.5, 0,
		0.5, 0.5, 0,
		0.5, -0.5, 0,
	}
)

func parseFlags() {
	flag.IntVar(&format, "format", format, "Sets the output format. 0 is for round and alive percentage, 1 is for detailed, 2 for comma separated, and 3 is for space separated.")
	flag.IntVar(&maxrounds, "maxrounds", maxrounds, "Integer for how many rounds to execute.")
	flag.IntVar(&grid, "grid", grid, "Sets the number of rows and columns for the game grid.")
	flag.Int64Var(&seed, "seed", seed, "Sets the starting seed of the game, used to randomize the initial state.")
	flag.Float64Var(&chance, "chance", chance, "A percentage between 0 and 1 used in conjunction with the -seed to determine if a cell starts alive. For example, 0.15 means each cell has a 15% chance of starting alive.")
	flag.IntVar(&fps, "fps", fps, "Sets the frames-per-second, used set the speed of the simulation.")
	flag.StringVar(&timerun, "timerun", timerun, "Sets the amount of time to run the game.")
	flag.StringVar(&timedelay, "timedelay", timedelay, "Sets the amount of time to delay at the end of the game.")
	flag.BoolVar(&shownext, "shownext", shownext, "Boolean to determine if next alive cell is shown.")
	flag.BoolVar(&showcolor, "showcolor", showcolor, "Boolean to determine if cells are colored.")
	flag.Parse()
	if fps > 60 || fps < 0 {
		fps = fps_default
	}
	if chance > 1.0 || chance < 0.01 {
		chance = chance_default
	}
	//chance = float64(grid) / 100.0  * chance
	width = 5 * grid // TODO
	height = width   // TODO
	timeDuration, _ = time.ParseDuration(timerun)
	timeToSleep, _ = time.ParseDuration(timedelay) // TODO ERROR
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
	fmt.Println("Using: fps =", fps, "and chance =", chance)
	runtime.LockOSThread()

	window := initGlfw()
	defer glfw.Terminate()
	program := initOpenGL()
	prog = program
	cells := makeCells()
	cellsTotal := float64(len(cells) * 100.0)
	rounds := 0
	aliveTotal := chance
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
		aliveTotalRepeated = roundCheck(aliveTotal, aliveTotalLast, aliveTotalRepeated, rounds, totalTime)
		rounds += 1
		outputFormat(aliveTotal, cellsTotal, rounds)
		draw(cells, window, program)

		time.Sleep(time.Second/time.Duration(fps) - time.Since(t))
	}
}

func roundCheck(aliveTotal float64, aliveTotalLast float64, aliveTotalRepeated int, rounds int, totalTime time.Duration) (int) {

		if aliveTotal == 0 {
			fmt.Println("Life has died out completely.")
			os.Exit(EXIT_NO_LIFE)
		}
		if aliveTotal == aliveTotalLast { // make checkAlive function TODO
			aliveTotalRepeated += 1
			if aliveTotalRepeated > 1 {
				fmt.Println("Initial chance of life", fmt.Sprintf("% 5.2f%%",
					chance), "has stabilized at", aliveTotal,
					"lives after", rounds, "rounds")
				fmt.Println("Sleeping for", fmt.Sprintf("%s", timeToSleep))
				time.Sleep(timeToSleep) // TODO -t timeout
				os.Exit(EXIT_STABLE_LIFE)
			} 
		} else {
			aliveTotalRepeated = 0
		}
		if timeDuration > time.Second && totalTime > timeDuration {
			fmt.Println("Life has stopped running after", fmt.Sprintf("%v", timerun),
				"according to the timerun parameter")
			os.Exit(EXIT_TOTAL_TIME)
		}

		if maxrounds > 0 && rounds > maxrounds {

			fmt.Println("Life has stopped running after", fmt.Sprintf("%v", rounds),
				"according to the maxrounds parameter")
			os.Exit(EXIT_TOTAL_ROUNDS)
		}
		return aliveTotalRepeated
	}
func outputFormat(aliveTotal float64, cellsTotal float64, rounds int) {
	alivePercent := aliveTotal / cellsTotal * 100
	alivePercentString := fmt.Sprintf("% 9.2f%%", alivePercent)
	switch format {
	case 1:
		fmt.Println(alivePercentString, " life with", aliveTotal,
			"cells alive and", cellsTotal, "total cells after", rounds, "rounds")
	case 2:
		fmt.Printf("%v,%v,%v,%5.2f\n", rounds, aliveTotal, cellsTotal, alivePercent)
	case 3:
		fmt.Println(rounds, aliveTotal, cellsTotal, alivePercentString)
	default:
		fmt.Println("Round:", fmt.Sprintf("% 7.0f", float64(rounds)),
			"         Alive:", alivePercentString)

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
	if seed > 0 {
		seed = time.Now().UnixNano()
	}
	rand.Seed(seed)

	cells := make([][]*cell, grid, grid)
	for x := 0; x < grid; x++ {
		for y := 0; y < grid; y++ {
			c := newCell(x, y)

			c.alive = rand.Float64() < chance
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
		if showcolor {
		gl.BindVertexArray(c.drawable)
		gl.AttachShader(prog, c.color)
		gl.LinkProgram(prog)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
		gl.DetachShader(prog, c.color)

	} else {
                gl.BindVertexArray(c.drawable)
                gl.AttachShader(prog, fragmentShaderBlue)
                gl.AttachShader(prog, fragmentShaderRed)
                gl.AttachShader(prog, fragmentShaderGreen)
                gl.LinkProgram(prog)
                gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
                gl.DetachShader(prog, fragmentShaderBlue)
                gl.DetachShader(prog, fragmentShaderRed)
                gl.DetachShader(prog, fragmentShaderGreen)
	}
	} else if shownext && c.aliveNext {
		gl.BindVertexArray(c.drawable)
		gl.AttachShader(prog, fragmentShaderBlue)
		gl.AttachShader(prog, fragmentShaderRed)
		gl.LinkProgram(prog)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
		gl.DetachShader(prog, c.color)
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
	//width, height := window.GetFramebufferSize()
	//gl.Viewport(0, 0, int32(width), int32(height))
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
