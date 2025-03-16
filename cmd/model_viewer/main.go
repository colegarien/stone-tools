package main

import (
	"fmt"
	"os"
	"path/filepath"
	"stone-tools/lib"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	// Initialization
	//--------------------------------------------------------------------------------------
	const screenWidth = 800
	const screenHeight = 450

	// TODO don't hardcode this...
	o3dFile, err := os.Open(filepath.Join("..", "..", "out", "data", "DATA", "PROJECTILE", "DAGUE2.O3D"))
	if err != nil {
		panic(err)
	}

	o3dModel, err := lib.ExtractO3D(o3dFile)
	o3dFile.Close() // ensure closed NOW
	if err != nil {
		panic(err)
	}

	rl.InitWindow(screenWidth, screenHeight, "Stone Model Viewer")

	// Define the camera to look into our 3d world
	var distance float32 = 60
	var yaw float32 = 300.0
	var pitch float32 = 34.0
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

	var mesh rl.Mesh
	mesh.TriangleCount = 1
	mesh.VertexCount = mesh.TriangleCount * 3

	vertices := []float32{
		0, 0, 0,
		1, 0, 2,
		2, 0, 0,
	}
	texcoords := []float32{
		0, 0,
		0.5, 1,
		1, 0,
	}
	normals := []float32{
		0, 1, 0,
		0, 1, 0,
		0, 1, 0,
	}
	mesh.Vertices = (*float32)(unsafe.Pointer(&vertices[0]))
	mesh.Texcoords = (*float32)(unsafe.Pointer(&texcoords[0]))
	mesh.Normals = (*float32)(unsafe.Pointer(&normals[0]))

	rl.UploadMesh(&mesh, false)
	theModel := rl.LoadModelFromMesh(mesh)
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
				pitch -= 1
			}

			// wrap to keep numbers positive
			if pitch < 0 {
				pitch += 360
			}
		} else if rl.IsKeyDown(rl.KeyW) {
			if pitch+1 < 80 || pitch+1 > 90 {
				pitch += 1
			}

			// wrap to keep numbers positive
			if pitch > 360 {
				pitch -= 360
			}
		}

		// put camera into correct orbit
		camera.Position = rl.Vector3Transform(
			rl.Vector3{X: 0, Y: 0, Z: distance},
			rl.MatrixMultiply(rl.MatrixRotateX(rl.Deg2rad*pitch), rl.MatrixRotateY(rl.Deg2rad*yaw)),
		)

		// rendering
		rl.BeginDrawing()

		rl.ClearBackground(rl.Black)

		rl.BeginMode3D(camera)

		rl.DrawModel(theModel, rl.Vector3Zero(), 1.0, rl.White)

		// --- Wireframe O3D Model ---
		for _, face := range o3dModel.Faces {
			vertex0 := o3dModel.Vertices[face.V0]
			vertex1 := o3dModel.Vertices[face.V1]
			vertex2 := o3dModel.Vertices[face.V2]

			rl.DrawLine3D(rl.Vector3{X: vertex0.X, Y: vertex0.Y, Z: vertex0.Z}, rl.Vector3{X: vertex1.X, Y: vertex1.Y, Z: vertex1.Z}, rl.Color{R: 255, G: 255, B: 255, A: 255})
			rl.DrawLine3D(rl.Vector3{X: vertex1.X, Y: vertex1.Y, Z: vertex1.Z}, rl.Vector3{X: vertex2.X, Y: vertex2.Y, Z: vertex2.Z}, rl.Color{R: 255, G: 255, B: 255, A: 255})
			if face.V3 == uint16(lib.O3DUnused) {
				// finish up triangle
				rl.DrawLine3D(rl.Vector3{X: vertex2.X, Y: vertex2.Y, Z: vertex2.Z}, rl.Vector3{X: vertex0.X, Y: vertex0.Y, Z: vertex0.Z}, rl.Color{R: 255, G: 255, B: 255, A: 255})
			} else {
				// render as quad
				vertex3 := o3dModel.Vertices[face.V3]
				rl.DrawLine3D(rl.Vector3{X: vertex2.X, Y: vertex2.Y, Z: vertex2.Z}, rl.Vector3{X: vertex3.X, Y: vertex3.Y, Z: vertex3.Z}, rl.Color{R: 255, G: 255, B: 255, A: 255})
				rl.DrawLine3D(rl.Vector3{X: vertex3.X, Y: vertex3.Y, Z: vertex3.Z}, rl.Vector3{X: vertex0.X, Y: vertex0.Y, Z: vertex0.Z}, rl.Color{R: 255, G: 255, B: 255, A: 255})

			}
		}
		// --- END Wireframe O3D Model ---

		rl.DrawGrid(20, 5.0)
		rl.DrawLine3D(rl.Vector3{X: 0, Y: 0, Z: 0}, rl.Vector3{X: 10, Y: 0, Z: 0}, rl.Red)
		rl.DrawLine3D(rl.Vector3{X: 0, Y: 0, Z: 0}, rl.Vector3{X: 0, Y: 10, Z: 0}, rl.Green)
		rl.DrawLine3D(rl.Vector3{X: 0, Y: 0, Z: 0}, rl.Vector3{X: 0, Y: 0, Z: 10}, rl.Blue)

		rl.EndMode3D()
		rl.DrawText(fmt.Sprintf("Yaw (A/D): %f\nPitch (W/S): %f\nDistance (Q/E): %f", yaw, pitch, distance), 10, 10, 18, rl.Green)
		rl.DrawFPS(10, 80)
		rl.EndDrawing()
		//----------------------------------------------------------------------------------
	}
}
