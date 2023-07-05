package btc

import (
	"strconv"
	"strings"

	"btctx/themes"
)

type Output struct {
	value uint64
	outputScript Script
	outputType string
	address string
}

func (o *Output) GetSatoshis () uint64 {
	return o.value
}

func (o *Output) GetOutputScript () Script {
	return o.outputScript
}

func (o *Output) GetOutputType () string {
	return o.outputType
}

func (o *Output) GetAddress () string {
	return o.address
}

func (o *Output) GetHtml (outputIndex int, theme themes.Theme, minimized bool) string {

	html := theme.GetOutputHtmlTemplate (minimized)

	html = strings.Replace (html, "[[INDEX]]", strconv.Itoa (outputIndex), 1)
	html = strings.Replace (html, "[[TYPE]]", o.outputType, 1)
	html = strings.Replace (html, "[[VALUE]]", GetValueHtml (o.value), 1)

	address := o.address
	if len (address) == 0 { address = "No Address Format" }
	html = strings.Replace (html, "[[ADDRESS]]", address, 1)

	scriptFields := ""
	if !o.outputScript.IsEmpty () {
		fields := o.outputScript.GetFields ()
		for _, field := range fields {
			scriptFields += "<div>" + field + "</div>"
		}
	} else {
		scriptFields = "Empty"
	}
	html = strings.Replace (html, "[[OUTPUT-SCRIPT]]", scriptFields, 1)

	return html
}

/*
// This function can be used to read a raw transaction as a byte array.
// This method has been abandoned because it does not include bitcoin addresses.
// However, it is still included here, commented out, in case it becomes more
// convenient to read transactions this way if/when other bitcoin node types are supported.
func NewOutput (raw_bytes [] byte) (Output, int) {

	value_reader := ValueReader {}

	pos := 0

	value := value_reader.ReadNumeric (raw_bytes [pos : pos + 8])
	pos += 8

	script_len, byte_count := value_reader.ReadVarInt (raw_bytes [pos:])
	pos += byte_count

	script, byte_count := NewScript (raw_bytes [pos : pos + int (script_len)])
	pos += byte_count

	output_type := ""
	if script.IsTaprootOutput () { output_type = "Taproot"
	} else if script.IsP2wpkhOutput () { output_type = "P2WPKH"
	} else if script.IsP2wshOutput () { output_type = "P2WSH"
	} else if script.IsP2shOutput () { output_type = "P2SH"
	} else if script.IsP2pkhOutput () { output_type = "P2PKH"
	} else if script.IsMultiSigOutput () { output_type = "MultiSig"
	} else if script.IsP2pkOutput () { output_type = "P2PK"
	} else if script.IsNullDataOutput () { output_type = "OP_RETURN"
	} else if script.IsWitnessUnknownOutput () { output_type = "Witness Unknown"
	} else if script.IsNonstandardOutput () { output_type = "Non-Standard" }

	return Output { value: value,
					script: script,
					output_type: output_type,
					address: "" }, pos
}
*/
