package main

import (
	"log"
	"os"

	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type InitSequenceBootLine struct {
	Text  string
	Delay int // ms before showing line
}

type Mode int

const (
	ModeSafe Mode = iota
	ModeDestruction
	ModeDanger
)

var modeNames = []string{"SAFE", "DESTRUCTION", "DANGER"}

var (
	mplusNormalFont font.Face
	hackerGreen     = color.RGBA{51, 255, 51, 255}
	lowGlowGreen    = color.RGBA{0, 50, 0, 255} // For that background "hum"

)

type GameState int

const (
	StateBooting GameState = iota
	StateMenu
	StateFSInit
	StatePlaying
	StateWon
	StateLoose
)

var bootSequence = []InitSequenceBootLine{
	{"TERMI WAR V1.0.0", 500},
	{"CORE-OS LOADING....", 850},
	{"INITALIZING GRAPHICS DRIVERS.....", 400},
	{"GRAPHICS: OK", 400},
	{"INITALIZING CPU......", 500},
	{"CPU: OK", 400},
	{"MOUNTING FILESYSTEM...", 600},
	{"SCANNING FOR NODES...", 1000},
	{"WARNING: DESTRUCTION MODE DETECTED IN KERNEL", 500},
}

type Game struct {
	state                   GameState
	inputActive             bool
	finalFilesystemPath     string
	inputBuffer             string
	currentMode             Mode
	bootIndex               int
	lastUpdate              time.Time
	bootSquenceVisibleLines []string
	terminalColor           color.RGBA
	lastInputTime           time.Time
}

func init() {
	// 1. Read the font file
	fontData, err := os.ReadFile("VT323-Regular.ttf")
	if err != nil {
		log.Fatal(err)
	}

	// 2. Parse the font
	tt, err := opentype.Parse(fontData)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Create a font face (Adjust size here)
	const dpi = 72
	mplusNormalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    30,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (g *Game) Update() error {
	switch g.state {
	case StateBooting:
		// If we haven't finished the sequence
		if g.bootIndex < len(bootSequence) {
			// Check if enough time has passed to show the next line
			if time.Since(g.lastUpdate).Milliseconds() > int64(bootSequence[g.bootIndex].Delay) {
				g.bootSquenceVisibleLines = append(g.bootSquenceVisibleLines, bootSequence[g.bootIndex].Text)
				g.bootIndex++
				g.lastUpdate = time.Now()
			}
		} else if time.Since(g.lastUpdate).Seconds() > 2 {
			// Wait 2 seconds after finishing, then clear and move to Menu
			g.state = StateMenu
			g.bootSquenceVisibleLines = []string{}
		}
	case StateMenu:
		if !g.inputActive {
			if ebiten.IsKeyPressed(ebiten.KeyRight) && time.Now().Sub(g.lastInputTime).Seconds() == 1 {
				g.lastInputTime = time.Now()
				g.currentMode = (g.currentMode + 1) % Mode(len(modeNames))
			}
			if ebiten.IsKeyPressed(ebiten.KeyLeft) && time.Now().Sub(g.lastInputTime).Seconds() == 1 {
				g.lastInputTime = time.Now()
				g.currentMode = (g.currentMode - 1 + Mode(len(modeNames))) % Mode(len(modeNames))
			}
			if ebiten.IsKeyPressed(ebiten.KeyEnter) {
				// If they pick DANGER or DESTRUCTION, you could trigger your warning here
				g.inputActive = true
			}
			return nil
		}

		// PHASE 2: Capturing Keyboard Input (Filtered)
		// Capture characters (skips arrows/enter/etc automatically)
		var b []rune
		b = ebiten.AppendInputChars(b)
		g.inputBuffer += string(b)

		// Manual handling for Backspace
		if ebiten.IsKeyPressed(ebiten.KeyBackspace) && len(g.inputBuffer) > 0 {
			g.inputBuffer = g.inputBuffer[:len(g.inputBuffer)-1]
		}

		// Handle Enter to finish directory input
		if ebiten.IsKeyPressed(ebiten.KeyEnter) {
			g.finalFilesystemPath = g.inputBuffer
			go initalizeFilesystem()
			g.state = StateFSInit
		}
	}

	return nil
}

func initalizeFilesystem() {

}

func (g *Game) Draw(screen *ebiten.Image) {
	// Fill background with a very dark green/black
	screen.Fill(color.RGBA{0, 5, 0, 255})

	// Draw lines in "Hacker Green"
	//green := color.RGBA{51, 255, 51, 255}
	for i, line := range g.bootSquenceVisibleLines {
		text.Draw(screen, line, mplusNormalFont, 20, 20+(i*30), hackerGreen)
	}

	if g.state == StateMenu {
		text.Draw(screen, "CHOOSE MODE: ", mplusNormalFont, 20, 20, hackerGreen)
		text.Draw(screen, "SAFE", mplusNormalFont, 150, 20, hackerGreen)
		text.Draw(screen, "DESTRUCTION", mplusNormalFont, 200, 20, hackerGreen)
		text.Draw(screen, "DANGER", mplusNormalFont, 325, 20, hackerGreen)
		ebitenutil.DebugPrintAt(screen, "ENTER GAME DIRECTORY: > _", 20, 40)
	}
	startX := 30
	for i, name := range modeNames {
		displayColor := color.RGBA{0, 100, 0, 255} // Dim green for inactive
		prefix := "  "
		suffix := "  "

		if Mode(i) == g.currentMode {
			displayColor = hackerGreen // Bright green
			prefix = "[ "
			suffix = " ]"
		}

		str := prefix + name + suffix
		text.Draw(screen, str, mplusNormalFont, startX, 100, displayColor)

		// Offset the next word based on string length
		startX += 200
	}

	// 2. Draw the Input Line
	if g.inputActive {
		prompt := "ENTER TARGET DIRECTORY: " + g.inputBuffer
		// Add a blinking cursor
		if (time.Now().UnixMilli()/500)%2 == 0 {
			prompt += "_"
		}
		text.Draw(screen, prompt, mplusNormalFont, 30, 200, hackerGreen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {

	return 1920, 1080
}

func main() {
	ebiten.SetWindowSize(1920, 1080)

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Termi-War")
	if err := ebiten.RunGame(&Game{
		state:         StateBooting,
		terminalColor: color.RGBA{51, 255, 51, 255},
		lastUpdate:    time.Now(),
		lastInputTime: time.Now(),
	}); err != nil {
		log.Fatal(err)
	}
}
