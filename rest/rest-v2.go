package rest

import (
	"fmt"
	"strconv"
	"io"
	"encoding/json"

	"github.com/btc-script-explorer/scantool/btc"
	"github.com/btc-script-explorer/scantool/btc/node"
)

type RestApiV2 struct {
}

type PrevOutJsonResponse struct {
	InputTxId string
	InputIndex uint32
	PrevOut btc.Output
}

func (api *RestApiV2) GetVersion () uint16 {
	return 2
}

func (api *RestApiV2) HandleRequest (httpMethod string, functionName string, getParams [] string, requestBody io.ReadCloser) string {

	nodeProxy, err := node.GetNodeProxy ()
	if err != nil {
		fmt.Println (err.Error ())
		return ""
	}

	errorMessage := ""
	responseJson := ""

	switch functionName {

		case "block":

			if httpMethod != "POST" { errorMessage = fmt.Sprintf ("%s must be sent as a POST request.", functionName); break }

			// unpack the json

			var requestParams map [string] interface {}
			err := json.NewDecoder (requestBody).Decode (&requestParams)
			if err != nil { errorMessage = err.Error (); break }

			// try to figure out exactly which block the user is requesting, and identify its block hash
			// if both hash and height are provided, hash will be used
			// if neither hash nor height is provided, the most recent block will be returned

/*
blockHash := ""
if requestParams ["hash"] != nil {
	blockHash = requestParams ["hash"].(string)
} else if requestParams ["height"] != nil {

	blockHeight, isNumeric := api.convertToBlockHeight (requestParams ["height"])
	if !isNumeric { return "" }

	blockHash = nodeProxy.GetBlockHash (blockHeight)
	if len (blockHash) == 0 { return "" }

} else {
	blockHash = nodeProxy.GetCurrentBlockHash ()
}
*/

			// get the block request options

			blockRequestOptions := node.BlockRequestOptions {}
			if requestParams ["options"] != nil {
				requestOptions := requestParams ["options"].(map [string] interface {})
				blockRequestOptions = node.BlockRequestOptions { HumanReadable: requestOptions ["HumanReadable"] != nil && requestOptions ["HumanReadable"].(bool) }
			}

			// create the block request

			// try to determine whether the hash or height parameters are the right type
			blockRequest := node.BlockRequest { Options: blockRequestOptions }
			if requestParams ["hash"] != nil {
				switch requestParams ["hash"].(type) {
					case float64:
						fmt.Println ("malformed request: parameter hash is formatted as a number, ignoring request")
						return ""
					case string:
						blockRequest.BlockKey = requestParams ["hash"].(string)
				}
			} else if requestParams ["height"] != nil {
				switch requestParams ["height"].(type) {
					case float64:
						blockRequest.BlockKey = strconv.Itoa (int (requestParams ["height"].(float64)))
					case string:
						fmt.Println ("malformed request: parameter height is formatted as a string, ignoring request")
						return ""
				}
			}

			block := nodeProxy.GetBlock (blockRequest)

			// create the JSON response

/*
			blockJson := struct {
				Hash string
				PreviousHash string
				NextHash string
				Height uint32
				Timestamp int64
				TxCount uint16
			} {
				Hash: block.GetHash (),
				PreviousHash: block.GetPreviousHash (),
				NextHash: block.GetNextHash (),
				Height: block.GetHeight (),
				Timestamp: block.GetTimestamp (),
				TxCount: block.GetTxCount () }
*/

			var blockBytes [] byte
			if blockRequestOptions.HumanReadable {
//				blockBytes, err = json.MarshalIndent (blockJson, "", "\t")
blockBytes, err = json.MarshalIndent (block, "", "\t")
			} else {
//				blockBytes, err = json.Marshal (blockJson)
blockBytes, err = json.Marshal (block)
			}
			if err != nil { fmt.Println (err.Error ()) }

			responseJson = string (blockBytes)


		case "tx":

			if httpMethod != "POST" { errorMessage = fmt.Sprintf ("%s must be sent as a POST request.", functionName); break }

			// unpack the json
			var requestParams map [string] interface {}
			err := json.NewDecoder (requestBody).Decode (&requestParams)
			if err != nil { errorMessage = err.Error (); break }

			txRequest := node.TxRequest { TxId: requestParams ["id"].(string) }
			tx := nodeProxy.GetTx (txRequest)

			txRequestOptions := map [string] interface {} {}
			if requestParams ["options"] != nil { txRequestOptions = requestParams ["options"].(map [string] interface {}) }

			var txBytes [] byte
			if txRequestOptions ["HumanReadable"] != nil && txRequestOptions ["HumanReadable"].(bool) {
				txBytes, err = json.MarshalIndent (tx, "", "\t")
			} else {
				txBytes, err = json.Marshal (tx)
			}
			if err != nil { fmt.Println (err.Error ()) }

			responseJson = string (txBytes)


		case "current_block_height":

			if httpMethod != "GET" { errorMessage = fmt.Sprintf ("%s must be sent as a GET request.", functionName); break }

			height := nodeProxy.GetCurrentBlockHeight ()

	blockJsonData := struct { H int32 `json:"current_block_height"` } { H: height }
	jsonBytes, err := json.Marshal (blockJsonData)
	if err != nil { fmt.Println (err) }

			responseJson = string (jsonBytes)


/*
		// called after getting a block
		case "output_types":

			if httpMethod != "POST" { errorMessage = fmt.Sprintf ("%s must be sent as a POST request.", functionName); break }

			// unpack the json
			var requestedPreviousOutputs map [string] [] uint32
			err := json.NewDecoder (requestBody).Decode (&requestedPreviousOutputs)
			if err != nil { errorMessage = err.Error (); break }

			prevOutMap := api.GetPreviousOutputTypes (requestedPreviousOutputs)
			prevOutsBytes, err := json.Marshal (prevOutMap)
			if err != nil { fmt.Println (err.Error ()) }

			responseJson = string (prevOutsBytes)
*/


		// called when the tx id and input index need to be returned with the response
		case "previous_output":

			if httpMethod != "POST" { errorMessage = fmt.Sprintf ("%s must be sent as a POST request.", functionName); break }

			// unpack the json
			var previousOutputJsonIn btc.PreviousOutputRequest
			err := json.NewDecoder (requestBody).Decode (&previousOutputJsonIn)
			if err != nil { errorMessage = err.Error (); break }


			previousOutputRequest := make (chan btc.PreviousOutputRequest)
			previousOutputRequest <- previousOutputJsonIn
			previousOutputResponseChannel := nodeProxy.GetPreviousOutput (previousOutputRequest)
//			previousOutput := api.GetPreviousOutputResponseData (txId, uint32 (outputIndex))
previousOutput := <- previousOutputResponseChannel

			// return the json response
inputIndex := previousOutputJsonIn.InputIndex
inputTxId := previousOutputJsonIn.TxId
			previousOutputResponse := PrevOutJsonResponse { InputTxId: inputTxId,
																InputIndex: uint32 (inputIndex),
																PrevOut: previousOutput }

			jsonBytes, err := json.Marshal (previousOutputResponse)
			if err != nil { fmt.Println (err) }

			responseJson = string (jsonBytes)


/*
		case "serialized_script_usage":

			if httpMethod != "POST" { errorMessage = fmt.Sprintf ("%s must be sent as a POST request.", functionName); break }

			// unpack the json
			var requestParams map [string] interface {}
			err := json.NewDecoder (requestBody).Decode (&requestParams)
			if err != nil { errorMessage = err.Error (); break }

			blockHash := ""
			if requestParams ["hash"] != nil {
				blockHash = requestParams ["hash"].(string)
			} else if requestParams ["height"] != nil {

				blockHeight, isNumeric := api.convertToBlockHeight (requestParams ["height"])
				if !isNumeric { fmt.Println ("Failed to read height parameter."); break }

				nodeClient := btc.GetNodeClient ()
				blockHash = nodeClient.GetBlockHash (blockHeight)
				if len (blockHash) == 0 { fmt.Println ("Failed to read height parameter."); break }
			} else {
				fmt.Println ("hash or height parameter required for ", functionName)
				break
			}

			requestOptions := map [string] interface {} {}
			if requestParams ["options"] != nil { requestOptions = requestParams ["options"].(map [string] interface {}) }

			tap := requestOptions ["tap"] != nil && requestOptions ["tap"].(bool)
			redeem := requestOptions ["redeem"] != nil && requestOptions ["redeem"].(bool)
			witness := requestOptions ["witness"] != nil && requestOptions ["witness"].(bool)
			serializedScriptMap := api.getSerializedScriptJson (blockHash, tap, redeem, witness)

			var resultBytes [] byte
			if requestOptions ["HumanReadable"] != nil && requestOptions ["HumanReadable"].(bool) {
				resultBytes, err = json.MarshalIndent (serializedScriptMap, "", "\t")
			} else {
				resultBytes, err = json.Marshal (serializedScriptMap)
			}
			if err != nil { fmt.Println (err.Error ()) }

			responseJson = string (resultBytes)
*/


		default:
			errorMessage = fmt.Sprintf ("Unknown REST v1 function: %s", functionName)
	}

	if len (errorMessage) > 0 {
		fmt.Println (errorMessage)
		errBytes, _ := json.Marshal (RestError { Error: errorMessage })
		responseJson = string (errBytes)
	}

	return responseJson
}

