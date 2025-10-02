package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

type ProblemJson struct {
	ImageUrl string `json:"image_url"`
}

type SolutionJson struct {
	Code string `json:"code"`
}

func fetchImage(imageUrl string) (io.Reader, error) {
	imageRes, imageErr := http.Get(imageUrl)
	if imageErr != nil {
		return nil, fmt.Errorf("error fetching image")
	}

	// imageResponseBody, readerErr := io.ReadAll(imageRes.Body)
	// if readerErr != nil {
	// 	return nil, fmt.Errorf("error reading image response")
	// }

	return imageRes.Body, nil
}

func submitSolution(code string) (map[string]interface{}, error) {
	solutionJson, jsonErr := json.Marshal(SolutionJson{Code: code})
	if jsonErr != nil {
		return nil, fmt.Errorf("error marshalling json %w", jsonErr)
	}

	submitUrl := fmt.Sprintf("https://hackattic.com/challenges/reading_qr/solve?access_token=%s&playground=1", os.Getenv("ACCESS_TOKEN"))
	submitRes, submitErr := http.Post(submitUrl, "application/json", bytes.NewReader(solutionJson))

	if submitErr != nil {
		return nil, fmt.Errorf("error submitting solution %w", submitErr)
	}

	ioBytes, ioErr := io.ReadAll(submitRes.Body)
	if ioErr != nil {
		return nil, fmt.Errorf("error reading submit response %w", ioErr)
	}

	var submitResponseJson map[string]interface{}

	if err := json.Unmarshal(ioBytes, &submitResponseJson); err != nil {
		return nil, fmt.Errorf("error parsing submit response json %w", err)
	}

	return submitResponseJson, nil
}

func main() {
	problemUrl := fmt.Sprintf("https://hackattic.com/challenges/reading_qr/problem?access_token=%s", os.Getenv("ACCESS_TOKEN"))
	problemRes, problemErr := http.Get(problemUrl)
	if problemErr != nil {
		fmt.Println(fmt.Errorf("error getting the problem"))
	}

	problemResponseBody, err := io.ReadAll(problemRes.Body)
	if err != nil {
		fmt.Println(fmt.Errorf("error reading problem response body"))
	}

	var problemJson ProblemJson

	if err := json.Unmarshal(problemResponseBody, &problemJson); err != nil {
		fmt.Println(fmt.Errorf("error parsing problem json"))
	}

	imageResponse, imageErr := fetchImage(problemJson.ImageUrl)
	if imageErr != nil {
		fmt.Println(imageErr)
	}

	// fmt.Println(imageResponse)

	imageObj, _, imageDecodeErr := image.Decode(imageResponse)
	if imageDecodeErr != nil {
		fmt.Println(fmt.Errorf("error while decoding image response %w", imageDecodeErr))
	}

	imageBmp, imageBmpErr := gozxing.NewBinaryBitmapFromImage(imageObj)
	if imageBmpErr != nil {
		fmt.Println("error generating binary bitmap")
	}

	qrReader := qrcode.NewQRCodeReader()
	result, qrDecodeErr := qrReader.Decode(imageBmp, nil)
	if qrDecodeErr != nil {
		fmt.Println("error decoding qr %w", qrDecodeErr)
	}

	submitResult, submitErr := submitSolution(result.String())

	if submitErr != nil {
		fmt.Println(fmt.Errorf("error submitting solution %w", submitErr))
	}

	fmt.Println(submitResult)

}
