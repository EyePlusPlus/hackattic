package main

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/EyePlusPlus/hackattic/pkg/hackattic"
)

type RequiredDataStruct struct {
	Domain   string `json:"domain"`
	SerialNo string `json:"serial_number"`
	Country  string `json:"country"`
}

type Problem struct {
	PrivateKey   string             `json:"private_key"`
	RequiredData RequiredDataStruct `json:"required_data"`
}

type Solution struct {
	Certificate string `json:"certificate"`
}

var certificateTemplate x509.Certificate

func main() {
	challenge := "tales_of_ssl"
	problem, problemErr := hackattic.FetchProblem[Problem](challenge)
	if problemErr != nil {
		panic(problemErr)
	}

	serialNumber := new(big.Int)
	serialNumber, ok := serialNumber.SetString(problem.RequiredData.SerialNo, 0)
	if !ok {
		panic(fmt.Errorf("failed to parse big int"))
	}

	certificateTemplate.SerialNumber = serialNumber

	if _, exists := COUNTRY_CODES[problem.RequiredData.Country]; !exists {
		panic(fmt.Errorf("country code not found %s", problem.RequiredData.Country))
	}
	certificateTemplate.Subject = pkix.Name{Country: []string{COUNTRY_CODES[problem.RequiredData.Country]}, CommonName: problem.RequiredData.Domain}
	certificateTemplate.DNSNames = []string{problem.RequiredData.Domain}

	privateKeyPEM := fmt.Sprintf("-----BEGIN PRIVATE KEY-----\n%s\n-----END PRIVATE KEY-----", problem.PrivateKey)

	pBlock, _ := pem.Decode([]byte(privateKeyPEM))
	if pBlock == nil {
		panic(fmt.Errorf("pem block is nil"))
	}

	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(pBlock.Bytes)
	if err != nil {
		panic(err)
	}

	publicKey := rsaPrivateKey.PublicKey
	certificateTemplate.PublicKey = &rsaPrivateKey.PublicKey
	certificateTemplate.NotBefore = time.Now()
	certificateTemplate.NotAfter = time.Now().AddDate(1, 0, 0)

	derBytes, err := x509.CreateCertificate(rand.Reader, &certificateTemplate, &certificateTemplate, &publicKey, rsaPrivateKey)
	if err != nil {
		panic(err)
	}

	encoded := base64.StdEncoding.EncodeToString(derBytes)

	submitResult, submitErr := hackattic.SubmitSolution(challenge, Solution{Certificate: encoded})
	if submitErr != nil {
		panic(submitErr)
	}
	fmt.Printf("%s\n", submitResult)

}
