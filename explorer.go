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
	"btctx/themes"
	"btctx/btc"
	"btctx/test"
)

func serveFile (response http.ResponseWriter, request *http.Request) {
	// themes need support for the favicon yet
	if request.URL.Path == "/favicon.ico" {
		return
	}

	theme := themes.GetThemeForUserAgent (request.UserAgent ())
	http.ServeFile (response, request, theme.GetPath () + request.URL.Path)
}

// return an array of possible query types
func determineQueryTypes (queryParam string) [] string {

	paramLen := len (queryParam)
	if paramLen == 64 {
		// it is a block or transaction hash
		_, err := hex.DecodeString (queryParam)
		if err != nil { fmt.Println (queryParam + " is not a valid hex string."); return [] string {} }
		return [] string { "tx", "block" }
	} else {
		// it could be a block height
		_, err := strconv.ParseUint (queryParam, 10, 32)
		if err != nil { fmt.Println (queryParam + " is not a valid block height."); return [] string {} }
		return [] string { "block" }
	}

	return [] string {}
}

func homeController (response http.ResponseWriter, request *http.Request) {

	modifiedPath := request.URL.Path
	if modifiedPath [0] == '/' { modifiedPath = modifiedPath [1:] }

	var params [] string
	paramCount := 0
	if len (modifiedPath) > 0 {
		params = strings.Split (modifiedPath, "/")
		paramCount = len (params)
//		if paramCount > 2 {
			// TODO: Implement better error responses here
//			fmt.Println ("invalid path: ", request.URL.Path)
//		}
	}

	theme := themes.GetThemeForUserAgent (request.UserAgent ())
	nodeClient := btc.GetNodeClient ()
	html := ""

	queryTypes := [] string { "block" } // default query type
	if paramCount >= 1 { queryTypes [0] = params [0] }

	if queryTypes [0] == "search" {
		if paramCount < 2 { fmt.Println ("No search parameter provided."); return }
		queryTypes = determineQueryTypes (params [1])
	}

	for _, queryType := range queryTypes {
		switch queryType {
			case "block":
				blockParam := ""
				if paramCount >= 2 { blockParam = params [1] }
				paramLen := len (blockParam)

				// determine the block hash
				// the params could be a hash, height, height range or it could be left empty (current block)
				blockHash := ""
				if paramLen == 0 {
					// it is the default block, the current block
					blockHash = nodeClient.GetCurrentBlockHash ()
				} else if paramLen == 64 {
					// it could be a block hash
					blockHash = blockParam
				} else {
					// it could be a block height
					blockHeight, err := strconv.Atoi (blockParam)
					if err == nil {
						blockHash = nodeClient.GetBlockHash (blockHeight)
					}
				}

				if len (blockHash) > 0 {
					block := nodeClient.GetBlock (blockHash)
					if !block.IsNil () {
						html = theme.GetBlockHtml (block, "")
					}
				}

				break

			case "tx":
				if paramCount < 2 { fmt.Println ("Wrong number of parameters for tx. Request ignored."); return }

				txId := params [1]
				txIdBytes, err := hex.DecodeString (txId)
				if err != nil { panic (err.Error ()) }

				tx := nodeClient.GetTx (hex.EncodeToString (txIdBytes))
				if !tx.IsNil () {
					pendingInputsBytes, err := json.Marshal (tx.GetPendingInputs ())
					if err != nil { fmt.Println (err.Error ()) }

					pendingInputsJson := string (pendingInputsBytes)
					customJavascript := "var pending_inputs = JSON.parse ('" + pendingInputsJson + "');"
					html = theme.GetTxHtml (tx, customJavascript)
				}

				break

//			case "address":
//				break
		}

		if len (html) > 0 { break }
	}

	if len (html) == 0 {
		html = theme.GetExplorerPageHtml ()
	}

	fmt.Fprint (response, html)
}

func ajaxController (response http.ResponseWriter, request *http.Request) {

	nodeClient := btc.GetNodeClient ()
	theme := themes.GetThemeForUserAgent (request.UserAgent ())

	if request.Method == "POST" {
		switch request.FormValue ("method") {
			case "getpreviousoutput":
				txIdBytes, err := hex.DecodeString (request.FormValue ("Prev_out_tx_id"))
				if err != nil { fmt.Println (err.Error ()) }

				outputIndex, err := strconv.ParseUint (request.FormValue ("Prev_out_index"), 10, 32)
				if err != nil { fmt.Println (err.Error ()) }

				inputIndex, err := strconv.ParseUint (request.FormValue ("Input_index"), 10, 32)
				if err != nil {
					fmt.Println (err.Error ())
					return
				}

				previousOutput := nodeClient.GetPreviousOutput (hex.EncodeToString (txIdBytes), uint32 (outputIndex))
//				previousOutputHtml := theme.GetPreviousOutputHtml (uint32 (inputIndex), previousOutput)
				idPrefix := fmt.Sprintf ("previous-output-%d", inputIndex)
				classPrefix := fmt.Sprintf ("input-%d", inputIndex)
				previousOutputScriptHtml := theme.GetPreviousOutputScriptHtml (previousOutput.GetOutputScript (), idPrefix, classPrefix)

				// return the json response
				type previousOutputJson struct {
					Input_tx_id string
					Input_index uint32
					Prev_out_value uint64
					Prev_out_type string
					Prev_out_address string
					Prev_out_script_html string
				}

				satoshis := previousOutput.GetSatoshis ()
				previousOutputResponse := previousOutputJson { Input_tx_id: request.FormValue ("tx_id"),
																Input_index: uint32 (inputIndex),
																Prev_out_value: satoshis,
																Prev_out_type: previousOutput.GetOutputType (),
																Prev_out_address: previousOutput.GetAddress (),
//																Prev_out_html: previousOutput.GetHtmlData (outputIndex, false, 68)
																Prev_out_script_html: previousOutputScriptHtml }

				jsonBytes, err := json.Marshal (previousOutputResponse)
				if err != nil { fmt.Println (err) }

				fmt.Fprint (response, string (jsonBytes))
				break
		}
	}
}

func main () {
	settings := app.GetSettings ()

	if settings.Test.GetTestMode () != "" {
		settings.Test.ExitOnError ()
		test.RunTests (settings.Test)
		os.Exit (0)
	}

	settings.Website.ExitOnError ()
	settings.PrintListeningMessage ()

	mux := http.NewServeMux ()

	mux.HandleFunc ("/favicon.ico", serveFile)
	mux.HandleFunc ("/css/", serveFile)
	mux.HandleFunc ("/js/", serveFile)
	mux.HandleFunc ("/image/", serveFile)

	mux.HandleFunc ("/", homeController)
	mux.HandleFunc ("/ajax", ajaxController)

	log.Fatal (http.ListenAndServe (settings.Website.GetFullUrl (), mux))
}
