package test

import (
	"fmt"
	"os"
	"strings"
	"encoding/hex"
	"encoding/json"
	"bufio"

	"btctx/app"
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

	InputScriptFields [] btc.ScriptField
	RedeemScriptFields [] btc.ScriptField

	SegwitFields [] btc.SegwitField
	WitnessScriptFields [] btc.ScriptField
	TapScriptIndex int64
	TapScriptFields [] btc.ScriptField

	SpendType string
}

type outputJson struct {
	Value uint64
	OutputType string
	OutputScriptFields [] btc.ScriptField
	ParseError bool
	Address string
}

func encodeInput (input btc.Input) inputJson {
	inputScript := input.GetInputScript ()
	redeemScript := input.GetRedeemScript ()
	segwit := input.GetSegwit ()
	witnessScript := segwit.GetWitnessScript ()
	tapScript, tapScriptIndex := segwit.GetTapScript ()
	return inputJson { PreviousOutput: nil,
						InputScriptFields: inputScript.GetFields (),
						RedeemScriptFields: redeemScript.GetFields (),
						SegwitFields: segwit.GetFields (),
						WitnessScriptFields: witnessScript.GetFields (),
						TapScriptIndex: tapScriptIndex,
						TapScriptFields: tapScript.GetFields (),
						SpendType: input.GetSpendType () }
}

func encodeOutput (output btc.Output) outputJson {
	outputScript := output.GetOutputScript ()
	return outputJson { Value: output.GetValue (),
						OutputType: output.GetOutputType (),
						Address: output.GetAddress (),
						OutputScriptFields: outputScript.GetFields (),
						ParseError: outputScript.HasParseError () }
}

func RunTests () {

	nodeClient := btc.GetNodeClient ()

// TODO: Create TestSettings.PrintTestingMessage ().
//	testParams.PrintTestingMessage ()

	testMode := app.Settings.GetTestMode ()
	if testMode == "save" {

		// read the tx ids from the source file
		file, err := os.Open (app.Settings.GetTestSourceFile ())
		if err != nil { fmt.Println (err.Error ()) }
		defer file.Close ()

		scanner := bufio.NewScanner (file)
		for scanner.Scan () {
			line := scanner.Text ()

			// ignore blank lines and lines beginning with #
			if len (line) < 64 || line [0] == '#' {
				continue
			}

			// read only the first 64 characters, ignoring the rest of the line
			txId := line [0:64]
			txIdBytes, err := hex.DecodeString (txId)
			if err != nil {
				fmt.Println (err.Error ())
				continue
			}

			// write the JSON files
			tx := nodeClient.GetTx (hex.EncodeToString (txIdBytes))
			WriteTx (tx, app.Settings.GetTestDirectory ())
			fmt.Println (txId)
		}
		if err := scanner.Err (); err != nil {
			fmt.Println (err.Error ())
		}
	} else if testMode == "verify" {

		testDirectory := app.Settings.GetTestDirectory ()

		// get the files to read from
		files, err := os.ReadDir (testDirectory)
		if err != nil {
			fmt.Println (err.Error ())
			os.Exit (1)
		}

		// iterate through the transactions, getting data from the node to compare with the JSON file data
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
			tx := nodeClient.GetTx (hex.EncodeToString (txIdBytes))

			// verify it
			if VerifyTx (tx, testDirectory) {
				fmt.Println (txId, "OK")
			} else {
				fmt.Println (txId, "Failed")
			}
		}
	}
}

