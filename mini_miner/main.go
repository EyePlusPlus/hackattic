package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/EyePlusPlus/hackattic/pkg/hackattic"
)

type ProblemBlock struct {
	Nonce int             `json:"nonce"`
	Data  [][]interface{} `json:"data"`
}

type Problem struct {
	Difficulty int          `json:"difficulty"`
	Block      ProblemBlock `json:"block"`
}

type Solution struct {
	Nonce int `json:"nonce"`
}

func main() {
	problem, problemErr := hackattic.FetchProblem[Problem]("mini_miner")
	if problemErr != nil {
		panic(problemErr)
	}

	nonce := problem.Block.Nonce
	for {
		preJsonMap := map[string]interface{}{"data": problem.Block.Data, "nonce": nonce}
		jsonData, jsonErr := json.Marshal(preJsonMap)
		if jsonErr != nil {
			panic(jsonErr)
		}

		hashedBytes := sha256.Sum256(jsonData)
		hashString := fmt.Sprintf("%x", hashedBytes)

		numFullHexChars := problem.Difficulty / 4
		remainingBits := problem.Difficulty % 4

		requiredPrefix := strings.Repeat("0", numFullHexChars)
		if strings.HasPrefix(hashString, requiredPrefix) {
			if remainingBits > 0 {
				firstChar := hashString[numFullHexChars]
				val, err := strconv.ParseInt(string(firstChar), 16, 8)
				if err != nil {
					panic(err)
				}

				maxAllowedVal := (1 << (4 - remainingBits)) - 1

				if int(val) == maxAllowedVal {
					break
				}
			}

		}
		nonce++
	}

	submitResponse, submitErr := hackattic.SubmitSolution("mini_miner", Solution{Nonce: nonce})
	if submitErr != nil {
		panic(submitErr)
	}

	fmt.Println(submitResponse)

}
