package btc

import (
	"fmt"
	"os"
	"encoding/hex"
	"strings"
	"strconv"
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

func getTxHTMLTemplate () string {
	fileBytes, err := os.ReadFile ("./html/tx.html")
	if err != nil {
		fmt.Println (err.Error ())
		return ""
	}
	return string (fileBytes)
}

func (tx *Tx) IsCoinbase () bool {
	return tx.coinbase
}

func (tx *Tx) GetTxId () [32] byte {
	return tx.id
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

func (tx *Tx) GetHTML () string {
	html := getTxHTMLTemplate ()

	html = strings.Replace (html, "[[TX-HASH]]", hex.EncodeToString (tx.id [:]), 1)

	coinbase := "No"
	if tx.coinbase { coinbase = "Yes" }
	html = strings.Replace (html, "[[TX-COINBASE]]", coinbase, 1)

	bip141 := "No"
	if tx.bip141 { bip141 = "Yes" }
	html = strings.Replace (html, "[[TX-BIP-141]]", bip141, 1)

	html = strings.Replace (html, "[[TX-LOCK-TIME]]", strconv.FormatInt (int64 (tx.lockTime), 10), 1)

	// outputs
	outputCount := len (tx.outputs)
	totalOut := uint64 (0)
	for o := 0; o < outputCount; o++ {
		totalOut += tx.outputs [o].GetSatoshis ()
	}
	html = strings.Replace (html, "[[TX-VALUE-OUT]]", strconv.FormatUint (totalOut, 10), 1)

	outputCountLabel := strconv.Itoa (outputCount) + " Output"
	if outputCount > 1 { outputCountLabel += "s" }
	html = strings.Replace (html, "[[TX-OUTPUT-COUNT]]", outputCountLabel, 1)

	outputsHtml := ""
	for o := 0; o < len (tx.outputs); o++ {
		outputsHtml += tx.outputs [o].GetMinimizedHTML (o)
	}
	html = strings.Replace (html, "[[TX-OUTPUTS]]", outputsHtml, 1)

	// inputs
	// these are set to zero because the previous outputs will be read asyncronously
	if tx.coinbase {
		html = strings.Replace (html, "[[TX-VALUE-IN]]", strconv.FormatUint (totalOut, 10), 1)
	} else {
		html = strings.Replace (html, "[[TX-VALUE-IN]]", "0", 1)
	}
	html = strings.Replace (html, "[[TX-FEE]]", "0", 1)

	inputCount := len (tx.inputs)
	inputCountLabel := strconv.Itoa (inputCount) + " Input"
	if inputCount > 1 { inputCountLabel += "s" }
	html = strings.Replace (html, "[[TX-INPUT-COUNT]]", inputCountLabel, 1)

	inputsHtml := ""
	for i := 0; i < len (tx.inputs); i++ {
		if tx.coinbase && i == 0 {
			inputsHtml += tx.inputs [i].GetHTML (i, totalOut)
		} else {
			inputsHtml += tx.inputs [i].GetHTML (i, 0)
		}
	}
	html = strings.Replace (html, "[[TX-INPUTS]]", inputsHtml, 1)

	return html
}

type PendingInput struct {
	Input_index int
	Prev_tx_id string
	Output_index uint32
}

func (tx *Tx) GetPendingInputs () [] PendingInput {
	inputCount := len (tx.inputs)
	if tx.coinbase {
		inputCount = 0
	}

	pendingInputs := make ([] PendingInput, inputCount)
	for i := 0; i < inputCount; i++ {
		previousOutputTxId := tx.inputs [i].GetPreviousOutputTxId ()
		pendingInputs [i] = PendingInput { Input_index: i, Prev_tx_id: hex.EncodeToString (previousOutputTxId [:]), Output_index: tx.inputs [i].GetPreviousOutputIndex () }
	}

	return pendingInputs
}

