package main

import (
	"math/rand"
	"time"
)


func makeCells() [][]*cell {
	if seed > 0 {
		seed = time.Now().UnixNano()
	}
	rand.Seed(seed)

	cells := make([][]*cell, grid, grid)
	for x := 0; x < grid; x++ {
		for y := 0; y < grid; y++ {
			c := newCell(x, y)

			c.alive = rand.Float64() < odds
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

