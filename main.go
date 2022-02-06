package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"io/ioutil"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	levelWidth                      = 64
	levelHeight                     = 16
	ScreenWidth                     = 256
	ScreenHeight                    = 200
	playerVerticalMoveAcc           = 0.06
	playerHorizontalMoveAccOnGround = 0.011
	playerHorizontalMoveAccInAir    = 0.005
	playerVerticalVelocityMax       = 1.0
	playerHorizontalDrag            = playerHorizontalMoveAccOnGround * 4
	playerJumpSpeed                 = playerVerticalMoveAcc * 5
	playerHorizontalVelocityMax     = 0.1
	clampHorizontalVelocityBelow    = 0.01
	gravity                         = 0.012
)

var cameraPosX = 0.0
var cameraPosY = 0.0

var playerPosX = 0.0
var playerPosY = 0.0
var playerVelX = 0.0
var playerVelY = 0.0

var playerOnGround bool
var playerFacingRight bool
var playerCurrentFrame image.Image
var coinCurrentFrame image.Image
var time int
var coinCounter int

var (
	yellow           = color.NRGBA{0xff, 0xff, 0x0, 0xff}
	red              = color.NRGBA{0xff, 0x0, 0x0, 0xff}
	lightBlue        = color.NRGBA{0x0, 0xff, 0xff, 0xff}
	green            = color.NRGBA{0x0, 0xff, 0x0, 0xff}
	greenTransparent = color.NRGBA{0x0, 150, 0x0, 250}
)

type Game struct {
	level []rune
}

func (g *Game) GetTile(x, y int) rune {
	if x >= 0 && x < levelWidth && y >= 0 && y < levelHeight {
		return g.level[y*levelWidth+x]
	} else {
		// log.Fatalf("could not get tile at position: (%v, %v)", x, y)
		return ' '
	}
}

