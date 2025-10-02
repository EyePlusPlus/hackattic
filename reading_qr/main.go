package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"

	"github.com/EyePlusPlus/hackattic/pkg/hackattic"
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

	return imageRes.Body, nil
}

func main() {
	problem, err := hackattic.FetchProblem[ProblemJson]("reading_qr")
	if err != nil {
		panic(fmt.Errorf("error getting the problem: %w", err))
	}

	imageResponse, imageErr := fetchImage(problem.ImageUrl)
	if imageErr != nil {
		panic(imageErr)
	}

	imageObj, _, imageDecodeErr := image.Decode(imageResponse)
	if imageDecodeErr != nil {
		panic(fmt.Errorf("error while decoding image response %w", imageDecodeErr))
	}

	imageBmp, imageBmpErr := gozxing.NewBinaryBitmapFromImage(imageObj)
	if imageBmpErr != nil {
		panic("error generating binary bitmap")
	}

	qrReader := qrcode.NewQRCodeReader()
	result, qrDecodeErr := qrReader.Decode(imageBmp, nil)
	if qrDecodeErr != nil {
		panic(fmt.Errorf("error decoding qr %w", qrDecodeErr))
	}

	submitResult, submitErr := hackattic.SubmitSolution("reading_qr", SolutionJson{Code: result.String()})

	if submitErr != nil {
		panic(fmt.Errorf("error submitting solution %w", submitErr))
	}

	fmt.Println(submitResult)
}
