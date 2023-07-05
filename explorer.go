package main

import (
	"fmt"
	"strconv"
	"strings"
	"encoding/hex"
	"encoding/json"
	"bufio"
	"os"
	"net/http"
	"log"

	"btctx/app"
	"btctx/themes"
	"btctx/btc"
	"btctx/test"
)

func serveFile (response http.ResponseWriter, request *http.Request) {
	theme := themes.GetThemeForUserAgent (request.UserAgent ())
	http.ServeFile (response, request, theme.GetPath () + request.URL.Path)
}

func homeController (response http.ResponseWriter, request *http.Request) {

	params := strings.Split (request.URL.Path, "/")
	if len (params) > 3 {
		// TODO: Implement better error responses here
		fmt.Println ("invalid path: ", request.URL.Path)
	}

	queryId := ""
	queryResult := ""
	customJavascript := ""

	theme := themes.GetThemeForUserAgent (request.UserAgent ())

	if len (params) == 3 {
//fmt.Println (request.URL.Path)
		queryType := params [1]
		queryId = params [2]

		nodeClient := btc.GetNodeClient ()

		switch queryType {
			case "block":
				break
			case "tx":
				txIdBytes, err := hex.DecodeString (queryId)
				if err != nil { fmt.Println (err.Error ()) }

				tx := nodeClient.GetTx ([32] byte (txIdBytes [:]))
				queryResult = tx.GetHtml (theme)

				pendingInputsBytes, err := json.Marshal (tx.GetPendingInputs ())
				if err != nil { fmt.Println (err.Error ()) }

				pendingInputsJson := string (pendingInputsBytes)
				customJavascript = "var pending_inputs = JSON.parse ('" + pendingInputsJson + "');"
				break
			case "address":
				break
		}
	}

	fmt.Fprint (response, theme.GetExplorerPageHtml (queryId, queryResult, customJavascript))
}

func ajaxController (response http.ResponseWriter, request *http.Request) {

	nodeClient := btc.GetNodeClient ()
//	theme := themes.GetThemeForUserAgent (request.UserAgent ())

	if request.Method == "POST" {
		switch request.FormValue ("method") {
			case "getpreviousoutput":
				txIdBytes, err := hex.DecodeString (request.FormValue ("Prev_tx_id"))
				if err != nil { fmt.Println (err.Error ()) }

				outputIndex, err := strconv.ParseUint (request.FormValue ("Output_index"), 10, 32)
				if err != nil { fmt.Println (err.Error ()) }

				previousOutput := nodeClient.GetPreviousOutput ([32] byte (txIdBytes), uint32 (outputIndex))
				previousOutputScript := previousOutput.GetOutputScript ()

				// return the json response
				type previousOutputJson struct {
					Input_tx_id string
					Input_index uint32
					Value uint64
					Value_html string
					Output_type string
					Address string
					Script [] string
				}

				inputIndex, err := strconv.ParseUint (request.FormValue ("Input_index"), 10, 32)
				if err != nil {
					fmt.Println (err.Error ())
					return
				}

				satoshis := previousOutput.GetSatoshis ()
				previousOutputResponse := previousOutputJson { Input_tx_id: request.FormValue ("tx_id"),
														Input_index: uint32 (inputIndex),
														Value: satoshis,
														Value_html: btc.GetValueHtml (satoshis),
														Output_type: previousOutput.GetOutputType (),
														Address: previousOutput.GetAddress (),
														Script: previousOutputScript.GetFields () }

				jsonBytes, err := json.Marshal (previousOutputResponse)
				if err != nil { fmt.Println (err) }

				fmt.Fprint (response, string (jsonBytes))
				break
		}
	}
}

func main () {
	settings := app.GetSettings ()

	// if we are writing test files, write it to a file
	testMode := settings.Test.GetTestMode ()
	if testMode != "" {

		settings.Test.ExitOnError ()

		testDirectory := settings.Test.GetDirectory ()
		nodeClient := btc.GetNodeClient ()

// TODO: Create welcome banner for testing. Show the mode, directory and source file.

		if testMode == "save" {

			// read the tx ids from the source file
			file, err := os.Open (settings.Test.GetSourceFile ())
			if err != nil { fmt.Println (err.Error ()) }
			defer file.Close ()

			scanner := bufio.NewScanner (file)
			for scanner.Scan () {
				line := scanner.Text ()

				// ignore blank lines and lines beginning with #
				if len (line) < 64 || line [0] == '#' {
					continue
				}

				// read only the first 64 characters of each line, leaving the rest for comments
				txId := line [0:64]
				txIdBytes, err := hex.DecodeString (txId)
				if err != nil {
					fmt.Println (err.Error ())
					continue
				}

				// write the JSON files
				tx := nodeClient.GetTx ([32] byte (txIdBytes))
				test.WriteTx (tx, testDirectory)
				fmt.Println (txId)
			}
			if err := scanner.Err (); err != nil {
				fmt.Println (err.Error ())
			}

/*
			fileBytes, err := os.ReadFile (settings.Test.GetSourceFile ())
			if err != nil {
				fmt.Println (err.Error ())
				os.Exit (1)
			}

			// read the transactions
			fileSize := len (fileBytes)
			for b := 0; b < fileSize; b += 65 {

				// ignore blank lines and lines beginning with #

				// read only the first 64 characters of each line, leaving the rest for comments
				txId := string (fileBytes [b : b + 64])
				txIdBytes, err := hex.DecodeString (txId)
				if err != nil {
					fmt.Println (err.Error ())
					continue
				}

				// write the JSON files
				tx := nodeClient.GetTx ([32] byte (txIdBytes))
				test.WriteTx (tx, testDirectory)
				fmt.Println (txId)
			}
*/
		} else if testMode == "verify" {

			// get the files to read from
			files, err := os.ReadDir (testDirectory)
			if err != nil {
				fmt.Println (err.Error ())
				os.Exit (1)
			}

			// iterate through the transactions, getting data from the node to compare with the JSON file data
			nodeClient := btc.GetNodeClient ()
			for _, file := range files {

				// extract the tx id from the filename
				filePathParts := strings.Split (file.Name (), ".")
				txId := filePathParts [0]

				txIdBytes, err := hex.DecodeString (txId)
				if err != nil {
					fmt.Println (err.Error ())
					os.Exit (1)
				}

				// get the transaction from the node
				tx := nodeClient.GetTx ([32] byte (txIdBytes))

				// verify it
				if test.VerifyTx (tx, testDirectory) {
					fmt.Println (txId, "OK")
				} else {
					fmt.Println (txId, "Failed")
				}
			}
		}

		os.Exit (0)
	}

	settings.Website.ExitOnError ()
	settings.PrintListeningMessage ()

	mux := http.NewServeMux ()

	mux.HandleFunc ("/css/", serveFile)
	mux.HandleFunc ("/js/", serveFile)

	mux.HandleFunc ("/", homeController)
	mux.HandleFunc ("/ajax", ajaxController)

	log.Fatal (http.ListenAndServe (settings.Website.GetFullUrl (), mux))
}
