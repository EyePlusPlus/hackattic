package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/EyePlusPlus/hackattic/pkg/hackattic"
	"github.com/golang-jwt/jwt/v5"
)

type Problem struct {
	JWTSecret string `json:"jwt_secret"`
}

type Solution struct {
	AppUrl string `json:"app_url"`
}

type AppClaims struct {
	Append string `json:"append"`
	jwt.RegisteredClaims
}

type JSONResponse struct {
	Solution string `json:"solution"`
}

var app_url = os.Getenv("NGROK_APP_URL")
var jwtSecret string
var solution = ""

func handlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is accepted", http.StatusMethodNotAllowed)
		return
	}

	resBytes, resErr := io.ReadAll(r.Body)
	if resErr != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var claims AppClaims

	_, err := jwt.ParseWithClaims(string(resBytes), &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		http.Error(w, "Token validation failed: "+err.Error(), http.StatusUnauthorized)
		return
	}

	solution = solution + claims.Append

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(JSONResponse{Solution: solution})
}

func main() {
	var challenge = "jotting_jwts"

	problem, problemErr := hackattic.FetchProblem[Problem](challenge)
	if problemErr != nil {
		panic(problemErr)
	}

	jwtSecret = problem.JWTSecret

	http.HandleFunc("/", handlePost)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		fmt.Println("Starting server")
		if err := http.ListenAndServe(":3000", nil); err != nil {
			panic(err)
		}
	}()

	go func() {
		res, err := hackattic.SubmitSolution(challenge, Solution{AppUrl: app_url})
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", res)
		wg.Done()
	}()

	wg.Wait()

}
