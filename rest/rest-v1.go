package rest

import (
	"fmt"
	"strconv"
	"io"
	"encoding/json"

	"github.com/btc-script-explorer/scantool/btc"
	"github.com/btc-script-explorer/scantool/btc/node"
)

type RestApiV1 struct {
}

type binaryFieldJson struct {	Hex string `json:"hex"`
								Type string `json:"type"` }

func scriptToJson (script btc.Script) map [string] interface {} {

	json := make (map [string] interface {})

	fields := make ([] binaryFieldJson, script.GetFieldCount ())
	for f, field := range script.GetFields () {
		fields [f] = binaryFieldJson { Hex: field.AsHex (), Type: field.AsType () }
	}

	json ["hex"] = script.AsHex ()
	json ["fields"] = fields
	if script.IsOrdinal () { json ["is_ordinal"] = true }
	if script.IsOrdinal () { json ["is_ordinal"] = true }
	json ["parse_error"] = script.HasParseError ()

	return json
}

func segwitToJson (segwit btc.Segwit) map [string] interface {} {

	json := make (map [string] interface {})

	fields := make ([] map [string] interface {}, segwit.GetFieldCount ())
	for f, field := range segwit.GetFields () {
		fields [f] = make (map [string] interface {})
		fields [f] ["hex"] = field.AsHex ()
		if len (field.AsType ()) > 0 {
			fields [f] ["type"] = field.AsType ()
		}
	}

	json ["fields"] = fields

	witnessScript := segwit.GetWitnessScript ()
	if !witnessScript.IsNil () { json ["witness_script"] = scriptToJson (witnessScript) }

	tapScript, _ := segwit.GetTapScript ()
	if !tapScript.IsNil () { json ["tap_script"] = scriptToJson (tapScript) }

	return json
}

func outputToJson (output btc.Output) map [string] interface {} {

	json := make (map [string] interface {})

	json ["value"] = output.GetValue ()
	json ["output_script"] = scriptToJson (output.GetOutputScript ())
	json ["output_type"] = output.GetOutputType ()

	address := output.GetAddress ()
	if len (address) > 0 {
		json ["address"] = address
	}

	return json
}

func inputToJson (input btc.Input) map [string] interface {} {

	json := make (map [string] interface {})

	json ["coinbase"] = input.IsCoinbase ()
	if !input.IsCoinbase () {
		json ["previous_output_tx_id"] = input.GetPreviousOutputTxId ()
		json ["previous_output_index"] = input.GetPreviousOutputIndex ()
	}

	// input script
	inputScript := input.GetInputScript ()
//	if !inputScript.IsEmpty () { json ["input_script"] = scriptToJson (inputScript) }
	json ["input_script"] = scriptToJson (inputScript)

	// segwit, if there is one
	segwit := input.GetSegwit ()
	if !segwit.IsNil () {
		json ["segwit"] = segwitToJson (segwit)
	}

	json ["sequence"] = input.GetSequence ()

	previousOutput := input.GetPreviousOutput ()
	previousOutputIncluded := len (previousOutput.GetOutputType ()) > 0
	if previousOutputIncluded {

		// previous output
		json ["previous_output"] = outputToJson (previousOutput)

		// redeem script, if there is one
		if input.HasRedeemScript () {
			json ["redeem_script"] = scriptToJson (input.GetRedeemScript ())
		}

		// other data
		json ["spend_type"] = input.GetSpendType ()
	}

	return json
}

func txToJson (tx btc.Tx) map [string] interface {} {

	inputs := make ([] map [string] interface {}, tx.GetInputCount ())
	for i, input := range tx.GetInputs () {
		inputs [i] = inputToJson (input)
	}

	outputs := make ([] map [string] interface {}, tx.GetOutputCount ())
	for o, output := range tx.GetOutputs () {
		outputs [o] = outputToJson (output)
	}

	json := make (map [string] interface {})

	json ["id"] = tx.GetTxId ()
	json ["version"] = tx.GetVersion ()
	json ["inputs"] = inputs
	json ["outputs"] = outputs
	json ["locktime"] = tx.GetLockTime ()
	json ["coinbase"] = tx.IsCoinbase ()
	json ["bip141"] = tx.SupportsBip141 ()
	json ["blockhash"] = tx.GetBlockHash ()
	json ["blocktime"] = tx.GetBlockTime ()

	return json
}

func (api *RestApiV1) GetVersion () uint16 {
	return 1
}

