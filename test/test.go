package test

import (
	"fmt"
	"os"
	"encoding/hex"
	"encoding/json"

	"btctx/btc"
)

type txJson struct {
	BlockHash string
	BlockTime int64
	TxId string
	Version uint32
	LockTime uint32
	Bip141 bool
	Coinbase bool
	SatsIn uint64
	SatsOut uint64
	Fee uint64

	Inputs [] inputJson
	Outputs [] outputJson
}

type inputJson struct {
	PreviousOutput *outputJson
	ScriptFields [] string
	SegwitFields [] string
	SerializedScriptFields [] string
	WitnessSerializedScriptFields [] string
	SpendType string
}

type outputJson struct {
	Value uint64
	OutputType string
	ScriptFields [] string
	ParseError bool
	Address string
}

func encodeOutput (output btc.Output) outputJson {
	outputScript := output.GetOutputScript ()
	return outputJson { Value: output.GetSatoshis (), OutputType: output.GetOutputType (), Address: output.GetAddress (), ScriptFields: outputScript.GetFields (), ParseError: outputScript.HasParseError () }
}

func WriteTx (tx btc.Tx, dir string) bool {
	nodeClient := btc.GetNodeClient ()

	blockHash := tx.GetBlockHash ()
	txId := tx.GetTxId ()
	txInputs := tx.GetInputs ()
	txOutputs := tx.GetOutputs ()

	txEncoded := txJson {	BlockHash: hex.EncodeToString (blockHash [:]),
							BlockTime: tx.GetBlockTime (),
							TxId: hex.EncodeToString (txId [:]),
							Version: tx.GetVersion (),
							LockTime: tx.GetLockTime (),
							Bip141: tx.SupportsBip141 (),
							Coinbase: tx.IsCoinbase (),
							SatsIn: 0,
							SatsOut: 0,
							Fee: 0,
							Inputs: make ([] inputJson, len (txInputs)),
							Outputs: make ([] outputJson, len (txOutputs)) }

	totalOut := uint64 (0)
	for o := 0; o < len (txOutputs); o++ {
		totalOut += txOutputs [o].GetSatoshis ()
		txEncoded.Outputs [o] = encodeOutput (txOutputs [o])
	}

	totalIn := uint64 (0)
	for i := 0; i < len (txInputs); i++ {

		inputScript := txInputs [i].GetInputScript ()
		redeemScript := txInputs [i].GetRedeemScript ()
		segwit := txInputs [i].GetSegwit ()
		witnessScript := segwit.GetSerializedScript ()

		inputEncoded := inputJson {	PreviousOutput: nil,
									ScriptFields: nil,
									SegwitFields: segwit.GetFields (),
									SerializedScriptFields: redeemScript.GetFields (),
									WitnessSerializedScriptFields: witnessScript.GetFields (),
									SpendType: txInputs [i].GetSpendType () }

		if !txInputs [i].IsCoinbase () {
			// get the previous output
			previousOutput := nodeClient.GetPreviousOutput (txInputs [i].GetPreviousOutputTxId (), txInputs [i].GetPreviousOutputIndex ())
			outputEncoded := encodeOutput (previousOutput)

			totalIn += previousOutput.GetSatoshis ()

			inputEncoded.PreviousOutput = &outputEncoded
			inputEncoded.ScriptFields = inputScript.GetFields ()
		} else {
			totalIn += totalOut

			scriptFields := make ([] string, 1)
			scriptFields [0] = inputScript.GetHex ()
			inputEncoded.ScriptFields = scriptFields
		}

		txEncoded.Inputs [i] = inputEncoded
	}

	txEncoded.SatsIn = totalIn
	txEncoded.SatsOut = totalOut
	txEncoded.Fee = totalIn - totalOut

	// format it to be human-readable
	jsonBytes, err := json.MarshalIndent (txEncoded, "", "\t")
	if err != nil {
		err.Error ()
		return false
	}

	// write it to the file
	if dir [len (dir) - 1] != '/' { dir += "/" }
	err = os.WriteFile (dir + hex.EncodeToString (txId [:]) + ".json", jsonBytes, 0644)
	if err != nil {
		err.Error ()
		return false
	}

	return true
}

