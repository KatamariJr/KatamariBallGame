package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"image/gif"
	_ "image/gif"
	"image/png"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/atolVerderben/tentsuyu"
	"github.com/hajimehoshi/ebiten"
)

// Game implements ebiten.Game interface.
type Game struct {
	count  int
	Things []*Thing
}

type Thing struct {
	Name *tentsuyu.TextElement
	X    float64
	Y    float64
}

const (
	screenWidth  = 875
	screenHeight = 180
	layoutMult   = 1
	centerX      = screenWidth / 2
	serverURL    = "http://dev.katamarijr.com:4444/drain"
)

var (
	cowImage   *ebiten.Image
	rollGif    *gif.GIF
	rollImages []*ebiten.Image
	frameNum   int
	rollX      = float64(900)
	rollY      = float64(0)
	font       *truetype.Font
	gameText   *tentsuyu.TextElement
)

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update(screen *ebiten.Image) error {
	// Write your game's logical update.
	g.count++

	rollX--
	if rollX == -350 {
		rollX = 1600
		fmt.Println("respawned")
	}

	//check spacebar
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.SpawnThing("Bob")
	}

	//check collision
	for _, v := range g.Things {
		if v.X == rollX {
			g.DespawnThing(v)
		}
	}

	//maybe do network call
	if g.count%180 == 0 {
		err := g.NetworkCall()
		if err != nil {
			fmt.Printf("unable to make network call: %s\n", err.Error())
		}
	}

	return nil
}

type RequestType struct {
	Names []string
}

func (g *Game) NetworkCall() error {
	res, err := http.Get(serverURL)
	if err != nil {
		return err
	}

	var req RequestType

	err = json.NewDecoder(res.Body).Decode(&req)
	if err != nil {
		return err
	}

	for _, n := range req.Names {
		//truncate to 10char
		if len(n) > 10 {
			n = n[:10]
		}
		g.SpawnThing(n)
	}

	return nil
}

func (g *Game) SpawnThing(name string) {
	fmt.Println("spawned")
	t := &Thing{}

	t.X = float64(rand.Intn(900))
	t.Y = float64(rand.Intn(100) + 50)
	t.Name = tentsuyu.NewTextElement(t.X, t.Y-20, 150, 100, font, []string{name}, color.White, 24)
	g.Things = append(g.Things, t)
}

func (g *Game) DespawnThing(t *Thing) {
	//loop over thing array

	//remove thing from array

	for i, v := range g.Things {
		if v == t {
			g.Things = append(g.Things[:i], g.Things[i+1:]...)
		}
	}
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {

	screen.Fill(color.RGBA{
		R: 74,
		G: 198,
		B: 239,
		A: 255,
	})

	gameText.DrawApplyZoom(screen)

	for _, v := range g.Things {
		drawThing(screen, v)
	}

	drawBall(screen, g)

}

func drawThing(screen *ebiten.Image, v *Thing) {
	op := &ebiten.DrawImageOptions{}

	op.GeoM.Scale(0.2, 0.2)

	op.GeoM.Translate(v.X, v.Y)

	screen.DrawImage(cowImage, op)

	v.Name.DrawApplyZoom(screen)
}

func drawBall(screen *ebiten.Image, g *Game) {
	op := &ebiten.DrawImageOptions{}

	op.GeoM.Scale(0.65, 0.65)

	op.GeoM.Translate(rollX, rollY)

	screen.DrawImage(rollImages[(g.count/8)%frameNum], op)
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth * layoutMult, screenHeight * layoutMult
}

func NewGame() *Game {
	g := &Game{}
	g.Things = []*Thing{}
	return g
}

func main() {

	//open rolling gif
	f, err := os.Open("roll.gif")
	if err != nil {
		panic(err)
	}

	rollGif, err = gif.DecodeAll(f)
	if err != nil {
		panic(err)
	}

	for _, v := range rollGif.Image {

		frameNum++
		if frameNum == 1 {
			continue
		}

		img, err := ebiten.NewImageFromImage(v, ebiten.FilterDefault)
		if err != nil {
			panic(err)
		}
		rollImages = append(rollImages, img)
	}
	frameNum--

	//open cow img
	cow, err := os.Open("cow.png")
	if err != nil {
		panic(err)
	}
	cowPng, err := png.Decode(cow)
	if err != nil {
		panic(err)
	}
	cowImage, err = ebiten.NewImageFromImage(cowPng, ebiten.FilterDefault)
	if err != nil {
		panic(err)
	}

	//load font
	font, err = truetype.Parse(goregular.TTF)
	if err != nil {
		panic(err)
	}

	//load background text
	gameText = tentsuyu.NewTextElement(200, 25, 500, 250, font, []string{"-> dev.katamarijr.com <-"}, color.Gray{
		0,
	}, 50)

	rollX = centerX

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("katamari ball")
	ebiten.SetRunnableInBackground(true)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
