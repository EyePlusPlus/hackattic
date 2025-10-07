package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/EyePlusPlus/hackattic/pkg/hackattic"
)

type Problem struct {
	ZipUrl string `json:"zip_url"`
}

type Solution struct {
	Secret string `json:"secret"`
}

func main() {
	challenge := "brute_force_zip"

	problem, err := hackattic.FetchProblem[Problem](challenge)
	if err != nil {
		panic(err)
	}

	httpRes, err := http.Get(problem.ZipUrl)
	if err != nil {
		panic(err)
	}
	defer httpRes.Body.Close()

	zipData, err := io.ReadAll(httpRes.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("ZIP downloaded (%d bytes)\n", len(zipData))

	// Save ZIP file for manual testing
	err = os.WriteFile("/tmp/challenge.zip", zipData, 0644)
	if err != nil {
		fmt.Printf("Warning: couldn't save ZIP: %v\n", err)
	} else {
		fmt.Println("ZIP saved to /tmp/challenge.zip for manual testing")
	}

	cmd := exec.Command(
		"fcrackzip",
		"-c", "a1",
		"-l", "4-6",
		"-u", "/tmp/challenge.zip",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	cleanedOutput := strings.Split(strings.TrimSpace(string(output)), " ")
	password := cleanedOutput[len(cleanedOutput)-1]

	fmt.Printf("password: %s\n", password)

	unzipCmd := exec.Command(
		"unzip", "-P", password, "/tmp/challenge.zip",
	)

	_, unzipErr := unzipCmd.CombinedOutput()
	if unzipErr != nil {
		panic(unzipErr)
	}

	secretBytes, secretErr := os.ReadFile("./secret.txt")
	if secretErr != nil {
		panic(secretErr)
	}

	secretString := strings.TrimSpace(string(secretBytes))
	submitResult, submitErr := hackattic.SubmitSolution(challenge, Solution{Secret: secretString})

	if submitErr != nil {
		panic(submitErr)
	}
	fmt.Printf("%s\n", submitResult)

}