func (g *Game) SetTile(x, y int, r rune) {
	if x >= 0 && x < levelWidth && y >= 0 && y < levelHeight {
		g.level[y*levelWidth+x] = r
	} else {
		log.Fatalf("could not set tile at position: (%v, %v)", x, y)
	}
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// playerVelY = 0
	// playerVelX = 0

	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		playerVelY += -playerVerticalMoveAcc
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		playerVelY += playerVerticalMoveAcc
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		playerFacingRight = false
		if playerOnGround {
			playerVelX += -playerHorizontalMoveAccOnGround
		} else {
			playerVelX += -playerHorizontalMoveAccInAir
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		playerFacingRight = true
		if playerOnGround {
			playerVelX += playerHorizontalMoveAccOnGround
		} else {
			playerVelX += playerHorizontalMoveAccInAir
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		if playerOnGround {
			playerVelY = -playerJumpSpeed
		}
	}

	playerVelY += gravity

	// drag
	if playerOnGround {
		playerVelX += -playerHorizontalDrag * playerVelX
		if math.Abs(playerVelX) < clampHorizontalVelocityBelow {
			playerVelX = 0
		}
	}

	newPlayerPosX := playerPosX + playerVelX
	newPlayerPosY := playerPosY + playerVelY

	// coin collection
	// left upper corner
	if g.GetTile(int(newPlayerPosX+0.0), int(newPlayerPosY+0.0)) == 'C' {
		coinCounter++
		fmt.Println(coinCounter)
		g.SetTile(int(newPlayerPosX+0.0), int(newPlayerPosY+0.0), '.')
	}
	// left lower corner
	if g.GetTile(int(newPlayerPosX+0.0), int(newPlayerPosY+0.9)) == 'C' {
		coinCounter++
		fmt.Println(coinCounter)
		g.SetTile(int(newPlayerPosX+0.0), int(newPlayerPosY+0.9), '.')
	}
	// right lower corner
	if g.GetTile(int(newPlayerPosX+1.0), int(newPlayerPosY+0.0)) == 'C' {
		coinCounter++
		fmt.Println(coinCounter)
		g.SetTile(int(newPlayerPosX+1.0), int(newPlayerPosY+0.0), '.')
	}
	// right upper corner
	if g.GetTile(int(newPlayerPosX+1.0), int(newPlayerPosY+0.9)) == 'C' {
		coinCounter++
		fmt.Println(coinCounter)
		g.SetTile(int(newPlayerPosX+1.0), int(newPlayerPosY+0.9), '.')
	}
	// Collission
	playerOnGround = false
	if playerVelX <= 0 { // going left
		if g.GetTile(int(newPlayerPosX+0.0), int(playerPosY+0.0)) != '.' || g.GetTile(int(newPlayerPosX+0.0), int(playerPosY+0.9)) != '.' {
			newPlayerPosX = float64(int(newPlayerPosX + 1))
			playerVelX = 0
		}
	} else if playerVelX > 0 {
		if g.GetTile(int(newPlayerPosX+1.0), int(playerPosY+0.0)) != '.' || g.GetTile(int(newPlayerPosX+1.0), int(playerPosY+0.9)) != '.' {
			newPlayerPosX = float64(int(newPlayerPosX))
			playerVelX = 0
		}
	}

	if playerVelY <= 0 {
		if g.GetTile(int(newPlayerPosX), int(newPlayerPosY+0.0)) != '.' || g.GetTile(int(newPlayerPosX+0.9), int(newPlayerPosY+0.0)) != '.' {
			newPlayerPosY = float64(int(newPlayerPosY) + 1)
			playerVelY = 0
		}
	} else if playerVelY > 0 {
		if g.GetTile(int(newPlayerPosX), int(newPlayerPosY+1.0)) != '.' || g.GetTile(int(newPlayerPosX+0.9), int(newPlayerPosY+1.0)) != '.' {
			newPlayerPosY = float64(int(newPlayerPosY))
			playerVelY = 0
			playerOnGround = true
		}
	}

	// clamp velocities
	if playerVelX > playerHorizontalVelocityMax {
		playerVelX = playerHorizontalVelocityMax
	}
	if playerVelX < -playerHorizontalVelocityMax {
		playerVelX = -playerHorizontalVelocityMax
	}
	if playerVelY > playerVerticalVelocityMax {
		playerVelY = playerVerticalVelocityMax
	}
	if playerVelY < -playerVerticalVelocityMax {
		playerVelY = -playerVerticalVelocityMax
	}

	// animation
	time++
	if playerOnGround && playerVelX != 0 {
		playerCurrentFrame = characterRunning[time/3%len(characterRunning)]
	} else if playerVelY != 0 && !playerOnGround {
		playerCurrentFrame = characterJumping
	} else {
		playerCurrentFrame = characterStanding
	}
	coinCurrentFrame = coin[time/3%len(coin)]

	playerPosX = newPlayerPosX
	playerPosY = newPlayerPosY

	cameraPosX = playerPosX
	cameraPosY = playerPosY
	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	var tileWidth = 16
	var tileHeight = 16
	visibleTilesX := ScreenWidth / tileWidth
	visibleTilesY := ScreenHeight / tileHeight

	// Calculate top-leftmost visible tile
	offsetX := cameraPosX - float64(visibleTilesX)/2.0
	offsetY := cameraPosY - float64(visibleTilesY)/2.0
	// Clamp camera close to the boundaries
	if offsetX < 0 {
		offsetX = 0
	}
	if offsetY < 0 {
		offsetY = 0
	}
	if offsetX > float64(levelWidth-visibleTilesX) {
		offsetX = float64(levelWidth - visibleTilesX)
	}
	if offsetY > float64(levelHeight-visibleTilesY) {
		offsetY = float64(levelHeight - visibleTilesY)
	}

	// Calculate tile offests for smooth movement (partial tiles to display)
	tileOffsetX := (offsetX - float64(int(offsetX))) * float64(tileWidth)
	tileOffsetY := (offsetY - float64(int(offsetY))) * float64(tileHeight)
	screen.Fill(lightBlue)
	// img := ebiten.NewImage(tileWidth, tileHeight)
	// Draw one tile more from left and right to avoid weird glitches on the edges
	for x := -1; x < visibleTilesX+1; x++ {
		for y := -1; y < visibleTilesY+1; y++ {
			tileID := g.GetTile(x+int(offsetX), y+int(offsetY))
			switch tileID {
			case '.':
				// empty space, do not draw as it should take background color
			case '#':
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(x*tileWidth)-tileOffsetX, float64(y*tileHeight)-tileOffsetY)
				screen.DrawImage(ebiten.NewImageFromImage(background1), op)
			case 'G':
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(x*tileWidth)-tileOffsetX, float64(y*tileHeight)-tileOffsetY)
				screen.DrawImage(ebiten.NewImageFromImage(ground1), op)
			case 'C':
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(x*tileWidth)-tileOffsetX, float64(y*tileHeight)-tileOffsetY)
				screen.DrawImage(ebiten.NewImageFromImage(coinCurrentFrame), op)
			default:
				// op := &ebiten.DrawImageOptions{}
				// op.GeoM.Translate(float64(x*tileWidth), float64(y*tileHeight))
				// img.Fill(yellow)
				// screen.DrawImage(img, op)

			}
		}
	}
	// Draw player
	op := &ebiten.DrawImageOptions{}
	if !playerFacingRight {
		// flip horizontally
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(float64(tileWidth), 0)
	}
	op.GeoM.Translate(float64(playerPosX-offsetX)*float64(tileWidth), float64(playerPosY-offsetY)*float64(tileHeight))
	// img.Fill(green)
	// img.Bounds()
	// screen.DrawImage(img, op)

	screen.DrawImage(ebiten.NewImageFromImage(playerCurrentFrame), op)
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

