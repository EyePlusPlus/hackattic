package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/EyePlusPlus/hackattic/pkg/hackattic"
)

type Problem struct {
	Include string `json:"include"`
}

type Solution struct {
	Files []string `json:"files"`
}

func main() {
	challenge := "collision_course"
	solution := Solution{}
	problem, problemErr := hackattic.FetchProblem[Problem](challenge)
	if problemErr != nil {
		panic(problemErr)
	}

	if err := os.WriteFile("./prefix.bin", []byte(problem.Include), 0644); err != nil {
		panic(err)
	}

	cmd := exec.Command(
		"docker", "run", "--rm",
		"-v", fmt.Sprintf("%s:/work", os.Getenv("PWD")),
		"-w", "/work",
		"brimstone/fastcoll",
		"--prefixfile", "prefix.bin",
		"-o", "coll1.bin", "coll2.bin",
	)

	_, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	c := make(chan string)

	for _, file := range []string{"coll1.bin", "coll2.bin"} {
		wg.Add(1)
		go func() {
			fileBytes, err := os.ReadFile(fmt.Sprintf("./%s", file))
			if err != nil {
				panic(err)
			}

			file := base64.StdEncoding.EncodeToString(fileBytes)

			c <- file
			wg.Done()
		}()

	}
	go func() {
		wg.Wait()
		close(c)
	}()

	for file := range c {
		solution.Files = append(solution.Files, file)
	}

	submitResult, submitErr := hackattic.SubmitSolution(challenge, solution)
	if submitErr != nil {
		panic(submitErr)
	}

	fmt.Printf("%s", submitResult)

}
