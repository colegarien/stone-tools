package lib

import (
	"encoding/binary"
	"io"
)

const (
	O3DUnused uint16 = 0xFFFF
)

type O3DModel struct {
	NumberOfVertices uint32
	NumberOfFaces    uint32
	Ignored1         uint32 // maybe flags?
	Ignored2         uint32 // maybe more flags?
	Vertices         []O3DVertex
	Faces            []O3DFace
}

type O3DVertex struct {
	X float32
	Y float32
	Z float32
}

type O3DFace struct {
	MaybeRed   uint8
	MaybeGreen uint8
	MaybeBlue  uint8
	MaybeAlpha uint8

	// UV Coordinates
	Tx0 float32
	Ty0 float32
	Tx1 float32
	Ty1 float32
	Tx2 float32
	Ty2 float32
	Tx3 float32
	Ty3 float32

	// Vertex Indexes
	V0 uint16
	V1 uint16
	V2 uint16
	V3 uint16 // unsigned 0xFFFF/O3DUnused if "unsued"

	Ignore1    uint32 // maybe flags
	MaterialId uint16
}

func ExtractO3D(o3dFile io.Reader) (O3DModel, error) {
	var o3dModel O3DModel

	err := binary.Read(o3dFile, binary.LittleEndian, &o3dModel.NumberOfVertices)
	if err != nil {
		return o3dModel, err
	}

	err = binary.Read(o3dFile, binary.LittleEndian, &o3dModel.NumberOfFaces)
	if err != nil {
		return o3dModel, err
	}

	err = binary.Read(o3dFile, binary.LittleEndian, &o3dModel.Ignored1)
	if err != nil {
		return o3dModel, err
	}

	err = binary.Read(o3dFile, binary.LittleEndian, &o3dModel.Ignored2)
	if err != nil {
		return o3dModel, err
	}

	o3dModel.Vertices = make([]O3DVertex, o3dModel.NumberOfVertices)
	for i := range o3dModel.NumberOfVertices {
		var vertex O3DVertex

		err = binary.Read(o3dFile, binary.LittleEndian, &vertex.X)
		if err != nil {
			return o3dModel, err
		}

		err = binary.Read(o3dFile, binary.LittleEndian, &vertex.Y)
		if err != nil {
			return o3dModel, err
		}

		err = binary.Read(o3dFile, binary.LittleEndian, &vertex.Z)
		if err != nil {
			return o3dModel, err
		}

		o3dModel.Vertices[i] = vertex
	}

	o3dModel.Faces = make([]O3DFace, o3dModel.NumberOfFaces)
	for i := range o3dModel.NumberOfFaces {
		var face O3DFace

		err = binary.Read(o3dFile, binary.LittleEndian, &face.MaybeRed)
		if err != nil {
			return o3dModel, err
		}

		err = binary.Read(o3dFile, binary.LittleEndian, &face.MaybeGreen)
		if err != nil {
			return o3dModel, err
		}

		err = binary.Read(o3dFile, binary.LittleEndian, &face.MaybeBlue)
		if err != nil {
			return o3dModel, err
		}

		err = binary.Read(o3dFile, binary.LittleEndian, &face.MaybeAlpha)
		if err != nil {
			return o3dModel, err
		}

		err = binary.Read(o3dFile, binary.LittleEndian, &face.Tx0)
		if err != nil {
			return o3dModel, err
		}

		err = binary.Read(o3dFile, binary.LittleEndian, &face.Ty0)
		if err != nil {
			return o3dModel, err
		}

		err = binary.Read(o3dFile, binary.LittleEndian, &face.Tx1)
		if err != nil {
			return o3dModel, err
		}

		err = binary.Read(o3dFile, binary.LittleEndian, &face.Ty1)
		if err != nil {
			return o3dModel, err
		}

		err = binary.Read(o3dFile, binary.LittleEndian, &face.Tx2)
		if err != nil {
			return o3dModel, err
		}

		err = binary.Read(o3dFile, binary.LittleEndian, &face.Ty2)
		if err != nil {
			return o3dModel, err
		}

		err = binary.Read(o3dFile, binary.LittleEndian, &face.Tx3)
		if err != nil {
			return o3dModel, err
		}

		err = binary.Read(o3dFile, binary.LittleEndian, &face.Ty3)
		if err != nil {
			return o3dModel, err
		}

		err = binary.Read(o3dFile, binary.LittleEndian, &face.V0)
		if err != nil {
			return o3dModel, err
		}

		err = binary.Read(o3dFile, binary.LittleEndian, &face.V1)
		if err != nil {
			return o3dModel, err
		}

		err = binary.Read(o3dFile, binary.LittleEndian, &face.V2)
		if err != nil {
			return o3dModel, err
		}

		err = binary.Read(o3dFile, binary.LittleEndian, &face.V3)
		if err != nil {
			return o3dModel, err
		}

		err = binary.Read(o3dFile, binary.LittleEndian, &face.Ignore1)
		if err != nil {
			return o3dModel, err
		}

		err = binary.Read(o3dFile, binary.LittleEndian, &face.MaterialId)
		if err != nil {
			return o3dModel, err
		}

		o3dModel.Faces[i] = face
	}

	return o3dModel, nil
}
