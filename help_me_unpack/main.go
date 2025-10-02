package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/EyePlusPlus/hackattic/pkg/hackattic"
)

type JSONBody struct {
	Bytes string `json:"bytes"`
}

type SolutionBody struct {
	Int             int32   `json:"int"`
	Uint            uint32  `json:"uint"`
	Short           int16   `json:"short"`
	Float           float32 `json:"float"`
	Double          float64 `json:"double"`
	BigEndianDouble float64 `json:"big_endian_double"`
}

func unpackSolutionBytes(decoded []byte) (SolutionBody, error) {
	var solution SolutionBody

	reader := bytes.NewReader(decoded)
	readerErr := binary.Read(reader, binary.LittleEndian, &solution.Int)
	if readerErr != nil {
		return SolutionBody{}, readerErr
	}
	readerErr = binary.Read(reader, binary.LittleEndian, &solution.Uint)
	if readerErr != nil {
		return SolutionBody{}, readerErr
	}
	readerErr = binary.Read(reader, binary.LittleEndian, &solution.Short)
	if readerErr != nil {
		return SolutionBody{}, readerErr
	}
	_, seekErr := reader.Seek(2, io.SeekCurrent)
	if seekErr != nil {
		return SolutionBody{}, seekErr
	}
	readerErr = binary.Read(reader, binary.LittleEndian, &solution.Float)
	if readerErr != nil {
		return SolutionBody{}, readerErr
	}
	readerErr = binary.Read(reader, binary.LittleEndian, &solution.Double)
	if readerErr != nil {
		return SolutionBody{}, readerErr
	}
	readerErr = binary.Read(reader, binary.BigEndian, &solution.BigEndianDouble)
	if readerErr != nil {
		return SolutionBody{}, readerErr
	}

	return solution, nil
}

func main() {
	problemJson, fetchErr := hackattic.FetchProblem[JSONBody]("help_me_unpack")
	if fetchErr != nil {
		panic(fetchErr)
	}

	decoded, decodeErr := base64.StdEncoding.DecodeString(problemJson.Bytes)
	if decodeErr != nil {
		panic(decodeErr)
	}

	solution, unpackErr := unpackSolutionBytes(decoded)
	if unpackErr != nil {
		panic(unpackErr)
	}

	submitResponse, submitErr := hackattic.SubmitSolution("help_me_unpack", solution)
	if submitErr != nil {
		panic(submitErr)
	}

	fmt.Println(submitResponse)
}
