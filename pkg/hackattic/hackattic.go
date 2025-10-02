package hackattic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	baseUrl = "https://hackattic.com/challenges/%s/%s?access_token=%s"
)

func FetchProblem[T any](challenge_name string) (T, error) {
	var problem T
	url := fmt.Sprintf(baseUrl, challenge_name, "problem", os.Getenv("ACCESS_TOKEN"))
	res, err := http.Get(url)
	if err != nil {
		return problem, err
	}

	defer res.Body.Close()

	ioResult, ioErr := io.ReadAll(res.Body)
	if ioErr != nil {
		return problem, ioErr
	}

	jsonErr := json.Unmarshal(ioResult, &problem)
	if jsonErr != nil {
		return problem, jsonErr
	}

	return problem, nil
}

func SubmitSolution[T any](challenge_name string, solution T) (string, error) {
	jsonResult, jsonErr := json.Marshal(solution)
	if jsonErr != nil {
		return "", jsonErr
	}

	url := fmt.Sprintf(baseUrl, challenge_name, "solve", os.Getenv("ACCESS_TOKEN"))

	httpResult, httpErr := http.Post(url, "application/json", bytes.NewReader(jsonResult))
	if httpErr != nil {
		return "", httpErr
	}

	defer httpResult.Body.Close()

	ioResult, ioErr := io.ReadAll(httpResult.Body)
	if ioErr != nil {
		return "", ioErr
	}

	return string(ioResult), nil
}
