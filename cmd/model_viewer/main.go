package main

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"stone-tools/lib"
	"unsafe"

	tga "github.com/davehouse/go-targa"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type GraphicsMode int

func (m GraphicsMode) String() string {
	switch m {
	case GraphicsModeWireFrame:
		return "Wireframe"
	case GraphicsModeSolid:
		return "Solid"
	case GraphicsModeTextured:
		return "Textured"
	}

	return "Unknown"
}

const (
	GraphicsModeWireFrame = 0
	GraphicsModeSolid     = 1
	GraphicsModeTextured  = 2
)

func loadTexture(texturePath string) (rl.Texture2D, error) {
	/*
		As of today (2025-03-16), the raylib-go tga parser does not seem to fully appreaciate the targa files packed with Darkstone,
			as such, this is a nasty worky around to load the TGA with a different library then re-save it as a png
	*/
	tgaFile, err := os.Open(texturePath)
	if err != nil {
		return rl.Texture2D{}, err
	}
	defer tgaFile.Close()

	tgaImage, err := tga.Decode(tgaFile)
	if err != nil {
		return rl.Texture2D{}, err
	}

	pngFile, err := os.CreateTemp(".", "*.png")
	if err != nil {
		return rl.Texture2D{}, err
	}
	defer func() {
		// close and clean-up temp file
		pngFile.Close()
		os.Remove(pngFile.Name())
	}()

	err = png.Encode(pngFile, tgaImage)
	if err != nil {
		pngFile.Close()
		return rl.Texture2D{}, err
	}

	texture := rl.LoadTexture(pngFile.Name())
	return texture, nil
}

