package main

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"os/exec"

	_ "github.com/lib/pq"

	"github.com/EyePlusPlus/hackattic/pkg/hackattic"
)

type Problem struct {
	Dump string `json:"dump"`
}
type Solution struct {
	AliveSsns []string `json:"alive_ssns"`
}

type DbEnv struct {
	PGHOST     string
	PGUSER     string
	PGPASSWORD string
	PGDATABASE string
	PGPORT     string
}

func main() {
	problem, err := hackattic.FetchProblem[Problem]("backup_restore")
	if err != nil {
		panic(fmt.Errorf("error getting the problem %w", err))
	}

	decodedString, decodeErr := base64.StdEncoding.DecodeString(problem.Dump)
	if decodeErr != nil {
		panic(decodeErr)
	}

	gzipReader, readerErr := gzip.NewReader(bytes.NewReader(decodedString))
	if readerErr != nil {
		panic(readerErr)
	}

	ioResult, ioErr := io.ReadAll(gzipReader)
	if ioErr != nil {
		panic(ioErr)
	}

	dbEnv := []string{
		"PGHOST=localhost",
		"PGPORT=5432",
		"PGUSER=root",
		"PGPASSWORD=root",
		"PGDATABASE=hackattic",
	}

	cmd := exec.Command("/opt/homebrew/bin/psql")
	cmd.Stdin = bytes.NewReader(ioResult)
	cmd.Env = dbEnv

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error restoring database dump: %v\npsql output: %s\n", err, output)
		return
	}

	dbConn, dbErr := sql.Open("postgres", "postgresql://root:root@localhost/hackattic?sslmode=disable")
	if dbErr != nil {
		panic(dbErr)
	}
	defer dbConn.Close()

	pingResult := dbConn.Ping()
	if pingResult != nil {
		panic("Ping failed")
	}

	queryResults, queryErr := dbConn.Query("SELECT ssn from criminal_records WHERE status = 'alive'")
	if queryErr != nil {
		panic(queryErr)
	}

	var aliveSsns []string

	for queryResults.Next() {
		var ssn string

		if err := queryResults.Scan(&ssn); err != nil {
			panic(err)
		}
		aliveSsns = append(aliveSsns, ssn)
	}

	submitRes, submitErr := hackattic.SubmitSolution("backup_restore", Solution{AliveSsns: aliveSsns})
	if submitErr != nil {
		panic(submitErr)
	}

	fmt.Println(string(submitRes))

}
