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
	// problem := Problem{PrivateKey: "MIICXAIBAAKBgQCgljpCMBTL/15HD6OnFVJeE565QZYX+mGLOT79O4Uhd2NrjKp22/tqeEoHdau06yGf+bK0fTnZdesFm4fqFBF+tjzpAses0dvmdG5x/3M4SByboaRPPX1C07kjnIOpmizmMjeVU/BPu7J4s8TVsjPzzSU3VmXRsy8xUybdTXfSZwIDAQABAoGACwJhmhoLwvSN9Rc4ZAMNM4/qyM6bSKeMumvBSsEi3ml98mihYyavtNvsT1ic3flkw7/tpXwUVDlGVIsWJVEc5dQxWMgPiWY8J8zoM8orEy6NMj0mdJVpbCmVKe+K7kAfwIIpvNj1+eOXvlQDyM2/UmsVmYpMN3nYU9Nke7dQOLkCQQDMKY65PrA27ubml9hu+rJ7opLY69Sg6oBCZ6H+mMl/EZZmJO5R95bRyObOrguh5cBEkv6nqKmr/4suATJdpFADAkEAyVxJDouU3LQEytxtyNRLIhelZUFap6Bm4/ljuGttsSDcBg/cIL/eB9i1hQS83yBwXMEio1tz6yoVfsFS+aFAzQJAa/7To4PYnMZU18ec0l/EiAfgW+Srzg8dl4LQOyfA9nlsME36ztsEKaZ3CP8h4hrxUJTdJfze+7+qdMRnSwd+1wJBAKwUcN68TIhcU6glvrCdNGQ7Pv6MXnPYcXWsIBtvu3tfMIkBrsZSEeY0vdOim+I3L68k4nwmYKb8/QepIUbyFpUCQFkatm/e9tuAYq/5JiJI/f+K9y16Mm7ww8TZwnbiip47I7vKhNodii0rzrB9Q70yghhxMk3PFngYsfXKIy/o4DM=", RequiredData: RequiredDataStruct{Domain: "snowy-waterfall-2122.gov", SerialNo: "0x5ca1ab1e", Country: "Tokelau Islands"}}
	// fmt.Printf("problem: %s\n", problem)

	serialNumber := new(big.Int)
	serialNumber, ok := serialNumber.SetString(problem.RequiredData.SerialNo, 0)
	if !ok {
		panic(fmt.Errorf("failed to parse big int"))
	}

	certificateTemplate.SerialNumber = serialNumber

	if _, exists := COUNTRY_CODES[problem.RequiredData.Country]; !exists {
		panic(fmt.Errorf("country code not found %s", problem.RequiredData.Country))
	}
	fmt.Printf("DEBUG: Country from problem data is: '%s'\n", problem.RequiredData.Country)
	certificateTemplate.Subject = pkix.Name{Country: []string{COUNTRY_CODES[problem.RequiredData.Country]}, CommonName: problem.RequiredData.Domain}
	certificateTemplate.DNSNames = []string{problem.RequiredData.Domain}

	privateKeyPEM := fmt.Sprintf("-----BEGIN PRIVATE KEY-----\n%s\n-----END PRIVATE KEY-----", problem.PrivateKey)

	pBlock, _ := pem.Decode([]byte(privateKeyPEM))
	if pBlock == nil {
		panic(fmt.Errorf("pem block is nil"))
	}
	// fmt.Printf("%v\n", pBlock.Type)

	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(pBlock.Bytes)
	if err != nil {
		panic(err)
	}

	// fmt.Printf("%v\n", rsaPrivateKey)

	publicKey := rsaPrivateKey.PublicKey
	certificateTemplate.PublicKey = &rsaPrivateKey.PublicKey
	certificateTemplate.NotBefore = time.Now()
	certificateTemplate.NotAfter = time.Now().AddDate(1, 0, 0)

	derBytes, err := x509.CreateCertificate(rand.Reader, &certificateTemplate, &certificateTemplate, &publicKey, rsaPrivateKey)
	if err != nil {
		panic(err)
	}

	// fmt.Printf("%s\n", derBytes)

	encoded := base64.StdEncoding.EncodeToString(derBytes)

	// fmt.Printf("%s\n", encoded)

	submitResult, submitErr := hackattic.SubmitSolution(challenge, Solution{Certificate: encoded})
	if submitErr != nil {
		panic(submitErr)
	}
	fmt.Printf("%s\n", submitResult)

}
