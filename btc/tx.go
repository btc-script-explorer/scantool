package btc

import (
	"time"
	"strconv"

	"btctx/app"
)

type Tx struct {
	id string
	blockHeight uint32
	blockTime int64
	blockHash string
	version uint32
	coinbase bool
	bip141 bool
	inputs [] Input
	outputs [] Output
	lockTime uint32
}

func (tx *Tx) IsNil () bool {
	return len (tx.id) == 0
}

func (tx *Tx) IsCoinbase () bool {
	return tx.coinbase
}

func (tx *Tx) GetTxId () string {
	return tx.id
}

func (tx *Tx) GetBlockHash () string {
	return tx.blockHash
}

func (tx *Tx) GetBlockHeight () uint32 {
	return tx.blockHeight
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

func (tx *Tx) GetInputCount () int {
	return len (tx.inputs)
}

func (tx *Tx) GetInputs () [] Input {
	return tx.inputs
}

func (tx *Tx) GetOutputCount () int {
	return len (tx.outputs)
}

func (tx *Tx) GetOutputs () [] Output {
	return tx.outputs
}

func (tx *Tx) GetLockTime () uint32 {
	return tx.lockTime
}

func (tx *Tx) GetHtmlData () map [string] interface {} {

	htmlData := make (map [string] interface {})

	// transaction data
	settings := app.GetSettings ()
	htmlData ["BaseUrl"] = "http://" + settings.Website.GetFullUrl ()
	htmlData ["BlockHeight"] = tx.blockHeight
	htmlData ["BlockTime"] = time.Unix (tx.blockTime, 0).UTC ()
	htmlData ["BlockHash"] = tx.blockHash
	htmlData ["IsCoinbase"] = tx.coinbase
	htmlData ["SupportsBip141"] = tx.bip141
	htmlData ["LockTime"] = tx.lockTime

	// outputs
	totalOut := uint64 (0)
	outputCount := len (tx.outputs)

	outputCountLabel := strconv.Itoa (outputCount) + " Output"; if outputCount > 1 { outputCountLabel += "s" }
	htmlData ["OutputCountLabel"] = outputCountLabel

	outputHtmlData := make ([] OutputHtmlData, outputCount)
	for o := uint32 (0); o < uint32 (outputCount); o++ {
		totalOut += tx.outputs [o].GetSatoshis ()
		scriptHtmlId := "output-script-" + strconv.FormatUint (uint64 (o), 10)
		outputHtmlData [o] = tx.outputs [o].GetHtmlData (scriptHtmlId, "", o)
	}
	htmlData ["OutputData"] = outputHtmlData

	// totals for the transaction
	htmlData ["ValueOut"] = totalOut
	htmlData ["ValueIn"] = 0; if tx.coinbase { htmlData ["ValueIn"] = totalOut }
	htmlData ["Fee"] = 0

	// inputs
	inputCount := len (tx.inputs)

	inputCountLabel := strconv.Itoa (inputCount) + " Input"; if inputCount > 1 { inputCountLabel += "s" }
	htmlData ["InputCountLabel"] = inputCountLabel

	inputHtmlData := make ([] InputHtmlData, inputCount)
	for i := uint32 (0); i < uint32 (inputCount); i++ {
		valueIn := uint64 (0); if tx.coinbase && i == 0 { valueIn = totalOut }
		inputHtmlData [i] = tx.inputs [i].GetHtmlData (i, valueIn, tx.bip141)
	}
	htmlData ["InputData"] = inputHtmlData

	return htmlData
}

type PendingInput struct {
	Tx_id string
	Input_index uint32

	Prev_out_tx_id string
	Prev_out_index uint32
}

func (tx *Tx) GetPendingInputs () [] PendingInput {

	if tx.coinbase {
		return [] PendingInput {}
	}

	inputCount := len (tx.inputs)

	pendingInputs := make ([] PendingInput, inputCount)
	for i := uint32 (0); i < uint32 (inputCount); i++ {
		previousOutputTxId := tx.inputs [i].GetPreviousOutputTxId ()
		pendingInputs [i] = PendingInput { Tx_id: tx.GetTxId (), Input_index: i, Prev_out_tx_id: previousOutputTxId, Prev_out_index: tx.inputs [i].GetPreviousOutputIndex () }
	}

	return pendingInputs
}

