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

var url = "https://hackattic.com/challenges/help_me_unpack/%s?access_token=%s"

func fetchProblem() (JSONBody, error) {
	var jsonBody JSONBody
	res, err := http.Get(fmt.Sprintf(url, "problem", os.Getenv("ACCESS_TOKEN")))

	if err != nil {
		return jsonBody, fmt.Errorf("help_me_unpack#fetchProblem: %w", err)
	}

	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return jsonBody, fmt.Errorf("help_me_unpack#fetchProblem: %w", err)
	}

	if err := json.Unmarshal(resBody, &jsonBody); err != nil {
		return jsonBody, fmt.Errorf("help_me_unpack#fetchProblem: %w", err)
	}

	return jsonBody, nil
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

func submitSolution(solution SolutionBody) (map[string]interface{}, error) {
	jsonBody, err := json.Marshal(solution)
	if err != nil {
		return nil, err
	}

	submitUrl := fmt.Sprintf(url, "solve", os.Getenv("ACCESS_TOKEN"))
	response, err := http.Post(submitUrl, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var jsonResponse map[string]interface{}

	if err := json.Unmarshal(responseBody, &jsonResponse); err != nil {
		return nil, err
	}

	return jsonResponse, nil
}

func main() {

	problemJson, fetchErr := fetchProblem()
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

	submitResponseJson, submitErr := submitSolution(solution)
	if submitErr != nil {
		panic(submitErr)
	}

	fmt.Println(submitResponseJson)

}
