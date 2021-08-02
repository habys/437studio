package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
	"github.com/habys/437studio/charsets"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/mattn/go-runewidth"
)

/*
 * A "dot" should be a single unicode grapheme cluster.
 * This must be a string as one display character can be
 * represented by multiple runes.
 *
 * For this value, runewidth.StringWidth should return 1.
 *
 * They can be single or double display width. For double
 * width characters, the x+1 character might need to be
 * enforced to be empty.
 */

var defStyle tcell.Style

type modes uint8

const (
	pencil modes = iota
	box
	circle
	paint
	erase
	modeend
)

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func ColorScreen(s tcell.Screen) {
	lFmt := "l: %.02f"
	uFmt := "u: %.02f"
	vFmt := "v: %v02f"
	w, h := s.Size()
	log.Println("color screen")
	s.Clear()
	style := defStyle
	//c, err := colorful.Hex("#ea0064")
	//if err != nil {
	//	log.Fatal(err)
	//}
	drawX := 0
	drawY := 1
	white := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
EarlyExit:
	for l := 0.0; l <= 1.0; l += .01 {
		s.Show()
		if Pause(s) == "break" {
			break EarlyExit
		}
		drawY = 0
		for v := -1.0; v <= 1.0; v += .05 {
			drawX = -20
			drawY += 1
			if drawY >= h {
				drawY = 1
			}
			for u := -1.0; u <= 1.0; u += .03 {
				drawX += 1
				if drawX > w {
					drawX = 1
				}
				//if drawY >= h {
				//	drawY = 1
				//	s.Show()
				//	if Pause(s) == "break" {
				//		break EarlyExit
				//	}
				//}
				//c := colorful.Luv(l, u, v).Clamped()
				if drawX > 0 {
					c := colorful.Luv(l, u, v)
					log.Printf("LUV values: %.2f, %.2f, %.2f", l, u, v)
					if c.IsValid() != true {
						r, g, b, a := c.RGBA()
						log.Printf("Invalid RGB: R%d G%d G%d A%d", r, g, b, a)
					} else {
						r, g, b := c.RGB255()
						style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(int32(r), int32(g), int32(b)))
						s.SetContent(drawX, drawY, '█', nil, style)
						emitStr(s, w-8, h-3, white, fmt.Sprintf(lFmt, l))
						emitStr(s, w-8, h-2, white, fmt.Sprintf(uFmt, u))
						emitStr(s, w-8, h-1, white, fmt.Sprintf(vFmt, v))
					}
				}
			}
		}
	}

	/*
		for i = 1; i < 100; i++ {
			for j := 1; j < 50; j++ {
				l := (float64(i) * 3.6)
				u := .4
				v := ((float64(j) / 25) - 2) * -1
				c = colorful.Luv(l, u, v)
				log.Printf("LUV values: %.2f, %.2f, %.2f", l, u, v)
				//log.Printf("RGB values: %v, %v, %v", c.R, c.G, c.B)
				style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(int32(c.R), int32(c.G), int32(c.B)))
				//log.Printf("setcontent: %d", i)
				s.SetContent(int(i), j, '█', nil, style)
			}
		}
	*/
	log.Println("show")
	s.Show()
	Pause(s)
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func Pause(s tcell.Screen) string {
	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape {
				log.Println("Break!")
				return "break"
			} else if ev.Rune() != '0' {
				log.Println(ev.Rune())
				return ""
			}
		}
	}
}

func emitStr(s tcell.Screen, x, y int, style tcell.Style, str string) {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(x, y, c, comb, style)
		x += w
	}
}

func drawBox(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, r rune) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	for col := x1; col <= x2; col++ {
		s.SetContent(col, y1, tcell.RuneHLine, nil, style)
		s.SetContent(col, y2, tcell.RuneHLine, nil, style)
	}
	for row := y1 + 1; row < y2; row++ {
		s.SetContent(x1, row, tcell.RuneVLine, nil, style)
		s.SetContent(x2, row, tcell.RuneVLine, nil, style)
	}
	if y1 != y2 && x1 != x2 {
		// Only add corners if we need to
		s.SetContent(x1, y1, tcell.RuneULCorner, nil, style)
		s.SetContent(x2, y1, tcell.RuneURCorner, nil, style)
		s.SetContent(x1, y2, tcell.RuneLLCorner, nil, style)
		s.SetContent(x2, y2, tcell.RuneLRCorner, nil, style)
	}
	for row := y1 + 1; row < y2; row++ {
		for col := x1 + 1; col < x2; col++ {
			s.SetContent(col, row, r, nil, style)
		}
	}
}

