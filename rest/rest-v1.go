package rest

import (
	"fmt"
	"io"
	"encoding/json"

	"btctx/btc"
)

type RestApiV1 struct {
	version uint16
}

type BlockTxData struct {
	Index uint16
	TxId string
	Bip141 bool
	InputCount uint16
	OutputCount uint16
}

type FieldData struct {
	Hex string
	Type string
}

type OutputData struct {
	OutputIndex uint32
	OutputType string
	Value uint64
	Address string
	OutputScript map [string] interface {}
}

type PreviousOutputRequest struct {
	InputTxId string
	InputIndex uint32
	PrevOutTxId string
	PrevOutIndex uint32
}

type PreviousOutputResponse struct {
	Value uint64
	OutputType string
	Address string
	OutputScript map [string] interface {}
}

func (api *RestApiV1) HandleRequest (httpMethod string, functionName string, getParams [] string, requestBody io.ReadCloser) string {

	errorMessage := ""
	responseJson := ""

	switch functionName {

/*

pk
ms
pkh
sh

sh-wpkh
sh-wsh
wpkh
wsh
trk
trs


pk
ms
pkh
sh
wpkh
wsh
tr

		request:
		{
			"height": 789012,
			"hash": "00000000000000000005956ad0afdcba175f9be14e9fee92282c1a8a66b9a594",
			"options":
			{
				"spend-types": ["trs"],
				"output-types": []
			}
		}
		* If both height and hash are included, height will be ignored.

		curl -X POST -d '{"height":789012}' http://127.0.0.1:8888/rest/v1/block
		curl -X POST -d '{"hash":"00000000000000000005956ad0afdcba175f9be14e9fee92282c1a8a66b9a594"}' http://127.0.0.1:8888/rest/v1/block
		curl -X POST -d '{"height":789012,"options":{"NoTxs":true,"NoTypes":true}}' http://127.0.0.1:8888/rest/v1/block
		curl -X POST -d '{"options":{"NoTxs":true,"NoTypes":true,"WScriptUsage":true,"HumanReadable":true}}' http://127.0.0.1:8888/rest/v1/block

		response:
		{
			"height": 789012,
			"hash": "00000000000000000005956ad0afdcba175f9be14e9fee92282c1a8a66b9a594",
			"previous-hash":
			"next-hash":
			"timestamp":
			"txs":
			[
				{
					"index": 0
					"id": "",
					"bip141": true,
					"input-count": 4444,
					"output-count": 5555
				}
			]
		}
*/
		case "block":

			if httpMethod != "POST" { errorMessage = fmt.Sprintf ("%s must be sent as a POST request.", functionName); break }

			// unpack the json
			var requestParams map [string] interface {}
			err := json.NewDecoder (requestBody).Decode (&requestParams)
			if err != nil { errorMessage = err.Error (); break }

			blockData := api.GetBlockData (requestParams)

			blockRequestOptions := map [string] interface {} {}
			if requestParams ["options"] != nil { blockRequestOptions = requestParams ["options"].(map [string] interface {}) }

			var blockBytes [] byte
			if blockRequestOptions ["HumanReadable"] != nil && blockRequestOptions ["HumanReadable"].(bool) {
				blockBytes, err = json.MarshalIndent (blockData, "", "\t")
			} else {
				blockBytes, err = json.Marshal (blockData)
			}
			if err != nil { fmt.Println (err.Error ()) }

			responseJson = string (blockBytes)


/*
		request:
		{
			"id": "c3e384db67470346df163a2fa50024d674ef1b3e75aa97ec6534d806c82fee7e",
			"options":
			{
			}
		}
		curl -X POST -d '{"id":"61e26d407c17e8ee33a8b166c78f78c53cdcdc0078ae1f9405e6583cfb90eaf4","options":{"HumanReadable":true}}' http://127.0.0.1:8888/rest/v1/tx

		response:
		{
			"height": 789012,
			"hash": "00000000000000000005956ad0afdcba175f9be14e9fee92282c1a8a66b9a594",
			"previous-hash":
			"next-hash":
			"timestamp":
			"txs":
			[
				{
					"index": 0
					"id": "",
					"bip141": true,
					"input-count": 4444,
					"output-count": 5555
				}
			]
		}
*/
		case "tx":

			if httpMethod != "POST" { errorMessage = fmt.Sprintf ("%s must be sent as a POST request.", functionName); break }

			// unpack the json
			var requestParams map [string] interface {}
			err := json.NewDecoder (requestBody).Decode (&requestParams)
			if err != nil { errorMessage = err.Error (); break }

			txData := api.GetTxData (requestParams)

			txRequestOptions := map [string] interface {} {}
			if requestParams ["options"] != nil { txRequestOptions = requestParams ["options"].(map [string] interface {}) }

			var txBytes [] byte
			if txRequestOptions ["HumanReadable"] != nil && txRequestOptions ["HumanReadable"].(bool) {
				txBytes, err = json.MarshalIndent (txData, "", "\t")
			} else {
				txBytes, err = json.Marshal (txData)
			}
			if err != nil { fmt.Println (err.Error ()) }

			responseJson = string (txBytes)


/*
		request:
		curl -X GET http://127.0.0.1:8888/rest/v1/current_block_height

		response:
		{
			"Current_block_height": 802114
		}
*/
		case "current_block_height":

			if httpMethod != "GET" { errorMessage = fmt.Sprintf ("%s must be sent as a GET request.", functionName); break }

			responseJson = api.getCurrentBlockHeight ()

/*
		request:
		{
			"f742b911259dd11278e9e3d34f2538c7d77837daef15fc00a047e3f13253aa0b": [0, 24],
			"f76874221ce8d7961f91b9ad5e827fb558d5ce95bc60b2722112d3384069c61b": [17, 21],
			"f7cb6da78444ff5029c5bd1f9ae6c64a1d2d65c604f2989c90cf9dc323692ed8": [0, 2]
		}
		curl -X POST -d "{\"f742b911259dd11278e9e3d34f2538c7d77837daef15fc00a047e3f13253aa0b\":[0,24],\"f76874221ce8d7961f91b9ad5e827fb558d5ce95bc60b2722112d3384069c61b\":[17,21],\"f7cb6da78444ff5029c5bd1f9ae6c64a1d2d65c604f2989c90cf9dc323692ed8\":[0,2]}" http://127.0.0.1:8888/rest/v1/previous_output_types

		response:
		{
			"f742b911259dd11278e9e3d34f2538c7d77837daef15fc00a047e3f13253aa0b:0": "P2PKH",
			"f742b911259dd11278e9e3d34f2538c7d77837daef15fc00a047e3f13253aa0b:24": "P2PKH",
			"f76874221ce8d7961f91b9ad5e827fb558d5ce95bc60b2722112d3384069c61b:17": "P2PKH",
			"f76874221ce8d7961f91b9ad5e827fb558d5ce95bc60b2722112d3384069c61b:21": "P2PKH",
			"f7cb6da78444ff5029c5bd1f9ae6c64a1d2d65c604f2989c90cf9dc323692ed8:0": "P2PKH",
			"f7cb6da78444ff5029c5bd1f9ae6c64a1d2d65c604f2989c90cf9dc323692ed8:2": "P2PKH"
		}
*/
		// called after getting a block
		case "previous_output_types":

			if httpMethod != "POST" { errorMessage = fmt.Sprintf ("%s must be sent as a POST request.", functionName); break }

			// unpack the json
			var requestedPreviousOutputs map [string] [] uint32
			err := json.NewDecoder (requestBody).Decode (&requestedPreviousOutputs)
			if err != nil { errorMessage = err.Error (); break }

			prevOutMap := api.GetPreviousOutputTypes (requestedPreviousOutputs)
			prevOutsBytes, err := json.Marshal (prevOutMap)
			if err != nil { fmt.Println (err.Error ()) }

			responseJson = string (prevOutsBytes)

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

func (api *RestApiV1) addScriptFields (scriptData map [string] interface {}, script btc.Script) {
	fieldData := make ([] FieldData, script.GetFieldCount ())
	for f, field := range script.GetFields () {
		fieldData [f] = FieldData { Hex: field.AsHex (), Type: field.AsType () }
	}
	scriptData ["Fields"] = fieldData

	if script.HasParseError () {
		scriptData ["ParseError"] = true
	}
}

func (api *RestApiV1) GetTxData (txRequest map [string] interface {}) map [string] interface {} {

	nodeClient := btc.GetNodeClient ()

	txId := txRequest ["id"].(string)
	tx := nodeClient.GetTx (txId)
	if tx.IsNil () { return nil }

	txData := make (map [string] interface {})

	txData ["BlockHeight"] = tx.GetBlockHeight ()
	txData ["BlockTime"] = tx.GetBlockTime ()
	txData ["BlockHash"] = tx.GetBlockHash ()
	txData ["Id"] = tx.GetTxId ()
	txData ["IsCoinbase"] = tx.IsCoinbase ()
	txData ["SupportsBip141"] = tx.SupportsBip141 ()
	txData ["LockTime"] = tx.GetLockTime ()

	// inputs
	inputs := make ([] map [string] interface {}, tx.GetInputCount ())
	for i, input := range tx.GetInputs () {

		inputData := make (map [string] interface {})

		inputData ["InputIndex"] = uint32 (i)
		inputData ["Coinbase"] = input.IsCoinbase ()
		inputData ["SpendType"] = input.GetSpendType ()
		inputData ["PreviousOutputTxId"] = input.GetPreviousOutputTxId ()
		inputData ["PreviousOutputIndex"] = input.GetPreviousOutputIndex ()
		inputData ["Sequence"] = input.GetSequence ()

		// input script
		inputScript := input.GetInputScript ()
		if !inputScript.IsNil () {
			inputScriptData := make (map [string] interface {})
			api.addScriptFields (inputScriptData, inputScript)
			inputData ["InputScript"] = inputScriptData
		}

		// redeem script
		redeemScript := input.GetRedeemScript ()
		if !redeemScript.IsNil () {
			redeemScriptData := make (map [string] interface {})
			api.addScriptFields (redeemScriptData, redeemScript)
			inputData ["RedeemScript"] = redeemScriptData
		}

		// segwit
		segwit := input.GetSegwit ()
		if !segwit.IsEmpty () {

			segwitData := make (map [string] interface {})

			// segwit fields
			fieldData := make ([] FieldData, segwit.GetFieldCount ())
			for f, field := range segwit.GetFields () {
				fieldData [f] = FieldData { Hex: field.AsHex (), Type: field.AsType () }
			}
			segwitData ["Fields"] = fieldData

			// witness script
			witnessScript := segwit.GetWitnessScript ()
			if !witnessScript.IsNil () {
				witnessScriptData := make (map [string] interface {})
				api.addScriptFields (witnessScriptData, witnessScript)
				segwitData ["WitnessScript"] = witnessScriptData
			}

			// tap script
			tapScript, _ := segwit.GetTapScript ()
			if !tapScript.IsNil () {
				tapScriptData := make (map [string] interface {})
				api.addScriptFields (tapScriptData, tapScript)
				if tapScript.IsOrdinal () { tapScriptData ["Ordinal"] = true }
				segwitData ["TapScript"] = tapScriptData
			}

			inputData ["Segwit"] = segwitData
		}

		inputs [i] = inputData
	}
	txData ["Inputs"] = inputs

	// previous outputs
/*
	previousOutputCount := len (previousOutputRequests)
	if previousOutputCount > 0 {
		previousOutputResponses := make ([] PreviousOutputResponse, previousOutputCount)
		for o, prevOut := range previousOutputRequests {
			previousOutputResponses [o] = GetPreviousOutputResponseData (prevOut.PrevOutTxId, prevOut.PrevOutIndex)
		}
		txData ["PreviousOutputs"] = previousOutputResponses
	}
*/
	txData ["PreviousOutputRequests"] = api.getPreviousOutputRequestData (tx)

	// outputs
	outputs := make ([] OutputData, tx.GetOutputCount ())
	for o, output := range tx.GetOutputs () {

		outputScript := output.GetOutputScript ()

		outputScriptData := make (map [string] interface {})
		api.addScriptFields (outputScriptData, outputScript)

		outputs [o] = OutputData { OutputIndex: uint32 (o), OutputType: output.GetOutputType (), Value: output.GetValue (), Address: output.GetAddress (), OutputScript: outputScriptData }
	}
	txData ["Outputs"] = outputs

	return txData
}

func (api *RestApiV1) GetPreviousOutputResponseData (txId string, outputIndex uint32) PreviousOutputResponse {
	nodeClient := btc.GetNodeClient ()
	previousOutput := nodeClient.GetPreviousOutput (txId, uint32 (outputIndex))

	outputScript := previousOutput.GetOutputScript ()
	scriptFields := outputScript.GetFields ()
	fieldData := make ([] FieldData, len (scriptFields))
	for f, field := range scriptFields {
		fieldData [f] = FieldData { Hex: field.AsHex (), Type: field.AsType () }
	}

	return PreviousOutputResponse { Value: previousOutput.GetValue (), OutputType: previousOutput.GetOutputType (), Address: previousOutput.GetAddress (), OutputScript: map [string] interface {} { "Fields": fieldData } }
}

func (api *RestApiV1) getPreviousOutputRequestData (tx btc.Tx) [] PreviousOutputRequest {

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

func (api *RestApiV1) GetBlockData (blockRequest map [string] interface {}) map [string] interface {} {

	// determine the block hash
	nodeClient := btc.GetNodeClient ()
	blockHash := ""
	if blockRequest ["hash"] != nil {
		blockHash = blockRequest ["hash"].(string)
	} else if blockRequest ["height"] != nil {
		// find an integer type that works
		// this can vary depending on the software used to send the request
		blockHeight := uint32 (0)
		uint32Test, ok := blockRequest ["height"].(uint32)
		if ok {
			blockHeight = uint32Test
		} else {
			float64Test := float64 (0)
			float64Test, ok = blockRequest ["height"].(float64)
			if ok { blockHeight = uint32 (float64Test) }
		}
		if !ok { fmt.Println ("Failed to determine integer type of block height: ", blockHeight) }
		blockHash = nodeClient.GetBlockHash (blockHeight)
	} else {
		blockHash = nodeClient.GetCurrentBlockHash ()
	}

	block := nodeClient.GetBlock (blockHash, true)
	if block.IsNil () { return nil }

	blockRequestOptions := map [string] interface {} {}
	if blockRequest ["options"] != nil { blockRequestOptions = blockRequest ["options"].(map [string] interface {}) }

	blockData := make (map [string] interface {})

	previousHash := block.GetPreviousHash ()
	if len (previousHash) > 0 { blockData ["PreviousHash"] = previousHash }
	nextHash := block.GetNextHash ()
	if len (nextHash) > 0 { blockData ["NextHash"] = nextHash }

	blockData ["Hash"] = block.GetHash ()
	blockData ["Height"] = block.GetHeight ()
	blockData ["Timestamp"] = block.GetTimestamp ()

	blockData ["InputCount"], blockData ["OutputCount"] = block.GetInputOutputCounts ()

	if blockRequestOptions ["NoTypes"] == nil || !blockRequestOptions ["NoTypes"].(bool) {
		blockData ["KnownSpendTypeMap"] = api.getKnownSpendTypes (block)
		blockData ["UnknownSpendTypeMap"] = block.GetUnknownPreviousOutputs ()
		blockData ["OutputTypeMap"] = api.getOutputTypes (block)
	}

	if blockRequestOptions ["NoTxs"] == nil || !blockRequestOptions ["NoTxs"].(bool) {
		bip141Count := uint16 (0)
		var txs [] BlockTxData
		for t, tx := range block.GetTxs () {
			if tx.SupportsBip141 () { bip141Count++ }
			txs = append (txs, BlockTxData { Index: uint16 (t), TxId: tx.GetTxId (), Bip141: tx.SupportsBip141 (), InputCount: tx.GetInputCount (), OutputCount: tx.GetOutputCount () })
		}
		blockData ["Bip141Count"] = bip141Count
		blockData ["Txs"] = txs
	}

	if blockRequestOptions != nil {
		if blockRequestOptions ["WScriptUsage"] != nil && blockRequestOptions ["WScriptUsage"].(bool) {

			witnessScriptMultisigCount := uint16 (0)
			witnessScriptCount := uint16 (0)
			tapScriptOrdinalCount := uint16 (0)
			tapScriptCount := uint16 (0)

			for _, tx := range block.GetTxs () {
				for _, input := range tx.GetInputs () {

					st := input.GetSpendType ()
					if st == btc.OUTPUT_TYPE_P2WSH || st == btc.SPEND_TYPE_P2SH_P2WSH {
						witnessScriptCount++
						if input.HasMultisigWitnessScript () {
							witnessScriptMultisigCount++
						}
					} else if st == btc.SPEND_TYPE_P2TR_Script {
						tapScriptCount++
						if input.HasOrdinalTapScript () {
							tapScriptOrdinalCount++
						}
					}
				}
			}

			if witnessScriptCount > 0 {
				blockData ["WitnessScriptMultisigCount"] = witnessScriptMultisigCount
				blockData ["WitnessScriptCount"] = witnessScriptCount
			}
			if tapScriptCount > 0 {
				blockData ["TapScriptOrdinalCount"] = tapScriptOrdinalCount
				blockData ["TapScriptCount"] = tapScriptCount
			}
		}
	}

	return blockData
}

func (api *RestApiV1) getKnownSpendTypes (block btc.Block) map [string] uint16 {

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

func (api *RestApiV1) getOutputTypes (block btc.Block) map [string] uint16 {

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

func (r *RestApiV1) GetPreviousOutputTypes (previousOutputs map [string] [] uint32) map [string] string {

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

func (r *RestApiV1) getCurrentBlockHeight () string {

	nodeClient := btc.GetNodeClient ()
	blockHash := nodeClient.GetCurrentBlockHash ()
	block := nodeClient.GetBlock (blockHash, false)

	blockJsonData := struct { Current_block_height uint32 } { Current_block_height: block.GetHeight () }
	jsonBytes, err := json.Marshal (blockJsonData)
	if err != nil { fmt.Println (err) }

	return string (jsonBytes)
}