func VerifyTx (tx btc.Tx, dir string) bool {
	nodeClient := btc.GetNodeClient ()

	blockHash := tx.GetBlockHash ()
	txId := tx.GetTxId ()
	txInputs := tx.GetInputs ()
	txOutputs := tx.GetOutputs ()

	fileName := hex.EncodeToString (txId [:]) + ".json"

	// read the json from the file
	if dir [len (dir) - 1] != '/' { dir += "/" }
	jsonData, err := os.ReadFile (dir + fileName)
	if err != nil {
		fmt.Println (err.Error ())
		return false
	}

	// decode it
	var txEncoded txJson
	err = json.Unmarshal (jsonData, &txEncoded)
	if err != nil {
		fmt.Println (err.Error ())
		return false
	}

	// verify the tx data
	if txEncoded.BlockHash != hex.EncodeToString (blockHash [:]) { return false }
	if txEncoded.BlockTime != tx.GetBlockTime () { return false }
	if txEncoded.TxId != hex.EncodeToString (txId [:]) { return false }
	if txEncoded.Version != tx.GetVersion () { return false }
	if txEncoded.LockTime != tx.GetLockTime () { return false }
	if txEncoded.Bip141 != tx.SupportsBip141 () { return false }
	if txEncoded.Coinbase != tx.IsCoinbase () { return false }

	// number of inputs and outputs
	if len (txEncoded.Outputs) != len (txOutputs) { return false }
	if len (txEncoded.Inputs) != len (txInputs) { return false }

	// outputs
	totalOut := uint64 (0)
	for o := 0; o < len (txOutputs); o++ {
		totalOut += txOutputs [o].GetSatoshis ()
		outputScript := txOutputs [o].GetOutputScript ()

//////////////////////////////////////////////////
		if txEncoded.Outputs [o].Value != txOutputs [o].GetSatoshis () { return false }
		if txEncoded.Outputs [o].OutputType != txOutputs [o].GetOutputType () { return false }
		if len (txEncoded.Outputs [o].ScriptFields) != len (outputScript.GetFields ()) { return false }
		for f := 0; f < len (txEncoded.Outputs [o].ScriptFields); f++ {
			if txEncoded.Outputs [o].ScriptFields [f] != outputScript.GetFields () [f] { return false }
		}
		if txEncoded.Outputs [o].ParseError != outputScript.HasParseError () { return false }
		if txEncoded.Outputs [o].Address != txOutputs [o].GetAddress () { return false }
//////////////////////////////////////////////////
	}
	if txEncoded.SatsOut != totalOut { return false }

	// inputs
	totalIn := uint64 (0)
	for i := 0; i < len (txInputs); i++ {

		inputScript := txInputs [i].GetInputScript ()
		redeemScript := txInputs [i].GetRedeemScript ()
		segwit := txInputs [i].GetSegwit ()
		witnessScript := segwit.GetSerializedScript ()

		inputEncoded := inputJson {	PreviousOutput: nil,
									ScriptFields: nil,
									SegwitFields: segwit.GetFields (),
									SerializedScriptFields: redeemScript.GetFields (),
									WitnessSerializedScriptFields: witnessScript.GetFields (),
									SpendType: txInputs [i].GetSpendType () }

		// segwit fields
		if len (txEncoded.Inputs [i].SegwitFields) != len (segwit.GetFields ()) { return false }
		for f := 0; f < len (txEncoded.Inputs [i].SegwitFields); f++ {
			if txEncoded.Inputs [i].SegwitFields [f] != segwit.GetFields () [f] { return false }
		}

		// witness serialized script
		if len (txEncoded.Inputs [i].WitnessSerializedScriptFields) != len (witnessScript.GetFields ()) { return false }
		for f := 0; f < len (txEncoded.Inputs [i].WitnessSerializedScriptFields); f++ {
			if txEncoded.Inputs [i].WitnessSerializedScriptFields [f] != witnessScript.GetFields () [f] { return false }
		}

		if txEncoded.Inputs [i].SpendType != txInputs [i].GetSpendType () { return false }

		if !txInputs [i].IsCoinbase () {
			// previous output
			previousOutput := nodeClient.GetPreviousOutput (txInputs [i].GetPreviousOutputTxId (), txInputs [i].GetPreviousOutputIndex ())
			previousOutputScript := previousOutput.GetOutputScript ()

// same as above, could be a function
//////////////////////////////////////////////////
			if txEncoded.Inputs [i].PreviousOutput.Value != previousOutput.GetSatoshis () { return false }
			if txEncoded.Inputs [i].PreviousOutput.OutputType != previousOutput.GetOutputType () { return false }
			if len (txEncoded.Inputs [i].PreviousOutput.ScriptFields) != len (previousOutputScript.GetFields ()) { return false }
			for f := 0; f < len (txEncoded.Inputs [i].PreviousOutput.ScriptFields); f++ {
				if txEncoded.Inputs [i].PreviousOutput.ScriptFields [f] != previousOutputScript.GetFields () [f] { return false }
			}
			if txEncoded.Inputs [i].PreviousOutput.ParseError != previousOutputScript.HasParseError () { return false }
			if txEncoded.Inputs [i].PreviousOutput.Address != previousOutput.GetAddress () { return false }
//////////////////////////////////////////////////

			// input script
			if len (txEncoded.Inputs [i].ScriptFields) != len (inputScript.GetFields ()) { return false }
			for f := 0; f < len (txEncoded.Inputs [i].ScriptFields); f++ {
				if txEncoded.Inputs [i].ScriptFields [f] != inputScript.GetFields () [f] { return false }
			}

			// redeem script
//fmt.Println (len (txEncoded.Inputs [i].SerializedScriptFields), len (redeemScript.GetFields ()))
			if len (txEncoded.Inputs [i].SerializedScriptFields) != len (redeemScript.GetFields ()) { return false }
			for f := 0; f < len (txEncoded.Inputs [i].SerializedScriptFields); f++ {
				if txEncoded.Inputs [i].SerializedScriptFields [f] != redeemScript.GetFields () [f] { return false }
			}

			totalIn += previousOutput.GetSatoshis ()
		} else {
			totalIn += totalOut

			if len (txEncoded.Inputs [i].ScriptFields) != 1 { return false }
			if txEncoded.Inputs [i].ScriptFields [0] != inputScript.GetHex () { return false }
		}

		txEncoded.Inputs [i] = inputEncoded
	}

	if txEncoded.SatsIn != totalIn { return false }
	if txEncoded.Fee != totalIn - totalOut { return false }

	return true
}

