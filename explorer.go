package main

import (
	"fmt"
	"strconv"
	"strings"
	"encoding/hex"
	"encoding/json"
	"os"
	"net/http"
	"log"

	"btctx/app"
	"btctx/layouts"
	"btctx/btc"
	"btctx/test"
)

/*
Line A
Line D

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
73be398c4bdc43709db7398106609eea2a7841aaf3a4fa2000dc18184faa2a7e // November 2022 bug
*/

func serveFile (response http.ResponseWriter, request *http.Request) {
	http.ServeFile (response, request, "." + request.URL.Path)
}

func homeController (response http.ResponseWriter, request *http.Request) {
	desktopLayout := layouts.GetLayout (true)
	html, err := os.ReadFile ("./html/explorer.html")
	if err != nil {
		fmt.Println (err.Error ())
		return
	}
	fmt.Fprint (response, desktopLayout.GetMainLayout (string (html), ""))
}

func ajaxController (response http.ResponseWriter, request *http.Request) {

	nodeClient := btc.GetNodeClient ()

	if request.Method == "POST" {
		switch request.FormValue ("method") {
			case "gettx":
				txIdBytes, err := hex.DecodeString (request.FormValue ("hash"))
				if err != nil { fmt.Println (err.Error ()) }

				tx := nodeClient.GetTx ([32] byte (txIdBytes))

				type txData struct {
					Tx_html string
					Pending_inputs [] btc.PendingInput
				}

				txResponse := txData { Tx_html: tx.GetHTML (), Pending_inputs: tx.GetPendingInputs () }

				jsonBytes, err := json.Marshal (txResponse)
				if err != nil { fmt.Println (err.Error ()) }

				fmt.Fprint (response, string (jsonBytes))
				break
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
	if testMode != app.TEST_NONE {

		settings.Test.ExitOnError ()

		testDirectory := settings.Test.GetDirectory ()
		nodeClient := btc.GetNodeClient ()

		if testMode == app.TEST_SAVE {

			// read the tx ids from the source file
			fileBytes, err := os.ReadFile (settings.Test.GetSourceFile ())
			if err != nil {
				fmt.Println (err.Error ())
				os.Exit (1)
			}

			// read the transactions and write the JSON files
			fileSize := len (fileBytes)
			for b := 0; b < fileSize; b += 65 {
				txId := string (fileBytes [b : b + 64])
				txIdBytes, err := hex.DecodeString (txId)
				if err != nil {
					fmt.Println (err.Error ())
					continue
				}

				tx := nodeClient.GetTx ([32] byte (txIdBytes))
				test.WriteTx (tx, testDirectory)
				fmt.Println (txId)
			}
		} else if testMode == app.TEST_VERIFY {

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
