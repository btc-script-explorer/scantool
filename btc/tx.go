package btc

import (
	"encoding/hex"
//	"strings"
//	"strconv"

//	"btctx/themes"
)

type Tx struct {
	id [32] byte
	blockHash [32] byte
	blockTime int64
	version uint32
	coinbase bool
	bip141 bool
	inputs [] Input
	outputs [] Output
	lockTime uint32
}

func (tx *Tx) IsCoinbase () bool {
	return tx.coinbase
}

func (tx *Tx) GetTxId () [32] byte {
	return tx.id
}

func (tx *Tx) GetTxIdStr () string {
	return hex.EncodeToString (tx.id [:])
}

func (tx *Tx) GetBlockHash () [32] byte {
	return tx.blockHash
}

func (tx *Tx) GetBlockTime () int64 {
	return tx.blockTime
}

func (tx *Tx) GetVersion () uint32 {
	return tx.version
}

func (tx *Tx) SupportsBip141 () bool {
	return tx.bip141
}

func (tx *Tx) GetInputs () [] Input {
	return tx.inputs
}

func (tx *Tx) GetOutputs () [] Output {
	return tx.outputs
}

func (tx *Tx) GetLockTime () uint32 {
	return tx.lockTime
}

func (tx *Tx) GetHtmlData () map [string] interface {} {

	boxWidths := uint16 (112)

	htmlData := make (map [string] interface {})

	// transaction data
	htmlData ["IsCoinbase"] = tx.coinbase
	htmlData ["SupportsBip141"] = tx.bip141
	htmlData ["LockTime"] = tx.lockTime

	// outputs
	totalOut := uint64 (0)
	outputCount := len (tx.outputs)

	htmlData ["OutputCount"] = outputCount
	outputCountLabel := "Output"; if outputCount > 1 { outputCountLabel += "s" }
	htmlData ["OutputCountLabel"] = outputCountLabel

	outputHtmlData := make ([] OutputHtmlData, outputCount)
	for o := uint32 (0); o < uint32 (outputCount); o++ {
		totalOut += tx.outputs [o].GetSatoshis ()
		outputHtmlData [o] = tx.outputs [o].GetHtmlData (o, true, boxWidths)
	}
	htmlData ["OutputData"] = outputHtmlData

	// totals for the transaction
	htmlData ["ValueOut"] = totalOut
	htmlData ["ValueIn"] = 0; if tx.coinbase { htmlData ["ValueIn"] = totalOut }
	htmlData ["Fee"] = 0

	// inputs
	inputCount := len (tx.inputs)

	htmlData ["InputCount"] = inputCount
	inputCountLabel := "Input"; if inputCount > 1 { inputCountLabel += "s" }
	htmlData ["InputCountLabel"] = inputCountLabel

	inputHtmlData := make ([] InputHtmlData, inputCount)
	for i := uint32 (0); i < uint32 (inputCount); i++ {
		valueIn := uint64 (0); if tx.coinbase && i == 0 { valueIn = totalOut }
		inputHtmlData [i] = tx.inputs [i].GetHtmlData (i, valueIn, tx.bip141, boxWidths)
	}
	htmlData ["InputData"] = inputHtmlData

	return htmlData
}

type PendingInput struct {
	Tx_id string
	Input_index uint32

	Prev_out_tx_id string
	Prev_out_index uint32

	Tap_script_index int64
}

func (tx *Tx) GetPendingInputs () [] PendingInput {
	inputCount := len (tx.inputs)
	if tx.coinbase {
		inputCount = 0
	}

	pendingInputs := make ([] PendingInput, inputCount)
	for i := uint32 (0); i < uint32 (inputCount); i++ {
		previousOutputTxId := tx.inputs [i].GetPreviousOutputTxId ()
		pendingInputs [i] = PendingInput { Tx_id: tx.GetTxIdStr (), Input_index: i, Prev_out_tx_id: hex.EncodeToString (previousOutputTxId [:]), Prev_out_index: tx.inputs [i].GetPreviousOutputIndex () }

		segwit := tx.inputs [i].GetSegwit ()
		if !segwit.IsNil () {
			tapScript, tapScriptIndex := segwit.GetTapScript ()
			if !tapScript.IsNil () {
				pendingInputs [i].Tap_script_index = tapScriptIndex
			}
		}
	}

	return pendingInputs
}

/*
// This function can be used to read a raw transaction as a byte array.
// This method has been abandoned because it does not include bitcoin addresses.
// However, it is still included here, commented out, in case it becomes more
// convenient to read transactions this way if/when other bitcoin node types are supported.
func NewTx (hash string, raw_bytes [] byte) Tx {

	value_reader := ValueReader {}

	pos := 0

	version := value_reader.ReadNumeric (raw_bytes [pos : pos + 4])
	pos += 4


	// check for segwit support
	input_count, byte_count := value_reader.ReadVarInt (raw_bytes [pos:])
	pos += byte_count

	bip_141 := input_count == 0
	if bip_141 {
//		bip_141_flag := value_reader.ReadNumeric (raw_bytes [pos : pos + 1])
		pos += 1

		input_count, byte_count = value_reader.ReadVarInt (raw_bytes [pos:]);
		pos += byte_count
	}

	// inputs
	inputs := make ([] Input, input_count)
	for i := 0; i < int (input_count); i++ {
		input, byte_count := NewInput (raw_bytes [pos:])
		inputs [i] = input
		pos += byte_count
	}

	coinbase := inputs [0].coinbase

	// outputs
	output_count, byte_count := value_reader.ReadVarInt (raw_bytes [pos:])
	pos += byte_count
	
	outputs := make ([] Output, output_count)
	for o := 0; o < int (output_count); o++ {
		output, byte_count := NewOutput (raw_bytes [pos:])
		outputs [o] = output
		pos += byte_count
	}

	// segwit
	if bip_141 {
		for i := 0; i < int (input_count); i++ {
			segwit, byte_count := NewSegwit (raw_bytes [pos:])
			pos += byte_count

			if !segwit.IsEmpty () {
				inputs [i].SetSegwit (segwit)
			}
		}
	}

	// serialized scripts
	for i := 0; i < int (input_count); i++ {
		inputs [i].ParseSerializedScripts ()
	}

	lock_time := value_reader.ReadNumeric (raw_bytes [pos : pos + 4])
	pos += 4

	hash_bytes, _ := hex.DecodeString (hash)

	return Tx { hash: [32] byte (hash_bytes),
		version: uint32 (version),
		coinbase: coinbase,
		bip141: bip_141,
		inputs: inputs,
		outputs: outputs,
		lock_time: uint32 (lock_time) }
}
*/
