package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/EyePlusPlus/hackattic/pkg/hackattic"
)

const (
	challengeName = "dockerized_solutions"
	pushURLBase   = "https://hackattic.com/_/push"
	localRegistry = "localhost:5000"
	repoName      = "hack"
)

type ProblemData struct {
	Credentials struct {
		User     string `json:"user"`
		Password string `json:"password"`
	} `json:"credentials"`
	IgnitionKey  string `json:"ignition_key"`
	TriggerToken string `json:"trigger_token"`
}

type Solution struct {
	Secret string `json:"secret"`
}

type NgrokTunnel struct {
	Name      string `json:"name"`
	PublicURL string `json:"public_url"`
}
type NgrokTunnelsResponse struct {
	Tunnels []NgrokTunnel `json:"tunnels"`
}

func main() {
	problem, err := hackattic.FetchProblem[ProblemData](challengeName)
	if err != nil {
		panic(err)
	}

	if err := createHtpasswdFile(problem.Credentials.User, problem.Credentials.Password); err != nil {
		panic(err)
	}
	composeCmd := exec.Command("docker-compose", "up", "-d")
	composeCmd.Dir = ".."
	if err := composeCmd.Run(); err != nil {
		panic(err)
	}
	defer func() {
		downCmd := exec.Command("docker-compose", "down")
		downCmd.Dir = ".."
		downCmd.Run()
	}()
	time.Sleep(5 * time.Second) // Give registry time to start

	ngrokCmd := exec.Command("ngrok", "http", "5000")
	if err := ngrokCmd.Start(); err != nil {
		panic(err)
	}
	defer ngrokCmd.Process.Kill()
	time.Sleep(5 * time.Second) // Give ngrok time to start

	registryHost, err := getNgrokHost()
	if err != nil {
		panic(err)
	}

	if err := triggerPush(problem.TriggerToken, registryHost); err != nil {
		panic(err)
	}

	tags, err := findImageTags(problem.Credentials.User, problem.Credentials.Password)
	if err != nil {
		panic(err)
	}

	secret, err := findSecret(tags, problem.IgnitionKey, problem.Credentials.User, problem.Credentials.Password)
	if err != nil {
		panic(err)
	}

	submitRes, submitErr := hackattic.SubmitSolution(challengeName, Solution{Secret: secret})
	if submitErr != nil {
		panic("error while submitting solution")
	}

	log.Println(submitRes)
}

func createHtpasswdFile(user, password string) error {
	if err := os.Mkdir("../docker/registry/auth", 0755); err != nil && !os.IsExist(err) {
		return err
	}
	cmd := exec.Command("htpasswd", "-B", "-b", "-c", "../docker/registry/auth/htpasswd", user, password)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getNgrokHost() (string, error) {
	resp, err := http.Get("http://127.0.0.1:4040/api/tunnels")
	if err != nil {
		return "", fmt.Errorf("could not connect to ngrok API: %w", err)
	}
	defer resp.Body.Close()

	var tunnelsResp NgrokTunnelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tunnelsResp); err != nil {
		return "", err
	}

	for _, tunnel := range tunnelsResp.Tunnels {
		if strings.HasPrefix(tunnel.PublicURL, "https://") {
			return strings.TrimPrefix(tunnel.PublicURL, "https://"), nil
		}
	}
	return "", fmt.Errorf("no HTTPS tunnel found in ngrok response")
}

func triggerPush(token, host string) error {
	payload := map[string]string{"registry_host": host}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", pushURLBase, token), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("Push response status: %s", resp.Status)
	log.Printf("Push response body: %s", string(body))
	return nil
}

func findImageTags(user, pass string) ([]string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/v2/%s/tags/list", localRegistry, repoName), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(user, pass)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tagsList struct {
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tagsList); err != nil {
		return nil, err
	}
	return tagsList.Tags, nil
}

func findSecret(tags []string, ignitionKey, user, pass string) (string, error) {
	loginCmd := exec.Command("docker", "login", localRegistry, "-u", user, "--password-stdin")
	loginCmd.Stdin = strings.NewReader(pass)
	if err := loginCmd.Run(); err != nil {
		return "", fmt.Errorf("docker login failed: %w", err)
	}

	for _, tag := range tags {
		imageName := fmt.Sprintf("%s/hack:%s", localRegistry, tag)

		pullCmd := exec.Command("docker", "pull", imageName)
		if output, err := pullCmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to pull %s: %w%s", imageName, err, string(output))
		}

		runCmd := exec.Command("docker", "run", "--rm", "-e", fmt.Sprintf("IGNITION_KEY=%s", ignitionKey), imageName)
		output, err := runCmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("failed to run %s: %w%s", imageName, err, string(output))
		}

		secret := strings.TrimSpace(string(output))

		if !strings.Contains(secret, "oops, wrong image!") {
			secretLines := strings.Split(secret, "\n")
			return secretLines[len(secretLines)-1], nil
		}
	}

	return "", fmt.Errorf("could not find secret in any of the images")
}
