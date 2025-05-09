package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 1920 // Increased from 1000 to 1920 (standard HD width)
	screenHeight = 1080 // Increased from 480 to 1080 (standard HD height)
)

var (
	running  = true
	bkgColor = rl.NewColor(147, 211, 196, 255)

	grassSprite     rl.Texture2D
	groundSprite    rl.Texture2D
	playerSprite    rl.Texture2D
	nestSprite      rl.Texture2D
	creatureSprite  rl.Texture2D
	stoneTileSprite rl.Texture2D
	pineConeSprite  rl.Texture2D

	playerSrc                                     rl.Rectangle
	playerDest                                    rl.Rectangle
	playerMoving                                  bool
	playerDir                                     int
	playerUp, playerDown, playerRight, playerLeft bool
	playerFrame                                   int

	frameCount int

	playerSpeed float32 = 3

	droppedPineCones []rl.Vector2

	pineTreeSprite rl.Texture2D
	growingTrees   []struct {
		position rl.Vector2
		frame    int
		growing  bool
	}
	treeAnimationSpeed int = 60 // Made even slower for more visible growth

	camera rl.Camera2D // Add camera variable

	pineConeCount int // Number of pine cones in inventory

	bagBgSprite        rl.Texture2D
	pineConeIconSprite rl.Texture2D
)

type ItemType int

const (
	ItemNone ItemType = iota
	ItemPineCone
	// Add more item types here as needed
)

type InventorySlot struct {
	Item  ItemType
	Count int
}

var inventory [4]InventorySlot

