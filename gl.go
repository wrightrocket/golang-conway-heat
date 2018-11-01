package main

import (
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl" // OR: github.com/go-gl/gl/v2.1/gl
	"github.com/go-gl/glfw/v3.2/glfw"
	"log"
	"strings"
)

const (
	fragmentShaderSourceBlue = `
		#version 410
		out vec4 frag_colour;
		void main() {
			frag_colour = vec4(0, 0.0, 0.9, 0.8);
		}
	` + "\x00"

	fragmentShaderSourceGreen = `
		#version 410
		out vec4 frag_colour;
		void main() {
			frag_colour = vec4(0, 0.8, 0, 1.0);
		}
	` + "\x00"

	fragmentShaderSourcePurple = `
		#version 410
		out vec4 frag_colour;
		void main() {
			frag_colour = vec4(0.9, 0.0, 0.9, 0.6);
		}
	` + "\x00"

	fragmentShaderSourceRed = `
		#version 410
		out vec4 frag_colour;
		void main() {
			frag_colour = vec4(1.0, 0.0, 0, 1.0);
		}
	` + "\x00"

	fragmentShaderSourceWhite = `
		#version 410
		out vec4 frag_colour;
		void main() {
			frag_colour = vec4(1.0, 1.0, 1.0, 1.0);
		}
	` + "\x00"

	fragmentShaderSourceYellow = `
                #version 410
                out vec4 frag_colour;
                void main() {
                        frag_colour = vec4(0.9, 0.9, 0, 1.0);
                }
        ` + "\x00"

	fragmentVertexShaderSource = `
		#version 410
		in vec3 vp;
		void main() {
			gl_Position = vec4(vp, 1.0);
		}
	` + "\x00"
)

var (
	fragmentShaderBlue   uint32
	fragmentShaderGreen  uint32
	fragmentShaderPurple uint32
	fragmentShaderRed    uint32
	fragmentShaderWhite  uint32
	fragmentShaderYellow uint32
	fragmentVertexShader uint32

	height = width // assignment to variable not declared yet,yes!

	square = []float32{
		-0.5, 0.5, 0,
		-0.5, -0.5, 0,
		0.5, -0.5, 0,

		-0.5, 0.5, 0,
		0.5, 0.5, 0,
		0.5, -0.5, 0,
	}

	width = 5 * grid // TODO
)

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
			gl.AttachShader(program, fragmentShaderWhite)
			gl.LinkProgram(program)
			gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
			gl.DetachShader(program, fragmentShaderWhite)
		}
	} else if showNext && c.aliveNext {
		gl.BindVertexArray(c.drawable)
		gl.AttachShader(program, fragmentShaderPurple)
		gl.LinkProgram(program)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
		gl.DetachShader(program, fragmentShaderPurple)
	} else {
		return
	}
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
	var err error
	if err = gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	fragmentShaderBlue, err = compileShader(fragmentShaderSourceBlue, gl.FRAGMENT_SHADER)
	fragmentShaderGreen, err = compileShader(fragmentShaderSourceGreen, gl.FRAGMENT_SHADER)
	fragmentShaderPurple, err = compileShader(fragmentShaderSourcePurple, gl.FRAGMENT_SHADER)
	fragmentShaderRed, err = compileShader(fragmentShaderSourceRed, gl.FRAGMENT_SHADER)
	fragmentShaderWhite, err = compileShader(fragmentShaderSourceWhite, gl.FRAGMENT_SHADER)
	fragmentShaderYellow, err = compileShader(fragmentShaderSourceYellow, gl.FRAGMENT_SHADER)
	fragmentVertexShader, err := compileShader(fragmentVertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	program = gl.CreateProgram()
	gl.AttachShader(program, fragmentVertexShader)
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
