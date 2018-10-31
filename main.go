package main

import (
	"fmt"
	"github.com/4ydx/gltext"
	"github.com/4ydx/gltext/v4.1"
	"github.com/go-gl/glfw/v3.2/glfw"
	"golang.org/x/image/math/fixed"
	"os"
	"runtime"
)

var (
	font *v41.Font
	useStrictCoreProfile = (runtime.GOOS == "darwin")
)

func main() {
	parseFlags()
	runtime.LockOSThread()
	window = initGlfw()
	defer glfw.Terminate()
	 // initOpenGL()
	err := glfw.Init()
	if err != nil {
		panic("glfw error")
	}
	defer glfw.Terminate()

//	version := gl.GoStr(gl.GetString(gl.VERSION))
//	fmt.Println("Opengl version", version)

	// code from here
	gltext.IsDebug = true
	// width, height := window.GetSize()
//	font.ResizeWindow(float32(width), float32(height))

	loadFont()
	gameloop()
	//for _, text := range txts {
	//		text.Release()
	//}
	font.Release()
}

	func loadFont() {
		config, err := gltext.LoadTruetypeFontConfig("fontconfigs", "font_1_honokamin")
		if err == nil {
			font, err = v41.NewFont(config)
			if err != nil {
				panic(err)
			}
			fmt.Println("Font loaded from disk...")
		} else {
			fd, err := os.Open("font/font_1_honokamin.ttf")
			if err != nil {
				panic(err)
			}
			defer fd.Close()

			// Japanese character ranges
			// http://www.rikai.com/library/kanjitables/kanji_codes.unicode.shtml
			runeRanges := make(gltext.RuneRanges, 0)
			runeRanges = append(runeRanges, gltext.RuneRange{Low: 32, High: 128})
			runeRanges = append(runeRanges, gltext.RuneRange{Low: 0x3000, High: 0x3030})
			runeRanges = append(runeRanges, gltext.RuneRange{Low: 0x3040, High: 0x309f})
			runeRanges = append(runeRanges, gltext.RuneRange{Low: 0x30a0, High: 0x30ff})
			runeRanges = append(runeRanges, gltext.RuneRange{Low: 0x4e00, High: 0x9faf})
			runeRanges = append(runeRanges, gltext.RuneRange{Low: 0xff00, High: 0xffef})

			scale := fixed.Int26_6(32)
			runesPerRow := fixed.Int26_6(128)
			config, err = gltext.NewTruetypeFontConfig(fd, scale, runeRanges, runesPerRow, 5)
			if err != nil {
				panic(err)
			}
			err = config.Save("fontconfigs", "font_1_honokamin")
			if err != nil {
				panic(err)
			}
			font, err = v41.NewFont(config)
			if err != nil {
				panic(err)
			}
		}
	}
