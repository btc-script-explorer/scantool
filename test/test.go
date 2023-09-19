package test

import (
	"fmt"
	"os"
	"strings"
	"encoding/json"
//	"bufio"

//	"btctx/app"
	"btctx/rest"
)

/*
func RunTests () {

// TODO: Create TestSettings.PrintTestingMessage ().
//	testParams.PrintTestingMessage ()

	testMode := app.Settings.GetTestMode ()
	if testMode == "save" {

		// read the tx ids from the source file
		file, err := os.Open (app.Settings.GetTestSourceFile ())
		if err != nil { fmt.Println (err.Error ()) }
		defer file.Close ()

		scanner := bufio.NewScanner (file)
		for scanner.Scan () {
			line := scanner.Text ()

			// ignore blank lines and lines beginning with #
			if len (line) < 64 || line [0] == '#' {
				continue
			}

			// read only the first 64 characters, ignoring the rest of the line
			txId := line [0:64]
			WriteTx (txId, app.Settings.GetTestDirectory ())
			fmt.Println (txId)
		}
		if err := scanner.Err (); err != nil {
			fmt.Println (err.Error ())
		}
	} else if testMode == "verify" {

		testDirectory := app.Settings.GetTestDirectory ()

		// get the files to read from
		files, err := os.ReadDir (testDirectory)
		if err != nil {
			fmt.Println (err.Error ())
			os.Exit (1)
		}

		// iterate through the transactions, getting data from the node to compare with the JSON file data
		for _, file := range files {

			// verify it
			succeeded, txId := VerifyTx (file.Name (), testDirectory)
			if succeeded {
				fmt.Println (txId, "OK")
			} else {
				fmt.Println (txId, "Failed")
			}
		}
	}
}
*/

func WriteTx (txId string, dir string) bool {
	restApi := rest.RestApiV1 {}

	txEncoded := restApi.GetTxData (map [string] interface {} { "id": txId })
	if txEncoded == nil {
		// TODO: need better error handling
		fmt.Println ("Empty response from server.")
		return false
	}

	// format it to be human-readable
	jsonBytes, err := json.MarshalIndent (txEncoded, "", "\t")
	if err != nil {
		err.Error ()
		return false
	}

	// write it to the file
	if dir [len (dir) - 1] != '/' { dir += "/" }
	err = os.WriteFile (dir + txId + ".json", jsonBytes, 0644)
	if err != nil {
		err.Error ()
		return false
	}

	return true
}

func VerifyTx (filename string, dir string) (bool, string) {
	restApi := rest.RestApiV1 {}

	filePathParts := strings.Split (filename, ".")
	txId := filePathParts [0]

	txEncoded := restApi.GetTxData (map [string] interface {} { "id": txId })
	if txEncoded == nil {
		// TODO: need better error handling
		fmt.Println ("Empty response from server.")
		return false, txId
	}

	// format it to be human-readable
	jsonBytes, err := json.MarshalIndent (txEncoded, "", "\t")
	if err != nil {
		err.Error ()
		return false, txId
	}

	currentGeneratedJson := string (jsonBytes)

	// read the json from the file
	if dir [len (dir) - 1] != '/' { dir += "/" }
	jsonData, err := os.ReadFile (dir + filename)
	if err != nil {
		fmt.Println (err.Error ())
		return false, txId
	}

	savedJson := string (jsonData)

	return currentGeneratedJson == savedJson, txId
}

