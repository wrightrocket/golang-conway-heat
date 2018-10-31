package main

import (
	"fmt"
	"github.com/4ydx/gltext"
	"github.com/4ydx/gltext/v4.1"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"golang.org/x/image/math/fixed"
	"math"
	"os"
	"runtime"
	"time"
)

var (
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

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("Opengl version", version)

	// code from here
	gltext.IsDebug = true

	var font *v41.Font

	
	width, height := window.GetSize()
	font.ResizeWindow(float32(width), float32(height))

	loadFont()
	str0 := "Welcome to Conway's Game of Life! "
	str1 := "Uses heat map for population density"
	str2 := "By Keith Wright (wrightrocket)"
	str3 := "October 30, 2018"
/*
	str0 := "! \" # $ ' ( ) + , - . / 0123456789 : "
	str1 := "; <=> ? @ ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	str2 := "[^_`] abcdefghijklmnopqrstuvwxyz {|}"
	str3 := "大好きどなにｂｃｄｆｇｈｊｋｌｍｎｐｑ"

	scaleMin, scaleMax := float32(1.0), float32(1.1)
	strs := []string{str0, str1, str2, str3}
	txts := []*v41.Text{}
	/* for _, str := range strs {
		text := v41.NewText(font, scaleMin, scaleMax)
		text.SetString(str)
		text.SetColor(mgl32.Vec3{1, 1, 1})
		text.FadeOutPerFrame = 0.01
		if gltext.IsDebug {
			for _, s := range str {
				fmt.Printf("%c: %d\n", s, rune(s))
			}
		}
		txts = append(txts, text)
	}

	start := time.Now()

	gl.ClearColor(0.4, 0.4, 0.4, 0.0)
	*/
	gameloop()
	for _, text := range txts {
		text.Release()
	}
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
