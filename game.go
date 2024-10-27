package main

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	screenWidth  = 1000
	screenHeight = 480
)

var (
	running  = true
	bkgColor = rl.NewColor(147, 211, 196, 255)

	grassSprite  rl.Texture2D
	playerSprite rl.Texture2D

	playerSrc                                     rl.Rectangle
	playerDest                                    rl.Rectangle
	playerMoving                                  bool
	playerDir                                     int
	playerUp, playerDown, playerRight, playerLeft bool
	playerFrame                                   int

	frameCount int

	playerSpeed float32 = 3
)

func drawScene() {
	rl.DrawTexture(grassSprite, 100, 50, rl.White)
	rl.DrawTexturePro(playerSprite, playerSrc, playerDest, rl.NewVector2(playerDest.Width, playerDest.Height), 0, rl.White)
}

func input() {
	if rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyUp) {
		playerDest.Y -= playerSpeed
		playerMoving = true
		playerDir = 1
		playerUp = true
	}
	if rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyDown) {
		playerMoving = true
		playerDest.Y += playerSpeed
		playerDir = 0
		playerDown = true
	}
	if rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyLeft) {
		playerMoving = true
		playerDest.X -= playerSpeed
		playerDir = 2
		playerLeft = true
	}
	if rl.IsKeyDown(rl.KeyD) || rl.IsKeyDown(rl.KeyRight) {
		playerMoving = true
		playerDest.X += playerSpeed
		playerDir = 3
		playerRight = true
	}
}
func update() {
	running = !rl.WindowShouldClose()

	if playerMoving {
		if playerUp {
			playerDest.Y -= playerSpeed
		}
		if playerDown {
			playerDest.Y += playerSpeed
		}
		if playerRight {
			playerDest.X += playerSpeed
		}
		if playerLeft {
			playerDest.X -= playerSpeed
		}

		if frameCount%8 == 1 {
			playerFrame++
		}

	}
	frameCount++
	if playerFrame > 3 {
		playerFrame = 0
	}

	playerSrc.X = playerSrc.Width * float32(playerFrame)

	playerSrc.Y = playerSrc.Height * float32(playerDir)

}

func render() {
	rl.BeginDrawing()

	rl.ClearBackground(bkgColor)

	playerMoving = false
	playerUp, playerDown, playerRight, playerLeft = false, false, false, false

	drawScene()

	rl.EndDrawing()
}

func init() {
	rl.InitWindow(screenWidth, screenHeight, "Totoro")
	rl.SetExitKey(0)
	rl.SetTargetFPS(60)
	grassSprite = rl.LoadTexture("res/Tilesets/Grass.png")
	playerSprite = rl.LoadTexture("res/Characters/Basic Charakter Spritesheet.png")

	playerSrc = rl.NewRectangle(0, 0, 48, 48)
	playerDest = rl.NewRectangle(200, 200, 100, 100)

}
func quit() {
	rl.CloseWindow()
	rl.UnloadTexture(grassSprite)
	rl.UnloadTexture(playerSprite)
}

func main() {

	for running {
		input()
		update()
		render()
	}
	quit()
}
