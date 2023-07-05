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
	if txEncoded.BlockHash != hex.EncodeToString (blockHash [:]) { fmt.Println ("Expecting block hash ", txEncoded.BlockHash, ", found ", hex.EncodeToString (blockHash [:])); return false }
	if txEncoded.BlockTime != tx.GetBlockTime () { fmt.Println ("Expecting block time ", txEncoded.BlockTime, ", found ", tx.GetBlockTime ()); return false }
	if txEncoded.TxId != hex.EncodeToString (txId [:]) { fmt.Println ("Expecting tx id ", txEncoded.TxId, ", found ", hex.EncodeToString (txId [:])); return false }
	if txEncoded.Version != tx.GetVersion () { fmt.Println ("Expecting version ", txEncoded.Version, ", found ", uint64 (tx.GetVersion ())); return false }
	if txEncoded.LockTime != tx.GetLockTime () { fmt.Println ("Expecting lock time ", txEncoded.LockTime, ", found ", tx.GetLockTime ()); return false }
	if txEncoded.Bip141 != tx.SupportsBip141 () { return false }
	if txEncoded.Coinbase != tx.IsCoinbase () { return false }

	// number of inputs and outputs
	if len (txEncoded.Outputs) != len (txOutputs) { fmt.Println ("Wrong number of outputs."); return false }
	if len (txEncoded.Inputs) != len (txInputs) { fmt.Println ("Wrong number of inputs."); return false }

	// outputs
	totalOut := uint64 (0)
	for o := 0; o < len (txOutputs); o++ {
		totalOut += txOutputs [o].GetSatoshis ()
		outputScript := txOutputs [o].GetOutputScript ()

//////////////////////////////////////////////////
		if txEncoded.Outputs [o].Value != txOutputs [o].GetSatoshis () { fmt.Println ("Wrong number of outputs."); return false }
		if txEncoded.Outputs [o].OutputType != txOutputs [o].GetOutputType () { fmt.Println ("Wrong output type."); return false }
		if len (txEncoded.Outputs [o].ScriptFields) != len (outputScript.GetFields ()) { fmt.Println ("Wrong number of outputs script fields."); return false }
		for f := 0; f < len (txEncoded.Outputs [o].ScriptFields); f++ {
			if txEncoded.Outputs [o].ScriptFields [f] != outputScript.GetFields () [f] { fmt.Println ("Wrong output script."); return false }
		}
		if txEncoded.Outputs [o].ParseError != outputScript.HasParseError () { fmt.Println ("Wrong output script status."); return false }
		if txEncoded.Outputs [o].Address != txOutputs [o].GetAddress () { fmt.Println ("Wrong address."); return false }
//////////////////////////////////////////////////
	}
	if txEncoded.SatsOut != totalOut { fmt.Println ("Wrong total outputs value."); return false }

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
		if len (txEncoded.Inputs [i].SegwitFields) != len (segwit.GetFields ()) { fmt.Println ("Wrong number of segwit fields."); return false }
		for f := 0; f < len (txEncoded.Inputs [i].SegwitFields); f++ {
			if txEncoded.Inputs [i].SegwitFields [f] != segwit.GetFields () [f] { fmt.Println ("Wrong segwit field."); return false }
		}

		// witness serialized script
		if len (txEncoded.Inputs [i].WitnessSerializedScriptFields) != len (witnessScript.GetFields ()) { fmt.Println ("Wrong number of witness script fields."); return false }
		for f := 0; f < len (txEncoded.Inputs [i].WitnessSerializedScriptFields); f++ {
			if txEncoded.Inputs [i].WitnessSerializedScriptFields [f] != witnessScript.GetFields () [f] { fmt.Println ("Wrong witness script field."); return false }
		}

		if txEncoded.Inputs [i].SpendType != txInputs [i].GetSpendType () { fmt.Println ("Wrong spend type."); return false }

		if !txInputs [i].IsCoinbase () {
			// previous output
			previousOutput := nodeClient.GetPreviousOutput (txInputs [i].GetPreviousOutputTxId (), txInputs [i].GetPreviousOutputIndex ())
			previousOutputScript := previousOutput.GetOutputScript ()

// same as above, could be a function
//////////////////////////////////////////////////
			if txEncoded.Inputs [i].PreviousOutput.Value != previousOutput.GetSatoshis () { fmt.Println ("Wrong previous output value."); return false }
			if txEncoded.Inputs [i].PreviousOutput.OutputType != previousOutput.GetOutputType () { fmt.Println ("Wrong previous output type."); return false }
			if len (txEncoded.Inputs [i].PreviousOutput.ScriptFields) != len (previousOutputScript.GetFields ()) { fmt.Println ("Wrong number of previous output script fields."); return false }
			for f := 0; f < len (txEncoded.Inputs [i].PreviousOutput.ScriptFields); f++ {
				if txEncoded.Inputs [i].PreviousOutput.ScriptFields [f] != previousOutputScript.GetFields () [f] { fmt.Println ("Wrong previous output script field."); return false }
			}
			if txEncoded.Inputs [i].PreviousOutput.ParseError != previousOutputScript.HasParseError () { fmt.Println ("Wrong previous output script status."); return false }
			if txEncoded.Inputs [i].PreviousOutput.Address != previousOutput.GetAddress () { fmt.Println ("Wrong previous output address."); return false }
//////////////////////////////////////////////////

			// input script
			if len (txEncoded.Inputs [i].ScriptFields) != len (inputScript.GetFields ()) { fmt.Println ("Wrong input script field count."); return false }
			for f := 0; f < len (txEncoded.Inputs [i].ScriptFields); f++ {
				if txEncoded.Inputs [i].ScriptFields [f] != inputScript.GetFields () [f] { fmt.Println ("Wrong input script field."); return false }
			}

			// redeem script
//fmt.Println (len (txEncoded.Inputs [i].SerializedScriptFields), len (redeemScript.GetFields ()))
			if len (txEncoded.Inputs [i].SerializedScriptFields) != len (redeemScript.GetFields ()) { fmt.Println ("Expected ", len (txEncoded.Inputs [i].SerializedScriptFields), " redeem script fields, found ", len (redeemScript.GetFields ())); return false }
			for f := 0; f < len (txEncoded.Inputs [i].SerializedScriptFields); f++ {
				if txEncoded.Inputs [i].SerializedScriptFields [f] != redeemScript.GetFields () [f] { fmt.Println ("Wrong redeem script field."); return false }
			}

			totalIn += previousOutput.GetSatoshis ()
		} else {
			totalIn += totalOut

			if len (txEncoded.Inputs [i].ScriptFields) != 1 { fmt.Println ("Wrong input script field count."); return false }
			if txEncoded.Inputs [i].ScriptFields [0] != inputScript.GetHex () { fmt.Println ("Wrong input script field."); return false }
		}

		txEncoded.Inputs [i] = inputEncoded
	}

	if txEncoded.SatsIn != totalIn { fmt.Println ("Wrong value in."); return false }
	if txEncoded.Fee != totalIn - totalOut { fmt.Println ("Wrong fee."); return false }

	return true
}