func drawSelect(s tcell.Screen, x1, y1, x2, y2 int, sel bool) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}
	for row := y1; row <= y2; row++ {
		for col := x1; col <= x2; col++ {
			mainc, combc, style, width := s.GetContent(col, row)
			if style == tcell.StyleDefault {
				style = defStyle
			}
			style = style.Reverse(sel)
			s.SetContent(col, row, mainc, combc, style)
			col += width - 1
		}
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type RGB struct {
	r uint8
	g uint8
	b uint8
}

func dirInc(direction *uint8) {
	*direction += 1
	if *direction > 2 {
		*direction = 0
	}
}

func (dot *Dot) Inc(direction string) string {
	switch direction {
	case "up":
		for {
			dirInc(&dot.line.up)
			if dot.line.GetChar() != 'x' {
				break
			}
		}
	case "down":
		for {
			dirInc(&dot.line.down)
			if dot.line.GetChar() != 'x' {
				break
			}
		}
	case "left":
		for {
			dirInc(&dot.line.left)
			if dot.line.GetChar() != 'x' {
				break
			}
		}
	case "right":
		for {
			dirInc(&dot.line.right)
			if dot.line.GetChar() != 'x' {
				break
			}
		}
	}
	dot.gc = string(dot.line.GetChar())
	return dot.gc
}

type linestyle uint8

const (
	skinnyFat linestyle = iota
	dotted
	dashed
	singleDouble
	rounded
	styleEnd
)

type Line struct {
	style linestyle
	up    uint8
	down  uint8
	left  uint8
	right uint8
}

func (l Line) GetChar() rune {
	switch l.style {
	case skinnyFat:
		return charsets.SkinnyFatChars[l.up][l.down][l.left][l.right]
	case singleDouble:
		return charsets.SingleDoubleChars[l.up][l.down][l.left][l.right]
	default:
		return 'x'
	}
}

type Dot struct {
	fg   RGB
	bg   RGB
	line Line
	gc   string
}

type Page struct {
	dots [][]Dot
	maxX int
	maxY int
	adjX int
	adjY int
}

func (p *Page) GetDot(x, y int) *Dot {
	return &p.dots[x+p.adjX][y+p.adjY]
}

func (p Page) Shift(x, y int) {
	p.adjX += x
	if p.adjX > p.maxX {
		p.adjX = p.maxX
	}
	p.adjY += y
	if p.adjY > p.maxY {
		p.adjY = p.maxY
	}
}

func (p Page) Draw(s tcell.Screen) {

	w, h := s.Size()
	white := tcell.StyleDefault.
		Foreground(tcell.ColorWhite).Background(tcell.ColorRed)
	var printString string
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			printString = p.GetStr(x, y)
			if printString != "" {
				log.Printf("x: %d y: %d char: %s", x, y, printString)
				emitStr(s, x, y, white, printString)
			}
		}
	}

}

func (p Page) GetStr(x, y int) string {
	return string(p.dots[x+p.adjX][y+p.adjY].gc)

}

