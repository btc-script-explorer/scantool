package btc

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Input struct {
	coinbase bool
	previousOutputTxId [32] byte
	previousOutputIndex uint32
	spendType string
	inputScript Script
	redeemScript Script
	segwit Segwit
	sequence uint32
}

func (i *Input) HasRedeemScript () bool {
	return i.redeemScript.GetFields () != nil
}

func (i *Input) GetInputScript () Script {
	return i.inputScript
}

func (i *Input) GetRedeemScript () Script {
	return i.redeemScript
}

func (i *Input) GetSegwit () Segwit {
	return i.segwit
}

func (i *Input) IsCoinbase () bool {
	return i.coinbase
}

func (i *Input) GetPreviousOutputTxId () [32] byte {
	return i.previousOutputTxId
}

func (i *Input) GetPreviousOutputIndex () uint32 {
	return i.previousOutputIndex
}

func (i *Input) GetSpendType () string {
	return i.spendType
}

func (i *Input) GetSequence () uint32 {
	return i.sequence
}

func (i *Input) getPendingHTMLTemplate () string {
	fileBytes, err := os.ReadFile ("./html/input-minimized.html")
	if err != nil {
		fmt.Println (err.Error ())
	}
	return string (fileBytes)
}

func (i *Input) GetHTML (inputIndex int, satoshis uint64) string {

	html := i.getPendingHTMLTemplate ()

	html = strings.Replace (html, "[[INDEX]]", strconv.Itoa (inputIndex), -1)
	html = strings.Replace (html, "[[TX-TYPE]]", i.spendType, 1)

	inputValue := ""
	if i.IsCoinbase () && satoshis > 0 { inputValue = strconv.FormatUint (satoshis, 10) }
	html = strings.Replace (html, "[[INPUT-VALUE]]", inputValue, 1)

	return html
}

