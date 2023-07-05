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
/*
aa747bfaf23eea360267cfbaa09a1afe0a1bbdeb1d327c7d01dd446d56f7230c // coinbase tx

2f10a1e81954e00284afdd292dbe3bb60a05f1788af8437829a59c77bcfb28f6 // 23 p2wpkh inputs, 2 outputs
3ccdd8921c941f29ba5981c04646d434b437e29b8c59e9ab4b5f4ad7db3cb2d9 // 1 input, 23 outputs

a674b9bb6d07990bcb3f4b05b0a57ba7b8eb866eda475ce568c08d6d922564ea // 1 p2sh input, 2 p2sh outputs

042c4f45e5bd4a0e24262436fcdc48dff83d98ee16a841ada62c2f460572a414 // 6 p2sh-p2wsh inputs, 2 p2sh outputs
d9aacc6dd9fe7d77f371b6d18e4f29b260af45ab6e4895bf631995696613f4fd // 8 p2sh-p2wpkh inputs, 1 p2wpkh output
65bd0cdd52a3603ae4d210c9c052151d83ec8ecdf29146e4b45428986f0c8ebe // Taproot Key Path inputs
5c1e2a5afb5bf0bca79474778d17909fdb71f7288371b59d557af40119b4f468 // 3 Taproot Script Path inputs

58443c231ee617227aad18e1974a88c537a7bd7fe2779623df18bde8b57ad671 // one input uses segwit fields, the other does not
eb4624e9febbada4162bd7d873377b02bbce087196315e0f4ae0bab484dd1918 // unparceable output script
c3e384db67470346df163a2fa50024d674ef1b3e75aa97ec6534d806c82fee7e // zero-length redeem scripts, 0-of-0 multisig
7393096d97bfee8660f4100ffd61874d62f9a65de9fb6acf740c4c386990ef73 // October 2022 bug, 998-of-999 multisig using OP_CHECKSIGADD
73be398c4bdc43709db7398106609eea2a7841aaf3a4fa2000dc18184faa2a7e // November 2022 bug

5f4d2593c859833db2e2d25c672a46e98f7f8564b991af9642a8b37e88af62bc // 20000 inputs
dd9f6bbf80ab36b722ca95d93268667a3ea6938288e0d4cf0e7d2e28a7a91ab3 // 13107 outputs
225ed8bc432d37cf434f80717286fd5671f676f12b573294db72a2a8f9b1e7ba
4af9047d8b4b6ffffaa5c74ee36d0506a6741ba6fc6b39fe20e4e08df799cf99 // 7000+ Tap Script fields
*/

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
					Output_type string
					Address string
					Script [] string
				}

				inputIndex, err := strconv.ParseUint (request.FormValue ("Input_index"), 10, 32)
				if err != nil {
					fmt.Println (err.Error ())
					return
				}
				previousOutputResponse := previousOutputJson { Input_tx_id: request.FormValue ("tx_id"),
														Input_index: uint32 (inputIndex),
														Value: previousOutput.GetSatoshis (),
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