func WriteTx (tx btc.Tx, dir string) bool {
	nodeClient := btc.GetNodeClient ()

	blockHash := tx.GetBlockHash ()
	txId := tx.GetTxId ()
	txInputs := tx.GetInputs ()
	txOutputs := tx.GetOutputs ()

	txEncoded := txJson {	BlockHash: blockHash,
							BlockTime: tx.GetBlockTime (),
							TxId: txId,
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
		totalOut += txOutputs [o].GetValue ()
		txEncoded.Outputs [o] = encodeOutput (txOutputs [o])
	}

	totalIn := uint64 (0)
	for i := 0; i < len (txInputs); i++ {

//		inputScript := txInputs [i].GetInputScript ()
//		redeemScript := txInputs [i].GetRedeemScript ()
//		segwit := txInputs [i].GetSegwit ()
//		witnessScript := segwit.GetWitnessScript ()
//		tapScript, tapScriptIndex := segwit.GetTapScript ()

		txEncoded.Inputs [i] = encodeInput (txInputs [i])
/*
		inputEncoded := inputJson {	PreviousOutput: nil,
									InputScriptFields: nil,
									SegwitFields: segwit.GetFields (),
									RedeemScriptFields: redeemScript.GetFieldsAsHex (),
									WitnessScriptFields: witnessScript.GetFieldsAsHex (),
									TapScriptIndex: tapScriptIndex,
									TapScriptFields: tapScript.GetFieldsAsHex (),
									SpendType: txInputs [i].GetSpendType () }
*/

		if !txInputs [i].IsCoinbase () {
			// get the previous output
			previousOutput := nodeClient.GetPreviousOutput (txInputs [i].GetPreviousOutputTxId (), txInputs [i].GetPreviousOutputIndex ())
			outputEncoded := encodeOutput (previousOutput)

			totalIn += previousOutput.GetValue ()

			txEncoded.Inputs [i].PreviousOutput = &outputEncoded
//			inputEncoded.InputScriptFields = inputScript.GetFieldsAsHex ()
		} else {
			totalIn += totalOut

//			scriptFields := make ([] string, 1)
//			scriptFields [0] = inputScript.GetHex ()
//			inputEncoded.InputScriptFields = scriptFields
		}

//		txEncoded.Inputs [i] = inputEncoded
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
	err = os.WriteFile (dir + txId + ".json", jsonBytes, 0644)
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

	fileName := txId + ".json"

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
	if txEncoded.BlockHash != blockHash { fmt.Println ("Expecting block hash ", txEncoded.BlockHash, ", found ", blockHash); return false }
	if txEncoded.BlockTime != tx.GetBlockTime () { fmt.Println ("Expecting block time ", txEncoded.BlockTime, ", found ", tx.GetBlockTime ()); return false }
	if txEncoded.TxId != txId { fmt.Println ("Expecting tx id ", txEncoded.TxId, ", found ", txId); return false }
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
		totalOut += txOutputs [o].GetValue ()
		outputScript := txOutputs [o].GetOutputScript ()

//////////////////////////////////////////////////
		if txEncoded.Outputs [o].Value != txOutputs [o].GetValue () { fmt.Println ("Wrong number of outputs."); return false }
		if txEncoded.Outputs [o].OutputType != txOutputs [o].GetOutputType () { fmt.Println ("Wrong output type."); return false }
		if len (txEncoded.Outputs [o].OutputScriptFields) != len (outputScript.GetFieldsAsHex ()) { fmt.Println ("Wrong number of outputs script fields."); return false }
		for f := 0; f < len (txEncoded.Outputs [o].OutputScriptFields); f++ {
			if txEncoded.Outputs [o].OutputScriptFields [f].AsHex (0) != outputScript.GetFieldsAsHex () [f] { fmt.Println ("Wrong output script."); return false }
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
		witnessScript := segwit.GetWitnessScript ()
		tapScript, tapScriptIndex := segwit.GetTapScript ()

		inputEncoded := encodeInput (txInputs [i])
/*
		inputEncoded := inputJson {	PreviousOutput: nil,
									InputScriptFields: nil,
									SegwitFields: segwit.GetFields (),
									RedeemScriptFields: redeemScript.GetFieldsAsHex (),
									WitnessScriptFields: witnessScript.GetFieldsAsHex (),
									TapScriptIndex: tapScriptIndex,
									TapScriptFields: tapScript.GetFieldsAsHex (),
									SpendType: txInputs [i].GetSpendType () }
*/

		// segwit fields
		if len (txEncoded.Inputs [i].SegwitFields) != len (segwit.GetFields ()) { fmt.Println ("Wrong number of segwit fields."); return false }
		for f := 0; f < len (txEncoded.Inputs [i].SegwitFields); f++ {
			if txEncoded.Inputs [i].SegwitFields [f].AsHex (0) != segwit.GetFields () [f].AsHex (0) { fmt.Println ("Wrong segwit field."); return false }
		}

		// witness script
		if len (txEncoded.Inputs [i].WitnessScriptFields) != len (witnessScript.GetFieldsAsHex ()) { fmt.Println ("Wrong number of witness script fields."); return false }
		for f := 0; f < len (txEncoded.Inputs [i].WitnessScriptFields); f++ {
			if txEncoded.Inputs [i].WitnessScriptFields [f].AsHex (0) != witnessScript.GetFieldsAsHex () [f] { fmt.Println ("Wrong witness script field."); return false }
		}

		// tap script
		if txEncoded.Inputs [i].TapScriptIndex != tapScriptIndex { fmt.Println ("Wrong tap script index."); return false }
		if len (txEncoded.Inputs [i].TapScriptFields) != len (tapScript.GetFieldsAsHex ()) { fmt.Println ("Wrong number of tap script fields."); return false }
		for f := 0; f < len (txEncoded.Inputs [i].TapScriptFields); f++ {
			if txEncoded.Inputs [i].TapScriptFields [f].AsHex (0) != tapScript.GetFieldsAsHex () [f] { fmt.Println ("Wrong tap script field."); return false }
		}

		if txEncoded.Inputs [i].SpendType != txInputs [i].GetSpendType () { fmt.Println ("Wrong spend type."); return false }

		if !txInputs [i].IsCoinbase () {
			// previous output
			previousOutput := nodeClient.GetPreviousOutput (txInputs [i].GetPreviousOutputTxId (), txInputs [i].GetPreviousOutputIndex ())
			previousOutputScript := previousOutput.GetOutputScript ()

// same as above, could be a function
//////////////////////////////////////////////////
			if txEncoded.Inputs [i].PreviousOutput.Value != previousOutput.GetValue () { fmt.Println ("Wrong previous output value."); return false }
			if txEncoded.Inputs [i].PreviousOutput.OutputType != previousOutput.GetOutputType () { fmt.Println ("Wrong previous output type."); return false }
			if len (txEncoded.Inputs [i].PreviousOutput.OutputScriptFields) != len (previousOutputScript.GetFieldsAsHex ()) { fmt.Println ("Wrong number of previous output script fields."); return false }
			for f := 0; f < len (txEncoded.Inputs [i].PreviousOutput.OutputScriptFields); f++ {
				if txEncoded.Inputs [i].PreviousOutput.OutputScriptFields [f].AsHex (0) != previousOutputScript.GetFieldsAsHex () [f] { fmt.Println ("Wrong previous output script field."); return false }
			}
			if txEncoded.Inputs [i].PreviousOutput.ParseError != previousOutputScript.HasParseError () { fmt.Println ("Wrong previous output script status."); return false }
			if txEncoded.Inputs [i].PreviousOutput.Address != previousOutput.GetAddress () { fmt.Println ("Wrong previous output address."); return false }
//////////////////////////////////////////////////

			// input script
			if len (txEncoded.Inputs [i].InputScriptFields) != len (inputScript.GetFieldsAsHex ()) { fmt.Println ("Wrong input script field count."); return false }
			for f := 0; f < len (txEncoded.Inputs [i].InputScriptFields); f++ {
				if txEncoded.Inputs [i].InputScriptFields [f].AsHex (0) != inputScript.GetFieldsAsHex () [f] { fmt.Println ("Wrong input script field."); return false }
			}

			// redeem script
//fmt.Println (len (txEncoded.Inputs [i].RedeemScriptFields), len (redeemScript.GetFields ()))
			if len (txEncoded.Inputs [i].RedeemScriptFields) != len (redeemScript.GetFieldsAsHex ()) { fmt.Println ("Expected ", len (txEncoded.Inputs [i].RedeemScriptFields), " redeem script fields, found ", len (redeemScript.GetFields ())); return false }
			for f := 0; f < len (txEncoded.Inputs [i].RedeemScriptFields); f++ {
				if txEncoded.Inputs [i].RedeemScriptFields [f].AsHex (0) != redeemScript.GetFieldsAsHex () [f] { fmt.Println ("Wrong redeem script field."); return false }
			}

			totalIn += previousOutput.GetValue ()
		} else {
			totalIn += totalOut

			if len (txEncoded.Inputs [i].InputScriptFields) != 1 { fmt.Println ("Wrong input script field count."); return false }
			if txEncoded.Inputs [i].InputScriptFields [0].AsHex (0) != inputScript.GetHex () { fmt.Println ("Wrong input script field."); return false }
		}

		txEncoded.Inputs [i] = inputEncoded
	}

	if txEncoded.SatsIn != totalIn { fmt.Println ("Wrong value in."); return false }
	if txEncoded.Fee != totalIn - totalOut { fmt.Println ("Wrong fee."); return false }

	return true
}

