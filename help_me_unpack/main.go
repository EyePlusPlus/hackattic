package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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

func main() {

	url := fmt.Sprintf("https://hackattic.com/challenges/help_me_unpack/problem?access_token=%s", os.Getenv("ACCESS_TOKEN"))
	res, err := http.Get(url)
	if err != nil {
		panic("Get failed")
	}

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var jsonBody JSONBody
	if err := json.Unmarshal(respBody, &jsonBody); err != nil {
		panic(err)
	}

	decoded, err := base64.StdEncoding.DecodeString(jsonBody.Bytes)
	if err != nil {
		panic(err)
	}

	var solution SolutionBody

	reader := bytes.NewReader(decoded)
	readerErr := binary.Read(reader, binary.LittleEndian, &solution.Int)
	if readerErr != nil {
		panic(readerErr)
	}
	readerErr = binary.Read(reader, binary.LittleEndian, &solution.Uint)
	if readerErr != nil {
		panic(readerErr)
	}
	readerErr = binary.Read(reader, binary.LittleEndian, &solution.Short)
	if readerErr != nil {
		panic(readerErr)
	}
	_, seekErr := reader.Seek(2, io.SeekCurrent)
	if seekErr != nil {
		panic(seekErr)
	}
	readerErr = binary.Read(reader, binary.LittleEndian, &solution.Float)
	if readerErr != nil {
		panic(readerErr)
	}
	readerErr = binary.Read(reader, binary.LittleEndian, &solution.Double)
	if readerErr != nil {
		panic(readerErr)
	}
	readerErr = binary.Read(reader, binary.BigEndian, &solution.BigEndianDouble)
	if readerErr != nil {
		panic(readerErr)
	}

	solRes, err := json.Marshal(solution)
	if err != nil {
		panic(err)
	}

	url = fmt.Sprintf("https://hackattic.com/challenges/help_me_unpack/solve?access_token=%s", os.Getenv("ACCESS_TOKEN"))
	verRes, err := http.Post(url, "application/json", bytes.NewReader(solRes))
	if err != nil {
		panic(err)
	}

	verResBody, err := io.ReadAll(verRes.Body)
	if err != nil {
		panic(err)
	}

	var verJson map[string]interface{}

	if err := json.Unmarshal(verResBody, &verJson); err != nil {
		panic(err)
	}

	fmt.Println(verJson)

}
