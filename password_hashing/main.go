package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/pbkdf2"

	"github.com/EyePlusPlus/hackattic/pkg/hackattic"
	"golang.org/x/crypto/scrypt"
)

type Pbkdf2Attributes struct {
	Rounds int    `json:"rounds"`
	Hash   string `json:"hash"`
}

type ScryptAttributes struct {
	N       int    `json:"N"`
	R       int    `json:"r"`
	P       int    `json:"p"`
	BufLen  int    `json:"buflen"`
	Control string `json:"_control"`
}

type Problem struct {
	Password string           `json:"password"`
	Salt     string           `json:"salt"`
	Pbkdf2   Pbkdf2Attributes `json:"pbkdf2"`
	Scrypt   ScryptAttributes `json:"scrypt"`
}

type Solution struct {
	Sha256 string `json:"sha256"`
	Hmac   string `json:"hmac"`
	Pbkdf2 string `json:"pbkdf2"`
	Scrypt string `json:"scrypt"`
}

func main() {
	challenge := "password_hashing"
	solution := Solution{}
	problem, problemErr := hackattic.FetchProblem[Problem](challenge)
	if problemErr != nil {
		panic(problemErr)
	}

	passwordBytes := []byte(problem.Password)

	decodedSalt, decodeErr := base64.StdEncoding.DecodeString(problem.Salt)
	if decodeErr != nil {
		panic(decodeErr)
	}

	// sha256
	sha256hash := sha256.Sum256([]byte(problem.Password))
	solution.Sha256 = fmt.Sprintf("%x", sha256hash)

	// hmac
	h := hmac.New(sha256.New, decodedSalt)
	h.Write(passwordBytes)
	hmacDigest := h.Sum(nil)
	solution.Hmac = fmt.Sprintf("%x", hmacDigest)

	// pbkdf2
	pbkdf2bytes := pbkdf2.Key(passwordBytes, decodedSalt, problem.Pbkdf2.Rounds, 32, sha256.New)
	solution.Pbkdf2 = fmt.Sprintf("%x", pbkdf2bytes)

	// Scrypt
	scryptBytes, scryptErr := scrypt.Key(passwordBytes, decodedSalt, problem.Scrypt.N, problem.Scrypt.R, problem.Scrypt.P, problem.Scrypt.BufLen)
	if scryptErr != nil {
		panic(scryptErr)
	}
	solution.Scrypt = fmt.Sprintf("%x", scryptBytes)

	submitResult, submitErr := hackattic.SubmitSolution(challenge, solution)
	if submitErr != nil {
		panic(submitErr)
	}

	fmt.Println(submitResult)

}