func main() {
	// Initialization
	//--------------------------------------------------------------------------------------
	const screenWidth = 800
	const screenHeight = 450

	// TODO don't hardcode this...
	// o3dFile, err := os.Open(filepath.Join("..", "..", "out", "data", "DATA", "PROJECTILE", "DAGUE2.O3D"))
	o3dFile, err := os.Open(filepath.Join("..", "..", "out", "data", "DATA", "COMMON", "MESHES", "TORCHE.O3D"))
	if err != nil {
		panic(err)
	}

	o3dModel, err := lib.ExtractO3D(o3dFile)
	o3dFile.Close() // ensure closed NOW
	if err != nil {
		panic(err)
	}
	rl.InitWindow(screenWidth, screenHeight, "Stone Model Viewer")

	// TODO supports mod directories, multi-texture models, and lower-res R textures
	textureSearchPath := filepath.Join("..", "..", "out", "data", "DATA", "BANKDATABASE", "DRAGONBLADE", fmt.Sprintf("K%04d*.TGA", o3dModel.Faces[0].MaterialId))
	files, err := filepath.Glob(textureSearchPath)
	if err != nil {
		panic(err)
	} else if len(files) <= 0 {
		panic("could not find a texture file...")
	}

	// texture := rl.LoadTexture(files[0])
	texture, err := loadTexture(files[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rl.UnloadTexture(texture)

	// Define the camera to look into our 3d world
	var distance float32 = 60
	var yaw float32 = 300.0
	var pitch float32 = 34.0
	var graphicsMode GraphicsMode = GraphicsModeWireFrame
	wantsToChangeMode := false
	var isGridOn = true
	wantsToChangeGrid := false

	var camera rl.Camera
	camera.Position = rl.Vector3{
		X: 0.0,
		Y: 10.0,
		Z: 10.0,
	}
	camera.Target = rl.Vector3{
		X: 0.0,
		Y: 0.0,
		Z: 0.0,
	}
	camera.Up = rl.Vector3{
		X: 0.0,
		Y: 1.0,
		Z: 0.0,
	}
	camera.Fovy = 45.0
	camera.Projection = rl.CameraPerspective

	vertices := make([]float32, 0)
	texcoords := make([]float32, 0)
	colors := make([]uint8, 0)
	for _, face := range o3dModel.Faces {
		vertex0 := o3dModel.Vertices[face.V0]
		vertex1 := o3dModel.Vertices[face.V1]
		vertex2 := o3dModel.Vertices[face.V2]

		vertices = append(
			vertices,
			vertex0.X, vertex0.Y, vertex0.Z,
			vertex1.X, vertex1.Y, vertex1.Z,
			vertex2.X, vertex2.Y, vertex2.Z,
		)
		texcoords = append(
			texcoords,
			255.0/face.Tx0, 255.0/face.Ty0,
			255.0/face.Tx1, 255.0/face.Ty1,
			255.0/face.Tx2, 255.0/face.Ty2,
		)
		colors = append(
			colors,
			face.MaybeBlue, face.MaybeGreen, face.MaybeRed, 255,
			face.MaybeBlue, face.MaybeGreen, face.MaybeRed, 255,
			face.MaybeBlue, face.MaybeGreen, face.MaybeRed, 255,
		)

		if face.V3 != lib.O3DUnused {
			vertex3 := o3dModel.Vertices[face.V3]
			vertices = append(
				vertices,
				vertex2.X, vertex2.Y, vertex2.Z,
				vertex3.X, vertex3.Y, vertex3.Z,
				vertex0.X, vertex0.Y, vertex0.Z,
			)
			texcoords = append(
				texcoords,
				255.0/face.Tx2, 255.0/face.Ty2,
				255.0/face.Tx3, 255.0/face.Ty3,
				255.0/face.Tx0, 255.0/face.Ty0,
			)
			colors = append(
				colors,
				face.MaybeBlue, face.MaybeGreen, face.MaybeRed, 255,
				face.MaybeBlue, face.MaybeGreen, face.MaybeRed, 255,
				face.MaybeBlue, face.MaybeGreen, face.MaybeRed, 255,
			)
		}
	}

	var mesh rl.Mesh
	mesh.TriangleCount = int32(len(vertices) / 3)
	mesh.VertexCount = int32(len(vertices))
	mesh.Vertices = (*float32)(unsafe.Pointer(&vertices[0]))
	mesh.Texcoords = (*float32)(unsafe.Pointer(&texcoords[0]))
	mesh.Colors = (*uint8)(unsafe.Pointer(&colors[0]))

	rl.UploadMesh(&mesh, false)
	theModel := rl.LoadModelFromMesh(mesh)
	if texture.ID > 0 {
		// only apply texture if has a valid id
		theModel.Materials.Maps.Texture = texture
	}
	defer rl.UnloadModel(theModel)

	// run at 60 fps, close down window in finish
	rl.SetTargetFPS(60)
	defer rl.CloseWindow()
	for !rl.WindowShouldClose() { // if window closed or esc
		// handle input
		if rl.IsKeyDown(rl.KeyQ) {
			distance += 1
		} else if rl.IsKeyDown(rl.KeyE) {
			if distance-1 > 2 {
				distance -= 1
			}
		}
		if rl.IsKeyDown(rl.KeyA) {
			yaw += 1

			// wrap to keep numbers positive
			if yaw > 360 {
				yaw -= 360
			}
		} else if rl.IsKeyDown(rl.KeyD) {
			yaw -= 1

			// wrap to keep numbers positive
			if yaw < 0 {
				yaw += 360
			}
		}
		if rl.IsKeyDown(rl.KeyS) {
			if pitch-1 > 280 || pitch-1 < 270 {
				// keep at reasonable angles
				pitch -= 1
			}

			// wrap to keep numbers positive
			if pitch < 0 {
				pitch += 360
			}
		} else if rl.IsKeyDown(rl.KeyW) {
			if pitch+1 < 80 || pitch+1 > 90 {
				// keep at reasonable angles
				pitch += 1
			}

			// wrap to keep numbers positive
			if pitch > 360 {
				pitch -= 360
			}
		}

		if rl.IsKeyDown(rl.KeyUp) {
			camera.Target.Y += 0.5
		} else if rl.IsKeyDown(rl.KeyDown) {
			camera.Target.Y -= 0.5
		}

		if rl.IsKeyDown(rl.KeyC) {
			camera.Target.Y = 0
		}

		if wantsToChangeMode && rl.IsKeyUp(rl.KeyZ) {
			wantsToChangeMode = false
			graphicsMode = (graphicsMode + 1) % 3
		} else if rl.IsKeyDown(rl.KeyZ) {
			wantsToChangeMode = true
		}

		if wantsToChangeGrid && rl.IsKeyUp(rl.KeyG) {
			wantsToChangeGrid = false
			isGridOn = !isGridOn
		} else if rl.IsKeyDown(rl.KeyG) {
			wantsToChangeGrid = true
		}

		// put camera into correct orbit
		camera.Position = rl.Vector3Add(camera.Target, rl.Vector3Transform(
			rl.Vector3{X: 0, Y: 0, Z: distance},
			rl.MatrixMultiply(rl.MatrixRotateX(rl.Deg2rad*pitch), rl.MatrixRotateY(rl.Deg2rad*yaw)),
		))

		// rendering
		rl.BeginDrawing()

		rl.ClearBackground(rl.Black)

		rl.BeginMode3D(camera)

		if graphicsMode == GraphicsModeWireFrame {
			rl.DrawModelWires(theModel, rl.Vector3Zero(), 1.0, rl.White)
		} else {
			rl.DrawModel(theModel, rl.Vector3Zero(), 1.0, rl.White)
		}

		if isGridOn {
			rl.DrawGrid(20, 5.0)
		}

		rl.DrawLine3D(rl.Vector3{X: 0, Y: 0, Z: 0}, rl.Vector3{X: 10, Y: 0, Z: 0}, rl.Red)
		rl.DrawLine3D(rl.Vector3{X: 0, Y: 0, Z: 0}, rl.Vector3{X: 0, Y: 10, Z: 0}, rl.Green)
		rl.DrawLine3D(rl.Vector3{X: 0, Y: 0, Z: 0}, rl.Vector3{X: 0, Y: 0, Z: 10}, rl.Blue)

		rl.EndMode3D()
		rl.DrawText(fmt.Sprintf("Graphics Mode (Z): %s\nGrid On (G): %t", graphicsMode.String(), isGridOn), screenWidth-260, 10, 18, rl.Yellow)
		rl.DrawText(fmt.Sprintf("Yaw (A/D): %f\nPitch (W/S): %f\nDistance (Q/E): %f", yaw, pitch, distance), 10, 10, 18, rl.Green)
		rl.DrawFPS(10, 80)
		rl.EndDrawing()
		//----------------------------------------------------------------------------------
	}
}
