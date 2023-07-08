package btc

import (
	"encoding/hex"
	"strconv"
	"strings"

	"btctx/themes"
)

type Input struct {
	coinbase bool
	spendType string
	inputScript Script
	segwit Segwit
	sequence uint32

	previousOutputTxId [32] byte
	previousOutputIndex uint32
	redeemScript Script
}

func (i *Input) GetInputScript () Script {
	return i.inputScript
}

func (i *Input) HasRedeemScript () bool {
	return i.redeemScript.GetFields () != nil
}

func (i *Input) GetRedeemScript () Script {
	return i.redeemScript
}

func (i *Input) HasSegwit () bool {
	return !i.segwit.IsNil () && !i.segwit.IsEmpty ()
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

func (i *Input) GetHtml (inputIndex int, satoshis uint64, theme themes.Theme, minimized bool) string {

	html := theme.GetInputHtmlTemplate (minimized)

	html = strings.Replace (html, "[[INPUT-INDEX]]", strconv.Itoa (inputIndex), -1)
	html = strings.Replace (html, "[[INPUT-SPEND-TYPE]]", i.spendType, 1)

	if minimized {
		inputValue := ""
		if i.IsCoinbase () && satoshis > 0 { inputValue = GetValueHtml (satoshis) }
		html = strings.Replace (html, "[[INPUT-VALUE]]", inputValue, 1)
	} else {
		outpoint := "N/A"
		if !i.IsCoinbase () { outpoint = hex.EncodeToString (i.previousOutputTxId [:]) + " : " + strconv.Itoa (int (i.previousOutputIndex)) }

		html = strings.Replace (html, "[[INPUT-OUTPOINT]]", outpoint, 1)
		html = strings.Replace (html, "[[INPUT-SEQUENCE]]", strconv.Itoa (int (i.sequence)), 1)

		// previous output is deferred until later
		html = strings.Replace (html, "[[INPUT-PREVIOUS-OUTPUT-HTML]]", "", 1)

		// segwit, if there is one
		if i.HasSegwit () {
			segwitHtml := i.segwit.GetHtml (theme)
			html = strings.Replace (html, "[[INPUT-SEGWIT-HTML]]", segwitHtml, 1)
		} else {
			html = strings.Replace (html, "[[INPUT-SEGWIT-HTML]]", "", 1)
		}

		// input script
		scriptHtml := ""
		if !i.IsCoinbase () {
			scriptHtml = i.inputScript.GetHtml ("Input Script", theme)
		} else {
			scriptHtml = i.inputScript.GetTextHtml ("Coinbase Script Text", theme)
		}
		html = strings.Replace (html, "[[INPUT-SCRIPT-HTML]]", scriptHtml, 1)

		// redeem script, if there is one
		if i.HasRedeemScript () {
			redeemScriptHtml := i.redeemScript.GetHtml ("Redeem Script", theme)
			html = strings.Replace (html, "[[INPUT-REDEEM-SCRIPT-HTML]]", redeemScriptHtml, 1)
		} else {
			html = strings.Replace (html, "[[INPUT-REDEEM-SCRIPT-HTML]]", "", 1)
		}
	}

	return html
}

/*
// This function can be used to read a raw transaction as a byte array.
// This method has been abandoned because it does not include bitcoin addresses.
// However, it is still included here, commented out, in case it becomes more
// convenient to read transactions this way if/when other bitcoin node types are supported.
func NewInput (raw_bytes [] byte) (Input, int) {

	value_reader := ValueReader {}

	pos := 0

	prev_out_hash := value_reader.ReverseBytes (raw_bytes [pos : pos + 32])
	pos += 32

	prev_out_index := value_reader.ReadNumeric (raw_bytes [pos : pos + 4])
	pos += 4

	coinbase := true
	if hex.EncodeToString (prev_out_hash) != "0000000000000000000000000000000000000000000000000000000000000000" {
		coinbase = false
	}

	script_len, byte_count := value_reader.ReadVarInt (raw_bytes [pos:])
	pos += byte_count

	script, byte_count := NewScript (raw_bytes [pos : pos + int (script_len)])
	pos += byte_count

	sequence := value_reader.ReadNumeric (raw_bytes [pos : pos + 4])
	pos += 4

	return Input {
		coinbase: coinbase,
		prev_out_hash: [32] byte (prev_out_hash),
		prev_out_index: uint32 (prev_out_index),
		tx_type: "",
		script: script,
		has_redeem_script: false,
		has_segwit_fields: false,
		sequence: uint32 (sequence) }, pos
}

// attempt to parse the serialized script(s) without knowing the output type
// SetSegwit must be called first
func (i *Input) ParseSerializedScripts () {

	// inputs with a previous p2sh output type have redeem scripts
	// non-segwit p2sh inputs have no segwit
	if i.script.IsP2shP2wshInput () || i.script.IsP2shP2wpkhInput () || !i.has_segwit_fields {
		redeem_script, _ := NewScript (i.script.GetSerializedScript ())
		if !redeem_script.parse_error {
			i.tx_type = "P2SH"
			if i.script.IsP2shP2wshInput () {
				i.tx_type += "-P2WSH"
			} else if i.script.IsP2shP2wpkhInput () {
				i.tx_type += "-P2WSH"
			}
			i.has_redeem_script = true
			i.redeem_script = redeem_script
		}
	}

	// p2sh-p2wsh and p2wsh inputs have a witness script
	// Taproot Script Path inputs have a tap script
	// Taproot and p2wsh inputs have an empty input script
	if i.script.IsP2shP2wshInput () || i.script.IsEmpty () {
		if i.segwit.ParseSerializedScript () {
			if i.script.IsP2shP2wshInput () {
				i.tx_type += "P2SH-P2WSH"
			} else {
				i.tx_type += "Taproot Script Path"
			}
		}
	}
}

*/