func (api *RestApiV1) HandleRequest (httpMethod string, functionName string, getParams [] string, requestBody io.ReadCloser) string {

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

			// get the block request options
			// get the request options
			blockRequestOptions := map [string] interface {} {}
			if requestParams ["options"] != nil { blockRequestOptions = requestParams ["options"].(map [string] interface {}) }

			// try to determine whether the hash or height parameters are the right type
			blockRequest := node.BlockRequest {}
			if requestParams ["hash"] != nil {
				switch requestParams ["hash"].(type) {
					case float64:
						return "malformed request: parameter hash is formatted as a number"
					case string:
						blockRequest.BlockKey = requestParams ["hash"].(string)
						if len (blockRequest.BlockKey) != 64 {
							return "malformed request: parameter hash is not a valid block hash"
						}
				}
			} else if requestParams ["height"] != nil {
				switch requestParams ["height"].(type) {
					case float64:
						blockRequest.BlockKey = strconv.Itoa (int (requestParams ["height"].(float64)))
					case string:
						return "malformed request: parameter height is formatted as a string"
				}
			}

			// request the block from the node proxy

			block := nodeProxy.GetBlock (blockRequest)
			if block.IsNil () {
				return "block not found"
			}

			// create the JSON response

			blockJson := struct {
				Hash string `json:"hash"`
				PreviousHash string `json:"previous_hash"`
				NextHash string `json:"next_hash"`
				Height uint32 `json:"height"`
				Timestamp int64 `json:"timestamp"`
				TxIds [] string `json:"tx_ids"`
			} {
				Hash: block.GetHash (),
				PreviousHash: block.GetPreviousHash (),
				NextHash: block.GetNextHash (),
				Height: block.GetHeight (),
				Timestamp: block.GetTimestamp (),
				TxIds: block.GetTxIds () }

			var blockBytes [] byte
			if blockRequestOptions ["human_readable"] != nil && blockRequestOptions ["human_readable"].(bool) {
				blockBytes, err = json.MarshalIndent (blockJson, "", "\t")
			} else {
				blockBytes, err = json.Marshal (blockJson)
			}
			if err != nil { fmt.Println (err.Error ()) }

			responseJson = string (blockBytes)


		case "tx":

			if httpMethod != "POST" { errorMessage = fmt.Sprintf ("%s must be sent as a POST request.", functionName); break }

			// unpack the json
			var requestParams map [string] interface {}
			err := json.NewDecoder (requestBody).Decode (&requestParams)
			if err != nil { errorMessage = err.Error (); break }

			if requestParams ["id"] == nil {
				return "id parameter is required"
			}

			// get the request options
			txRequest := node.TxRequest {}

			switch requestParams ["id"].(type) {
				case string:
					txRequest.TxId = requestParams ["id"].(string)
					if len (txRequest.TxId) != 64 { return "malformed request: parameter id is not a valid transaction id" }
				default:
					return "malformed request: id must be a hex string"
			}

			txRequestOptions := map [string] interface {} {}
			if requestParams ["options"] != nil { txRequestOptions = requestParams ["options"].(map [string] interface {}) }

			txRequest.IncludeInputDetail = txRequestOptions ["include_input_detail"] != nil && txRequestOptions ["include_input_detail"].(bool)

			// get the tx from the node proxy
			tx := nodeProxy.GetTx (txRequest)
			if tx.IsNil () {
				return "transaction not found"
			}

			txJsonObj := txToJson (tx)

			var txBytes [] byte
			if txRequestOptions ["human_readable"] != nil && txRequestOptions ["human_readable"].(bool) {
				txBytes, err = json.MarshalIndent (txJsonObj, "", "\t")
			} else {
				txBytes, err = json.Marshal (txJsonObj)
			}
			if err != nil { fmt.Println (err.Error ()) }

			responseJson = string (txBytes)


		case "output":

			if httpMethod != "POST" { errorMessage = fmt.Sprintf ("%s must be sent as a POST request.", functionName); break }

			// unpack the json
			var requestParams map [string] interface {}
			err := json.NewDecoder (requestBody).Decode (&requestParams)
			if err != nil { errorMessage = err.Error (); break }

			if requestParams ["tx_id"] == nil {
				return "tx_id parameter is required"
			}

			if requestParams ["output_index"] == nil {
				return "output_index parameter is required"
			}

			// get the request options
			outputRequest := node.OutputRequest {}

			switch requestParams ["tx_id"].(type) {
				case string:
					outputRequest.TxId = requestParams ["tx_id"].(string)
					if len (outputRequest.TxId) != 64 { return "malformed request: parameter tx_id is not a valid transaction id" }
				default:
					return "malformed request: tx_id must be a hex string"
			}

			switch requestParams ["output_index"].(type) {
				case float64:
					outputRequest.OutputIndex = uint16 (requestParams ["output_index"].(float64))
				default:
					return "malformed request: output_index must be a numeric index"
			}

			outputRequestOptions := map [string] interface {} {}
			if requestParams ["options"] != nil { outputRequestOptions = requestParams ["options"].(map [string] interface {}) }

			// get the output from the node proxy
			output := nodeProxy.GetOutput (outputRequest)
			if len (output.GetOutputType ()) == 0 { return "output not found" }

			outputJsonObj := outputToJson (output)

			var outputBytes [] byte
			if outputRequestOptions ["human_readable"] != nil && outputRequestOptions ["human_readable"].(bool) {
				outputBytes, err = json.MarshalIndent (outputJsonObj, "", "\t")
			} else {
				outputBytes, err = json.Marshal (outputJsonObj)
			}
			if err != nil { fmt.Println (err.Error ()) }

			responseJson = string (outputBytes)


		case "input":

			if httpMethod != "POST" { errorMessage = fmt.Sprintf ("%s must be sent as a POST request.", functionName); break }

			// unpack the json
			var requestParams map [string] interface {}
			err := json.NewDecoder (requestBody).Decode (&requestParams)
			if err != nil { errorMessage = err.Error (); break }

			if requestParams ["tx_id"] == nil {
				return "tx_id parameter is required"
			}

			if requestParams ["input_index"] == nil {
				return "input_index parameter is required"
			}

			// get the request options
			txRequest := node.TxRequest {}

			switch requestParams ["tx_id"].(type) {
				case string:
					txRequest.TxId = requestParams ["tx_id"].(string)
					if len (txRequest.TxId) != 64 { return "malformed request: parameter tx_id is not a valid transaction id" }
				default: return "malformed request: tx_id must be a hex string"
			}

			input_index := uint16 (0xffff)
			switch requestParams ["input_index"].(type) {
				case float64:
					input_index = uint16 (requestParams ["input_index"].(float64))
				default: return "malformed request: input_index must be a numeric index"
			}

			inputRequestOptions := map [string] interface {} {}
			if requestParams ["options"] != nil { inputRequestOptions = requestParams ["options"].(map [string] interface {}) }

			// get the input from the node proxy
			tx := nodeProxy.GetTx (txRequest)
			if tx.IsNil () || input_index >= tx.GetInputCount () { return "input not found" }

			input := tx.GetInput (input_index)
			if !input.IsCoinbase () {
				previousOutput := nodeProxy.GetOutput (node.OutputRequest { TxId: input.GetPreviousOutputTxId (), OutputIndex: input.GetPreviousOutputIndex () })
				input.SetPreviousOutput (previousOutput)

				if len (input.GetSpendType ()) == 0 { return "input not found" }
			}

			inputJsonObj := inputToJson (input)

			var inputBytes [] byte
			if inputRequestOptions ["human_readable"] != nil && inputRequestOptions ["human_readable"].(bool) {
				inputBytes, err = json.MarshalIndent (inputJsonObj, "", "\t")
			} else {
				inputBytes, err = json.Marshal (inputJsonObj)
			}
			if err != nil { fmt.Println (err.Error ()) }

			responseJson = string (inputBytes)


		case "current_block_height":

			if httpMethod != "GET" { errorMessage = fmt.Sprintf ("%s must be sent as a GET request.", functionName); break }

			height := nodeProxy.GetCurrentBlockHeight ()

			blockJsonData := struct { H int32 `json:"current_block_height"` } { H: height }
			jsonBytes, err := json.Marshal (blockJsonData)
			if err != nil { fmt.Println (err) }

			responseJson = string (jsonBytes)

		default:
			errorMessage = fmt.Sprintf ("Unknown REST v%d function: %s", api.GetVersion (), functionName)
	}

	if len (errorMessage) > 0 {
		fmt.Println (errorMessage)
		errBytes, _ := json.Marshal (RestError { Error: errorMessage })
		responseJson = string (errBytes)
	}

	return responseJson
}