/*
func (api *RestApiV2) convertToBlockHeight (param interface {}) (uint32, bool) {
	// find an integer type for the height field
	// this can vary depending on the software used to send the request
	uint32Test, uint32Okay := param.(uint32)
	float64Test, float64Okay := param.(float64)

	if uint32Okay {
		return uint32Test, true
	} else if float64Okay {
		return uint32 (float64Test), true
	}

	// none of the types worked, it isn't a valid height or it is a different numeric type
	return 0xffffffff, false
}
*/

/*
// handles multiple outputs in multiple transactions
// returns outpoints and the output type of each
func (r *RestApiV2) GetPreviousOutputTypes (previousOutputs map [string] [] uint32) map [string] string {

	nodeClient := btc.GetNodeClient ()
	prevOutMap := make (map [string] string)
	for txId, outputIndexes := range previousOutputs {

		tx := nodeClient.GetTx (txId)
		outputs := tx.GetOutputs ()

		for _, prevOutIndex := range outputIndexes {
			key := fmt.Sprintf ("%s:%d", txId, prevOutIndex)
			value := outputs [prevOutIndex].GetOutputType ()
			prevOutMap [key] = value
		}
	}

	return prevOutMap
}

func (api *RestApiV2) getPreviousOutputRequestData (tx btc.Tx) [] PreviousOutputRequest {

	if tx.IsCoinbase () { return [] PreviousOutputRequest {} }

	txId := tx.GetTxId ()
	inputs := tx.GetInputs ()
	inputCount := len (inputs)
	previousOutputs := make ([] PreviousOutputRequest, inputCount)
	for i := uint32 (0); i < uint32 (inputCount); i++ {
		previousOutputs [i] = PreviousOutputRequest { InputTxId: txId, InputIndex: i, PrevOutTxId: inputs [i].GetPreviousOutputTxId (), PrevOutIndex: inputs [i].GetPreviousOutputIndex () }
	}

	return previousOutputs
}

func (api *RestApiV2) getSerializedScriptJson (blockHash string, tap bool, redeem bool, witness bool) map [string] [] map [string] interface {} {

	nodeClient := btc.GetNodeClient ()
	block := nodeClient.GetBlock (blockHash, true)

	var ordResults [] map [string] interface {}
	var redeemResults [] map [string] interface {}
	var witnessResults [] map [string] interface {}

	for _, tx := range block.GetTxs () {
		for i, input := range tx.GetInputs () {

			resultObj := make (map [string] interface {})
			resultObj ["tx"] = tx.GetTxId ()
			resultObj ["input"] = i

			st := input.GetSpendType ()
			if tap && st == btc.SPEND_TYPE_P2TR_Script && input.HasOrdinalTapScript () {

				segwit := input.GetSegwit ()
				script, _ := segwit.GetTapScript ()
				ordinalFields := script.GetFields ()

				ordBegin := 2
				if ordinalFields [3].AsHex () == "OP_DROP" { ordBegin = 4 }

				ordMimeType := ordinalFields [ordBegin + 4].AsText ()
				mimeTypeTextPlain := strings.Index (ordMimeType, "text/plain") != -1
				mimeTypeTextHtml := strings.Index (ordMimeType, "text/html") != -1
				mimeTypeApplicationJson := strings.Index (ordMimeType, "application/json") != -1
//				mimeTypeTextJavascript := strings.Index (ordMimeType, "text/javascript") != -1

				resultObj ["mimetype"] = ordMimeType

				ordParams := make (map [string] interface {})

				if mimeTypeTextPlain || mimeTypeTextHtml || mimeTypeApplicationJson {

					ordJson := ordinalFields [ordBegin + 6].AsBytes ()
					err := json.Unmarshal (ordJson, &ordParams)

					dataIsJson := err == nil
					if dataIsJson {
						resultObj ["ord"] = ordParams
					} else {
						resultObj ["ord"] = ordinalFields [ordBegin + 6].AsText ()
					}
				} else {
					dataSize := uint32 (0)
					for dataSegment := ordBegin + 6; dataSegment <= len (ordinalFields) - 2; dataSegment++ {
						dataSize += uint32 (len (ordinalFields [dataSegment].AsBytes ()))
					}
					resultObj ["data_size"] = dataSize
				}

				ordResults = append (ordResults, resultObj)
			} else if redeem && st == btc.OUTPUT_TYPE_P2SH && input.HasMultisigRedeemScript () {

				script := input.GetRedeemScript ()
				multisigFields := script.GetFields ()

				sigCount := multisigFields [0].AsBytes () [0]
				if sigCount >= 0x51 && sigCount <= 0x60 { sigCount -= 0x50 }

				keyCount := multisigFields [len (multisigFields) - 2].AsBytes () [0]
				if keyCount >= 0x51 && keyCount <= 0x60 { keyCount -= 0x50 }

				resultObj ["sig_count"] = uint8 (sigCount)
				resultObj ["key_count"] = uint8 (keyCount)

				redeemResults = append (redeemResults, resultObj)

			} else if witness && (st == btc.OUTPUT_TYPE_P2WSH || st == btc.SPEND_TYPE_P2SH_P2WSH) && input.HasMultisigWitnessScript () {

				segwit := input.GetSegwit ()
				script := segwit.GetWitnessScript ()
				multisigFields := script.GetFields ()

				sigCount := multisigFields [0].AsBytes () [0]
				if sigCount >= 0x51 && sigCount <= 0x60 { sigCount -= 0x50 }

				keyCount := multisigFields [len (multisigFields) - 2].AsBytes () [0]
				if keyCount >= 0x51 && keyCount <= 0x60 { keyCount -= 0x50 }

				resultObj ["sig_count"] = uint8 (sigCount)
				resultObj ["key_count"] = uint8 (keyCount)
				resultObj ["spend_type"] = st

				witnessResults = append (witnessResults, resultObj)
			}
		}
	}

	results := make (map [string] [] map [string] interface {})
	if len (ordResults) > 0 { results ["ordinals"] = ordResults }
	if len (redeemResults) > 0 { results ["redeem"] = redeemResults }
	if len (witnessResults) > 0 { results ["witness"] = witnessResults }

	return results
}

func (api *RestApiV2) getKnownSpendTypes (block btc.Block) map [string] uint16 {

	spendTypeMap := make (map [string] uint16)
	for _, tx := range block.GetTxs () {
		for _, input := range tx.GetInputs () {
			if input.IsCoinbase () { continue }

			spendType := input.GetSpendType ()
			if len (spendType) > 0 {
				spendTypeMap [spendType]++
			}
		}
	}

	return spendTypeMap
}

func (api *RestApiV2) getOutputTypes (block btc.Block) map [string] uint16 {

	outputTypeMap := make (map [string] uint16)
	for _, tx := range block.GetTxs () {
		for _, output := range tx.GetOutputs () {

			outputType := output.GetOutputType ()
			if len (outputType) > 0 {
				outputTypeMap [outputType]++
			}
		}
	}

	return outputTypeMap
}
*/

/*
legacy_spend_types

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