func main() {
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)

	log.Println("Hello world!")

	initX := 1000
	initY := 1000

	var page Page
	page.dots = make([][]Dot, initX)
	page.maxX = initX
	page.maxY = initY
	page.adjX = 500
	page.adjY = 500

	for i := range page.dots {
		page.dots[i] = make([]Dot, initY)
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		if runtime.GOOS == "windows" {
			shell = "CMD.EXE"
		} else {
			shell = "/bin/sh"
		}
	}

	encoding.Register()

	s, e := tcell.NewScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	if e := s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	defStyle = tcell.StyleDefault.
		Background(tcell.ColorReset).
		Foreground(tcell.ColorReset)
	s.SetStyle(defStyle)
	s.EnableMouse()
	s.EnablePaste()
	s.Clear()

	posfmt := "Mouse: %d, %d  "
	btnfmt := "Buttons: %s"
	keyfmt := "Keys: %s"
	//linefmt := "up: %d, down: %d, left: %d, right: %d"
	// pastefmt := "Paste: [%d] %s"
	white := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorRed)

	// mx, my := -1, -1
	ox, oy := -1, -1
	bx, by := -1, -1
	// Screen size
	w, h := s.Size()
	lchar := '*'
	bstr := ""
	lks := ""
	pstr := ""
	ecnt := 0
	pasting := false
	x, y := 0, 0

	for {

		drawBox(s, 1, 1, 42, 10, white, ' ')
		emitStr(s, 2, 2, white, "Press ESC twice to exit, C to clear.")
		// emitStr(s, 2, 3, white, fmt.Sprintf(posfmt, mx, my))
		// s.SetContent(x, y, GetChar(page[x][y].line), nil, white)
		emitStr(s, 2, 3, white, fmt.Sprintf(posfmt, x, y))
		emitStr(s, 2, 4, white, fmt.Sprintf(btnfmt, bstr))
		emitStr(s, 2, 5, white, fmt.Sprintf(keyfmt, lks))
		emitStr(s, 2, 6, white, fmt.Sprintf("char from array: '%s'", page.GetStr(x, y)))
		//emitStr(s, 2, 6, white, fmt.Sprintf(linefmt, page[x][y].line.up, page[x][y].line.down, page[x][y].line.left, page[x][y].line.right))
		//emitStr(s, 2, 6, white, fmt.Sprintf("RGB values: %.0f, %.0f, %.0f", c.R*256, c.G*256, c.B*256))

		//ps := pstr
		//if len(ps) > 26 {
		//	ps = "..." + ps[len(ps)-24:]
		//}
		//emitStr(s, 2, 6, white, fmt.Sprintf(pastefmt, len(pstr), ps))

		s.Show()
		bstr = ""
		ev := s.PollEvent()
		st := tcell.StyleDefault.Background(tcell.ColorRed)
		up := tcell.StyleDefault.
			Background(tcell.ColorBlue).
			Foreground(tcell.ColorBlack)
		w, h = s.Size()

		// always clear any old selection box
		if ox >= 0 && oy >= 0 && bx >= 0 {
			drawSelect(s, ox, oy, bx, by, false)
		}

		switch ev := ev.(type) {

		// Detect terminal resize
		case *tcell.EventResize:
			s.Sync()
			s.SetContent(w-1, h-1, 'R', nil, st)

		// KeypressG
		case *tcell.EventKey:
			s.SetContent(w-2, h-2, ev.Rune(), nil, st)
			if pasting {
				s.SetContent(w-1, h-1, 'P', nil, st)
				if ev.Key() == tcell.KeyRune {
					pstr = pstr + string(ev.Rune())
				} else {
					pstr = pstr + "\ufffd" // replacement for now
				}
				lks = ""
				continue
			}
			pstr = ""
			s.SetContent(w-1, h-1, 'K', nil, st)
			if ev.Key() == tcell.KeyEscape {
				ecnt++
				if ecnt > 1 {
					s.Fini()
					os.Exit(1)
				}
			} else if ev.Key() == tcell.KeyCtrlL {
				s.Sync()
			} else if ev.Key() == tcell.KeyCtrlZ {
				// CtrlZ doesn't really suspend the process, but we use it to execute a subshell.
				if err := s.Suspend(); err == nil {
					fmt.Printf("Executing shell (%s -l)...\n", shell)
					fmt.Printf("Exit the shell to return to the demo.\n")
					c := exec.Command(shell, "-l") // NB: -l works for cmd.exe too (ignored)
					c.Stdin = os.Stdin
					c.Stdout = os.Stdout
					c.Stderr = os.Stderr
					c.Run()
					if err := s.Resume(); err != nil {
						panic("failed to resume: " + err.Error())
					}
				}

				// Pressed Space
				//} else if ev.Key() == tcell.KeyCtrlC {
			} else if ev.Rune() == 'm' {
				ColorScreen(s)
				s.Clear()
				page.Draw(s)
			} else if ev.Rune() == ' ' {
				s.SetContent(x, y, 'X', nil, white)
				page.dots[x+page.adjX][y+page.adjY] = Dot{fg: RGB{255, 255, 0}, bg: RGB{0, 0, 0}, line: Line{}, gc: "X"}

				// Pressed Arrow Keys
			} else if ev.Key() == tcell.KeyUp {
				if y > 0 {
					y -= 1
				}
			} else if ev.Key() == tcell.KeyDown {
				y += 1
			} else if ev.Key() == tcell.KeyLeft {
				if x > 0 {
					x -= 1
				}
			} else if ev.Key() == tcell.KeyRight {
				x += 1

				// Pressed Directional line drawing keys
			} else if ev.Rune() == 'a' {
				emitStr(s, x, y, white, fmt.Sprint(page.GetDot(x, y).Inc("left")))
			} else if ev.Rune() == 'A' {
				if page.dots[x][y].line.left == 0 {
					page.dots[x][y].line.left = 2
				} else {
					page.dots[x][y].line.left -= 1
				}
				curChar := page.dots[x][y].line.GetChar()
				page.dots[x][y].gc = string(curChar)
				s.SetContent(x, y, curChar, nil, white)
			} else if ev.Rune() == 'w' {
				emitStr(s, x, y, white, fmt.Sprint(page.GetDot(x, y).Inc("up")))
			} else if ev.Rune() == 'W' {
				if page.dots[x][y].line.up == 0 {
					page.dots[x][y].line.up = 2
				} else {
					page.dots[x][y].line.up -= 1
				}
				curChar := page.dots[x][y].line.GetChar()
				page.dots[x][y].gc = string(curChar)
				s.SetContent(x, y, curChar, nil, white)
			} else if ev.Rune() == 'd' {
				emitStr(s, x, y, white, fmt.Sprint(page.GetDot(x, y).Inc("right")))
			} else if ev.Rune() == 'D' {
				if page.dots[x][y].line.right == 0 {
					page.dots[x][y].line.right = 2
				} else {
					page.dots[x][y].line.right -= 1
				}
				curChar := page.dots[x][y].line.GetChar()
				page.dots[x][y].gc = string(curChar)
				s.SetContent(x, y, curChar, nil, white)
			} else if ev.Rune() == 's' {
				emitStr(s, x, y, white, fmt.Sprint(page.GetDot(x, y).Inc("down")))
			} else if ev.Rune() == 'S' {
				if page.dots[x][y].line.down == 0 {
					page.dots[x][y].line.down = 2
				} else {
					page.dots[x][y].line.down -= 1
				}
				curChar := page.dots[x][y].line.GetChar()
				page.dots[x][y].gc = string(curChar)
				s.SetContent(x, y, curChar, nil, white)
			} else if ev.Rune() == 'p' {
				page.Draw(s)
			} else {
				ecnt = 0
				if ev.Rune() == 'C' || ev.Rune() == 'c' {
					s.Clear()
				}
			}
			lks = ev.Name()
		case *tcell.EventPaste:
			pasting = ev.Start()
			if pasting {
				pstr = ""
			}
		case *tcell.EventMouse:
			x, y = ev.Position()
			button := ev.Buttons()
			for i := uint(0); i < 8; i++ {
				if int(button)&(1<<i) != 0 {
					bstr += fmt.Sprintf(" Button%d", i+1)
				}
			}
			if button&tcell.WheelUp != 0 {
				bstr += " WheelUp"
			}
			if button&tcell.WheelDown != 0 {
				bstr += " WheelDown"
			}
			if button&tcell.WheelLeft != 0 {
				bstr += " WheelLeft"
			}
			if button&tcell.WheelRight != 0 {
				bstr += " WheelRight"
			}
			// Only buttons, not wheel events
			button &= tcell.ButtonMask(0xff)
			ch := '*'

			if button != tcell.ButtonNone && ox < 0 {
				ox, oy = x, y
			}
			switch ev.Buttons() {
			case tcell.ButtonNone:
				if ox >= 0 {
					bg := tcell.Color((lchar-'0')*2) | tcell.ColorValid
					drawBox(s, ox, oy, x, y,
						up.Background(bg),
						lchar)
					ox, oy = -1, -1
					bx, by = -1, -1
				}
			case tcell.Button1:
				ch = '1'
			case tcell.Button2:
				ch = '2'
			case tcell.Button3:
				ch = '3'
			case tcell.Button4:
				ch = '4'
			case tcell.Button5:
				ch = '5'
			case tcell.Button6:
				ch = '6'
			case tcell.Button7:
				ch = '7'
			case tcell.Button8:
				ch = '8'
			default:
				ch = '*'

			}
			if button != tcell.ButtonNone {
				bx, by = x, y
			}
			lchar = ch
			s.SetContent(w-1, h-1, 'M', nil, st)
			// mx, my = x, y
		default:
			s.SetContent(w-1, h-1, 'X', nil, st)
		}

		if ox >= 0 && bx >= 0 {
			drawSelect(s, ox, oy, bx, by, true)
		}
	}
}