var charactersAtlas *ebiten.Image
var (
	characterStanding image.Image
	characterRunning  []image.Image
	characterJumping  image.Image
	background1       image.Image
	ground1           image.Image
	coin              []image.Image
)

func init() {
	spriteWidth := 16
	spriteHeight := 16
	fileContent, err := ioutil.ReadFile("./assets/pheasant.png")
	if err != nil {
		log.Fatal(err)
	}
	img, _, err := image.Decode(bytes.NewReader(fileContent))
	if err != nil {
		log.Fatal("failed to decode:", err)
	}
	charactersAtlas = ebiten.NewImageFromImage(img)
	characterStanding = charactersAtlas.SubImage(image.Rect(0, 0, spriteWidth, spriteHeight))
	characterRunning1 := charactersAtlas.SubImage(image.Rect(spriteWidth*1, 0, spriteWidth*2, spriteHeight))
	characterRunning2 := charactersAtlas.SubImage(image.Rect(spriteWidth*2, 0, spriteWidth*3, spriteHeight))
	characterRunning3 := charactersAtlas.SubImage(image.Rect(spriteWidth*3, 0, spriteWidth*4, spriteHeight))
	characterRunning4 := charactersAtlas.SubImage(image.Rect(spriteWidth*4, 0, spriteWidth*5, spriteHeight))
	characterJumping = charactersAtlas.SubImage(image.Rect(0, spriteHeight*1, spriteWidth, spriteHeight*2))
	characterRunning = []image.Image{
		characterRunning1,
		characterRunning2,
		characterRunning3,
		characterRunning4,
	}

	tilesFileContent, err := ioutil.ReadFile("./assets/level_tiles.png")
	if err != nil {
		log.Fatal(err)
	}
	tilesImg, _, err := image.Decode(bytes.NewReader(tilesFileContent))
	if err != nil {
		log.Fatal("failed to decode:", err)
	}
	levelTilesAtlas := ebiten.NewImageFromImage(tilesImg)
	background1 = levelTilesAtlas.SubImage(image.Rect(87, 351, 87+spriteWidth, 351+spriteHeight))
	ground1 = levelTilesAtlas.SubImage(image.Rect(161, 369, 161+spriteWidth, 369+spriteHeight))
	coin1 := levelTilesAtlas.SubImage(image.Rect(346, 351, 346+spriteWidth, 351+spriteHeight))
	coin2 := levelTilesAtlas.SubImage(image.Rect(346, 351+(spriteHeight+2), 346+spriteWidth, 351+(2*spriteHeight+2)))
	coin3 := levelTilesAtlas.SubImage(image.Rect(346, 351+2*(spriteHeight+2), 346+spriteWidth, 351+2*(spriteHeight+2)+spriteHeight))
	coin = []image.Image{coin1, coin2, coin3}
}

func main() {
	var levelString = "................................................................"
	levelString += "................................................................"
	levelString += "................................................................"
	levelString += "....CC.........................................................."
	levelString += ".....................................##........................."
	levelString += "....GG......#############..................#.#.................."
	levelString += "..........####.....................GG......#.#.................."
	levelString += ".........#####.......CC........................................."
	levelString += "##################################.###########...###############"
	levelString += ".................................#.#.............#.............."
	levelString += ".......................###########.#...........###.............."
	levelString += ".......................#...........#.......#####................"
	levelString += "............####.......#.###########....####...................."
	levelString += ".......................#..............####......................"
	levelString += ".......................################........................."
	levelString += "................................................................"
	levelString += "................................................................"
	game := &Game{
		level: []rune(levelString),
	}
	// Specify the window size as you like. Here, a doubled size is specified.
	ebiten.SetWindowSize(2*640, 2*480)
	ebiten.SetWindowTitle("Your game's title")
	// Call ebiten.RunGame to start your game loop.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
