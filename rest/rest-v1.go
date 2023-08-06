package rest

import (
	"fmt"
	"strings"
	"encoding/json"
	"net/http"

	"btctx/btc"
)

type RestApiV1 struct {
	version uint16
}

type PreviousOutput struct {
	InputTxId string
	InputIndex uint32

	Value uint64
	OutputType string
	Address string

	ScriptFieldsHex [] string
	ScriptFieldsType [] string
	ScriptFieldsText [] string
}

type RestError struct {
	Error string
}

func RestV1Handler (response http.ResponseWriter, request *http.Request) {

	modifiedPath := request.URL.Path
	if len (modifiedPath) > 0 && modifiedPath [0] == '/' { modifiedPath = modifiedPath [1:] }
	lastChar := len (modifiedPath) - 1
	if lastChar >= 0 && modifiedPath [lastChar] == '/' { modifiedPath = modifiedPath [: lastChar] }

	if len (modifiedPath) == 0 {
		errorMessage := fmt.Sprintf ("Invalid request: %s", request.URL.Path)
		fmt.Println (errorMessage)
		prevOutsBytes, _ := json.Marshal (RestError { Error: errorMessage })
		fmt.Fprint (response, string (prevOutsBytes))
		return
	}

	restApi := RestApiV1 {}

	requestParts := strings.Split (modifiedPath, "/")
	if requestParts [0] != "rest" || requestParts [1] != "v1" {
		errorMessage := fmt.Sprintf ("Invalid request: %s", request.URL.Path)
		fmt.Println (errorMessage)
		prevOutsBytes, _ := json.Marshal (RestError { Error: errorMessage })
		fmt.Fprint (response, string (prevOutsBytes))
		return
	}

	restApi.version = 1

	params := strings.Split (modifiedPath, "/")
	paramCount := len (params)
	functionName := params [paramCount - 1]

	var responseJson string

	switch functionName {
		case "get_legacy_spend_types":

			if request.Method != "POST" { fmt.Println (fmt.Sprintf ("%s must be sent as a POST request.", functionName)); break }

			// unpack the json
			var requestedPreviousOutputs map [string] [] uint32
			err := json.NewDecoder (request.Body).Decode (&requestedPreviousOutputs)
			if err != nil { fmt.Println (err.Error()) }

			responseJson = restApi.getLegacySpendTypes (requestedPreviousOutputs)

		case "get_current_block_height":

			if request.Method != "GET" { fmt.Println (fmt.Sprintf ("%s must be sent as a GET request.", functionName)); break }

			responseJson = restApi.getCurrentBlockHeight ()

		default:
			errorMessage := fmt.Sprintf ("Unknown REST v1 function: %s", functionName)
			fmt.Println (errorMessage)
			prevOutsBytes, _ := json.Marshal (RestError { Error: errorMessage })
			fmt.Fprint (response, string (prevOutsBytes))
			return
	}

	fmt.Fprint (response, responseJson)
}

/*
get_legacy_spend_types

segwit spend types can be determined by their input scripts and segwit fields, but legacy spend types can not
legacy spend types have the same name as their output types, so we simply return the output types
however, if the output type is a segwit output type, then this function will assume it is a non-standard spend type
therefore, this function should not be used for segwit inputs because their spend types are already known

JSON request should be an object with tx ids as the keys and an array of integers as the value, where each integer is the index of an output in that tx
Example requesting the output types for the given outputs in the given transactions:
{
	"f32a8023f2ff9a58c1b5e35237c541d9b60f03116acbc0dbdc525a3c462bc687": [5],
	"ebd76c982b9bedf7bbb9e72dd92bc041d2bd4b3fa3564c746bf8c07171bf0628": [104, 111, 185],
	"f30707fc3a89131d91952dbbd10f616650acf2af6463bd342a4ccdd94854572b": [14]
}

JSON response will be an object with outpoints as the keys and output types as the values
The example above would return:
{
	"f32a8023f2ff9a58c1b5e35237c541d9b60f03116acbc0dbdc525a3c462bc687:5": "P2PKH",
	"ebd76c982b9bedf7bbb9e72dd92bc041d2bd4b3fa3564c746bf8c07171bf0628:104": "P2SH",
	"ebd76c982b9bedf7bbb9e72dd92bc041d2bd4b3fa3564c746bf8c07171bf0628:111": "P2SH",
	"ebd76c982b9bedf7bbb9e72dd92bc041d2bd4b3fa3564c746bf8c07171bf0628:185": "P2SH",
	"f30707fc3a89131d91952dbbd10f616650acf2af6463bd342a4ccdd94854572b:14": "P2PKH"
}

{
	"f32a8023f2ff9a58c1b5e35237c541d9b60f03116acbc0dbdc525a3c462bc687": {"5": "P2PKH"},
	"ebd76c982b9bedf7bbb9e72dd92bc041d2bd4b3fa3564c746bf8c07171bf0628": {"104": "P2SH", "111": "P2SH", "185": "P2SH"},
	"f30707fc3a89131d91952dbbd10f616650acf2af6463bd342a4ccdd94854572b": {"14": "P2PKH"}
}

*/
func (r *RestApiV1) getLegacySpendTypes (previousOutputs map [string] [] uint32) string {

	nodeClient := btc.GetNodeClient ()
	prevOutMap := make (map [string] string)
	for txId, outputIndexes := range previousOutputs {

		tx := nodeClient.GetTx (txId)
		outputs := tx.GetOutputs ()

		for _, prevOutIndex := range outputIndexes {

			previousOutputType := outputs [prevOutIndex].GetOutputType ()

			// if this is not a legacy type, it must be non-standard
			if previousOutputType != btc.OUTPUT_TYPE_P2PK && previousOutputType != btc.OUTPUT_TYPE_MultiSig && previousOutputType != btc.OUTPUT_TYPE_P2PKH && previousOutputType != btc.OUTPUT_TYPE_P2SH {
				previousOutputType = btc.SPEND_TYPE_NonStandard
			}

			jsonKey := fmt.Sprintf ("%s:%d", txId, prevOutIndex)
			jsonValue := previousOutputType
			prevOutMap [jsonKey] = jsonValue
		}
	}

	prevOutsBytes, err := json.Marshal (prevOutMap)
	if err != nil { fmt.Println (err.Error ()) }

	return string (prevOutsBytes)
}

func (r *RestApiV1) getCurrentBlockHeight () string {

	nodeClient := btc.GetNodeClient ()
	blockHash := nodeClient.GetCurrentBlockHash ()
	block := nodeClient.GetBlock (blockHash, false)

	blockJsonData := struct { Current_block_height uint32 } { Current_block_height: block.GetHeight () }
	jsonBytes, err := json.Marshal (blockJsonData)
	if err != nil { fmt.Println (err) }

	return string (jsonBytes)
}