func drawScene() {
	// Calculate the visible area based on camera position
	visibleMinX := int32(camera.Target.X - float32(screenWidth)/2/camera.Zoom)
	visibleMinY := int32(camera.Target.Y - float32(screenHeight)/2/camera.Zoom)
	visibleMaxX := int32(camera.Target.X + float32(screenWidth)/2/camera.Zoom)
	visibleMaxY := int32(camera.Target.Y + float32(screenHeight)/2/camera.Zoom)

	// Draw the ground tiles only in visible area
	tileWidth := float32(groundSprite.Width)
	tileHeight := float32(groundSprite.Height)

	startX := (visibleMinX / int32(tileWidth)) * int32(tileWidth)
	startY := (visibleMinY / int32(tileHeight)) * int32(tileHeight)
	endX := visibleMaxX + int32(tileWidth)
	endY := visibleMaxY + int32(tileHeight)

	for y := float32(startY); float32(y) < float32(endY); y += tileHeight {
		for x := float32(startX); float32(x) < float32(endX); x += tileWidth {
			rl.DrawTexture(groundSprite, int32(x), int32(y), rl.White)
		}
	}

	stoneTileSize := float32(stoneTileSprite.Width)
	rl.DrawTexture(stoneTileSprite, 100, 100, rl.White)
	rl.DrawTexture(stoneTileSprite, int32(100+stoneTileSize), 100, rl.White)
	rl.DrawTexture(stoneTileSprite, 100, int32(100+stoneTileSize), rl.White)

	rl.DrawTexture(nestSprite, 300, 200, rl.White)

	creatureX := screenWidth - float32(creatureSprite.Width) - 20 // 20 pixels padding from right
	creatureY := 20                                               // 20 pixels padding from top
	rl.DrawTexture(creatureSprite, int32(creatureX), int32(creatureY), rl.White)

	for _, pos := range droppedPineCones {
		rl.DrawTexture(pineConeSprite, int32(pos.X)-pineConeSprite.Width/2, int32(pos.Y)-pineConeSprite.Height/2, rl.White)
		rl.DrawCircle(int32(pos.X), int32(pos.Y), 5, rl.Blue) // Debug: cone center
	}

	// Draw all trees (growing and fully grown)
	for _, tree := range growingTrees {
		treeHeight := float32(pineTreeSprite.Height)
		treeWidth := float32(pineTreeSprite.Width) / 4 // Assuming 4 frames in sprite sheet

		// Calculate how much of the tree to show based on growth frame
		growthProgress := float32(tree.frame+1) / 4.0 // Will go from 0.25 to 1.0
		visibleHeight := treeHeight * growthProgress

		// Source rectangle (full width of one frame, but growing in height from bottom)
		treeSrc := rl.NewRectangle(
			float32(tree.frame)*treeWidth, // X position in sprite sheet
			treeHeight-visibleHeight,      // Start from bottom
			treeWidth,                     // Full width of one frame
			visibleHeight,                 // Only show the growing portion
		)

		// Destination rectangle (grows upward from the ground position)
		treeDest := rl.NewRectangle(
			tree.position.X-treeWidth/2,   // Center horizontally
			tree.position.Y-visibleHeight, // Position from bottom
			treeWidth,                     // Same width as source
			visibleHeight,                 // Same height as visible portion
		)

		rl.DrawTexturePro(pineTreeSprite, treeSrc, treeDest, rl.Vector2{}, 0, rl.White)
	}

	rl.DrawTexturePro(playerSprite, playerSrc, playerDest, rl.NewVector2(playerDest.Width, playerDest.Height), 0, rl.White)
	rl.DrawCircle(int32(playerDest.X+playerDest.Width/2), int32(playerDest.Y+playerDest.Height/2), 5, rl.Red) // Debug: player center
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

	if rl.IsKeyPressed(rl.KeySpace) {
		dropPineCone()
		fmt.Println("Pine cone dropped!")
	}

	if rl.IsKeyPressed(rl.KeyG) {
		fmt.Println("G key pressed!")
		if onCone, conePos := isPlayerOnPineCone(); onCone {
			fmt.Println("Standing on pine cone! Starting tree growth at:", conePos)
			growingTrees = append(growingTrees, struct {
				position rl.Vector2
				frame    int
				growing  bool
			}{
				position: conePos,
				frame:    0,
				growing:  true,
			})
		} else {
			fmt.Println("Not standing on any pine cone")
		}
	}

	if rl.IsKeyPressed(rl.KeyV) {
		fmt.Println("V key pressed!")
		pickUpPineCone()
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

	// Update camera to follow player
	camera.Target = rl.Vector2{
		X: playerDest.X + playerDest.Width/2,
		Y: playerDest.Y + playerDest.Height/2,
	}

	// Smooth camera following
	const smoothness float32 = 0.1
	camera.Target.X = camera.Target.X + (playerDest.X+playerDest.Width/2-camera.Target.X)*smoothness
	camera.Target.Y = camera.Target.Y + (playerDest.Y+playerDest.Height/2-camera.Target.Y)*smoothness

	updateTrees()
}

func render() {
	rl.BeginDrawing()

	rl.ClearBackground(bkgColor)

	// Begin camera mode before drawing scene
	rl.BeginMode2D(camera)

	playerMoving = false
	playerUp, playerDown, playerRight, playerLeft = false, false, false, false

	drawScene()

	rl.EndMode2D() // End camera mode

	// Draw the backpack background at the bottom center of the screen
	bagX := 100                                                       // distance from the left side
	bagY := float32(screenHeight) - float32(bagBgSprite.Height) - 420 // higher up

	// Apply scaling factor to make the bag smaller (0.7 = 70% of original size)
	scaleFactor := float32(0.7)
	scaledBagWidth := float32(bagBgSprite.Width) * scaleFactor
	scaledBagHeight := float32(bagBgSprite.Height) * scaleFactor

	// Create source and destination rectangles for scaled drawing
	bagSrc := rl.NewRectangle(0, 0, float32(bagBgSprite.Width), float32(bagBgSprite.Height))
	bagDest := rl.NewRectangle(float32(bagX), float32(bagY), scaledBagWidth, scaledBagHeight)

	// Draw the bag with scaling
	rl.DrawTexturePro(bagBgSprite, bagSrc, bagDest, rl.Vector2{}, 0, rl.White)

	// Draw the inventory slots and items - adjust for new scale
	slotSize := scaledBagWidth / 4
	for i, slot := range inventory {
		// Adjust item positions according to the new scale
		slotX := int32(bagX) + int32(float32(i)*slotSize) + int32(slotSize/3) - int32(float32(pineConeIconSprite.Width)*scaleFactor/3)
		slotY := int32(bagY) + int32(scaledBagHeight/2) - int32(float32(pineConeIconSprite.Height)*scaleFactor/4)

		if slot.Item == ItemPineCone && slot.Count > 0 {
			// Draw the pinecone icon with the same scale factor
			iconSrc := rl.NewRectangle(0, 0, float32(pineConeIconSprite.Width), float32(pineConeIconSprite.Height))
			iconDest := rl.NewRectangle(float32(slotX), float32(slotY),
				float32(pineConeIconSprite.Width)*scaleFactor,
				float32(pineConeIconSprite.Height)*scaleFactor)

			rl.DrawTexturePro(pineConeIconSprite, iconSrc, iconDest, rl.Vector2{}, 0, rl.White)

			// Adjust text position and size
			textSize := int32(20 * scaleFactor)
			textX := slotX + int32(32*scaleFactor)
			textY := slotY + int32(32*scaleFactor)
			rl.DrawText(fmt.Sprintf("%d", slot.Count), textX, textY, textSize, rl.Black)
		}
		// Add more item types here as you add them
	}

	rl.DrawText(fmt.Sprintf("Pine Cones: %d", pineConeCount), 20, 20, 30, rl.Black)

	rl.EndDrawing()
}

func init() {
	rl.InitWindow(screenWidth, screenHeight, "Totoro")
	rl.SetExitKey(0)
	rl.SetTargetFPS(60)

	groundSprite = rl.LoadTexture("res/Tilesets/ground.png")
	playerSprite = rl.LoadTexture("res/Characters/Basic Charakter Spritesheet.png")
	nestSprite = rl.LoadTexture("res/Tilesets/nest.png")
	creatureSprite = rl.LoadTexture("res/Tilesets/creature.png")
	stoneTileSprite = rl.LoadTexture("res/Tilesets/stone_tiles.png")
	pineConeSprite = rl.LoadTexture("res/Objects/pine_cone.png")

	playerSrc = rl.NewRectangle(0, 0, 48, 48)
	playerDest = rl.NewRectangle(200, 200, 100, 100)

	droppedPineCones = make([]rl.Vector2, 0)
	pineConeCount = 5 // Start with 5 pinecones in inventory

	pineTreeSprite = rl.LoadTexture("res/Objects/pine_tree_growth.png") // Make sure to add this sprite
	growingTrees = make([]struct {
		position rl.Vector2
		frame    int
		growing  bool
	}, 0)

	// Initialize camera
	camera = rl.Camera2D{
		Target:   rl.Vector2{X: playerDest.X + playerDest.Width/2, Y: playerDest.Y + playerDest.Height/2},
		Offset:   rl.Vector2{X: float32(screenWidth) / 2, Y: float32(screenHeight) / 2},
		Rotation: 0,
		Zoom:     1.0,
	}

	bagBgSprite = rl.LoadTexture("res/UI/bag_bg.png")
	pineConeIconSprite = rl.LoadTexture("res/UI/pinecone_icon.png")

	// Initialize inventory
	for i := range inventory {
		inventory[i].Item = ItemNone
		inventory[i].Count = 0
	}

	// Set initial inventory
	updateInventory()
}

func quit() {
	rl.CloseWindow()
	rl.UnloadTexture(groundSprite)
	rl.UnloadTexture(playerSprite)
	rl.UnloadTexture(nestSprite)
	rl.UnloadTexture(creatureSprite)
	rl.UnloadTexture(stoneTileSprite)
	rl.UnloadTexture(pineConeSprite)
	rl.UnloadTexture(pineTreeSprite)
	rl.UnloadTexture(bagBgSprite)
	rl.UnloadTexture(pineConeIconSprite)
}

func updateInventory() {
	// Clear inventory first
	for i := range inventory {
		inventory[i].Item = ItemNone
		inventory[i].Count = 0
	}

	// Add pine cones to first slot
	if pineConeCount > 0 {
		inventory[0].Item = ItemPineCone
		inventory[0].Count = pineConeCount
	}

	// Add other items to other slots as needed
	// For example:
	// if waterCount > 0 {
	//     inventory[1].Item = ItemWater
	//     inventory[1].Count = waterCount
	// }
}

func dropPineCone() {
	// Only drop if we have pinecones in inventory
	if pineConeCount <= 0 {
		fmt.Println("No pine cones to drop!")
		return
	}

	// Get player center
	playerCenter := rl.Vector2{
		X: playerDest.X + playerDest.Width/2,
		Y: playerDest.Y + playerDest.Height/2,
	}

	// Drop offset based on the direction the player is facing
	var offsetX, offsetY float32

	// Use playerDir to determine drop position
	switch playerDir {
	case 0: // Down
		offsetX = 0
		offsetY = playerDest.Height/2 + 30 // Drop in front of player
	case 1: // Up
		offsetX = 0
		offsetY = -playerDest.Height/2 - 30 // Drop above player
	case 2: // Left
		offsetX = -playerDest.Width/2 - 30 // Drop to left of player
		offsetY = 0
	case 3: // Right
		offsetX = playerDest.Width/2 + 30 // Drop to right of player
		offsetY = 0
	}

	// Calculate the dropping position
	pineConePos := rl.Vector2{
		X: playerCenter.X + offsetX,
		Y: playerCenter.Y + offsetY,
	}

	droppedPineCones = append(droppedPineCones, pineConePos)
	pineConeCount-- // Decrease inventory count

	// Update the inventory UI
	updateInventory()

	fmt.Printf("Dropped pine cone at: %v (facing direction: %d)\n", pineConePos, playerDir)
}

func isPlayerOnPineCone() (bool, rl.Vector2) {
	playerCenter := rl.Vector2{
		X: playerDest.X + playerDest.Width/2,
		Y: playerDest.Y + playerDest.Height/2,
	}

	for i, cone := range droppedPineCones {
		// Calculate distance between player and pine cone
		distance := float32(
			math.Sqrt(
				float64(
					(playerCenter.X-cone.X)*(playerCenter.X-cone.X) +
						(playerCenter.Y-cone.Y)*(playerCenter.Y-cone.Y),
				),
			),
		)

		fmt.Printf("Distance to cone: %f\n", distance)

		// Increased interaction radius to 150 pixels
		if distance < 150 {
			fmt.Println("Successfully interacted with pine cone!")
			droppedPineCones = append(droppedPineCones[:i], droppedPineCones[i+1:]...)
			return true, cone
		}
	}
	return false, rl.Vector2{}
}

func updateTrees() {
	for i := range growingTrees {
		if growingTrees[i].growing {
			if frameCount%treeAnimationSpeed == 0 {
				growingTrees[i].frame++
				fmt.Printf("Tree %d animation frame: %d\n", i, growingTrees[i].frame)
				if growingTrees[i].frame >= 4 {
					growingTrees[i].growing = false
					growingTrees[i].frame = 3 // Keep final frame
					fmt.Printf("Tree %d finished growing\n", i)
				}
			}
		}
	}
}

func drawDebug() {
	// Draw player center point
	playerCenter := rl.Vector2{
		X: playerDest.X + playerDest.Width/2,
		Y: playerDest.Y + playerDest.Height/2,
	}
	rl.DrawCircle(int32(playerCenter.X), int32(playerCenter.Y), 3, rl.Red)

	// Draw interaction radius around pine cones (increased to 150)
	for _, cone := range droppedPineCones {
		rl.DrawCircleLines(int32(cone.X), int32(cone.Y), 150, rl.Green)
	}
}

func pickUpPineCone() {
	playerCenter := rl.Vector2{
		X: playerDest.X + playerDest.Width/2,
		Y: playerDest.Y + playerDest.Height/2,
	}

	// Log the player position for debugging
	fmt.Printf("Player center position: %v\n", playerCenter)

	for i, cone := range droppedPineCones {
		// Log each cone position
		fmt.Printf("Checking cone at position: %v\n", cone)

		distance := float32(math.Hypot(float64(playerCenter.X-cone.X), float64(playerCenter.Y-cone.Y)))
		fmt.Printf("Distance to cone: %f\n", distance)

		// Increased pickup radius to match the interaction radius from isPlayerOnPineCone
		if distance < 150 {
			fmt.Println("Pine cone picked up!")
			droppedPineCones = append(droppedPineCones[:i], droppedPineCones[i+1:]...)
			pineConeCount++

			// Update the inventory UI
			updateInventory()

			return // Added return to prevent checking other cones after picking one up
		}
	}
	fmt.Println("No pine cone in range to pick up")
}

func main() {
	for running {
		input()
		update()
		render()
	}
	quit()
}
