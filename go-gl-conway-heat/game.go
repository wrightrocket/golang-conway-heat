package main

import (
	"flag"
	"fmt"
	"github.com/go-gl/glfw/v3.2/glfw"
	"os"
	"runtime"
	"time"
)

const (
	EXIT_NO_LIFE     = 1
	EXIT_STABLE_LIFE = 2
	EXIT_TOTAL_TIME  = 3
	EXIT_TOTAL_TURNS = 4
)

var (
	alivePercent = 0.0
	fps          = 5
	fps_default  = 5
	grid         = 100 // TODO
	maxTurns     = 0
	odds         = 0.15
	odds_default = 0.15
	program      uint32
	report       = 0
	showColor    = true
	showNext     = true
	timeDelay    = "5s"
	timeDuration time.Duration
	timeExpire   = "0d0h0m0s"
	timeStart    = time.Now()
	timeToSleep  time.Duration
	timeTotal    = "0s"
)

func main() {
	var (
		aliveTotal         float64
		aliveTotalLast     float64
		aliveTotalRepeated int
		cells              [][]*cell
		cellsTotal         float64
		timeLast           time.Time
		turns              int
		totalTime          time.Duration
		window             *glfw.Window
	)
	parseFlags()
	runtime.LockOSThread()
	window = initGlfw()
	defer glfw.Terminate()
	initOpenGL()
	cells = makeCells()
	cellsTotal = float64(len(cells) * 100.0)
	for !window.ShouldClose() {
		timeLast = time.Now()
		totalTime = time.Since(timeStart)
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

		time.Sleep(time.Second/time.Duration(fps) - time.Since(timeLast))
	}
}

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
	flag.Float64Var(&odds, "o", odds, "Same as -odds.")
	flag.Float64Var(&odds, "odds", odds,
		"A percentage between 0 and 1 to determine if a cell starts alive. For example, 0.15 means each cell has a 15% chance of starting alive.")
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
	validateSettings()
	outputSettings()
}

func validateSettings() {
	if fps > 60 || fps < 0 {
		fps = fps_default
	}
	if odds > 1.0 || odds < 0.01 {
		odds = odds_default
	}
	//odds = float64(grid) / 100.0  * odds
	width = 5 * grid // TODO
	height = width   // TODO
	timeDuration, _ = time.ParseDuration(timeExpire)
	timeToSleep, _ = time.ParseDuration(timeDelay) // TODO ERROR
}

func outputSettings() {
	fmt.Println("Using following values:")
	fmt.Println("color", showColor)
	fmt.Println("delay", timeDelay)
	fmt.Println("expire", timeExpire)
	fmt.Println("fps", fps)
	fmt.Println("grid", grid)
	fmt.Println("next", showNext)
	fmt.Println("odds", odds)
	fmt.Println("report", report)
	fmt.Println("seed", seed)
	fmt.Println("turns", maxTurns)
}

func checkTurn(aliveTotal float64, aliveTotalLast float64, aliveTotalRepeated int, turns int, totalTime time.Duration) int {

	if aliveTotal == 0 {
		fmt.Println("Life has died out completely.")
		os.Exit(EXIT_NO_LIFE)
	}
	if aliveTotal == aliveTotalLast {
		aliveTotalRepeated += 1
		if aliveTotalRepeated > 1 {
			fmt.Println("Initial odds of life", fmt.Sprintf("% 5.2f%%",
				odds), "has stabilized at", aliveTotal,
				"lives after", turns, "turns")
			fmt.Println("Delaying for", fmt.Sprintf("%s", timeToSleep))
			time.Sleep(timeToSleep)
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
		os.Exit(EXIT_TOTAL_TURNS)
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
