package btc

import (
	"fmt"
	"html/template"
)

type Output struct {
	value uint64
	outputScript Script
	outputType string
	address string
}

func NewOutput (value uint64, script Script, address string) Output {

	// determine the output type
	outputType := ""
	if script.IsTaprootOutput () { outputType = "Taproot" } else
	if script.IsP2wpkhOutput () { outputType = "P2WPKH" } else
	if script.IsP2wshOutput () { outputType = "P2WSH" } else
	if script.IsP2shOutput () { outputType = "P2SH" } else
	if script.IsP2pkhOutput () { outputType = "P2PKH" } else
	if script.IsMultiSigOutput () { outputType = "MultiSig" } else
	if script.IsP2pkOutput () { outputType = "P2PK" } else
	if script.IsNullDataOutput () { outputType = "OP_RETURN" } else
	if script.IsWitnessUnknownOutput () { outputType = "Witness Unknown" } else
	{ outputType = "Non-Standard" }

	o := Output { value: value, outputScript: script, outputType: outputType, address: address }
	o.setFieldTypes ()

	return o
}

func (o *Output) setFieldTypes () {

	if o.outputType == "Taproot" { outputType := [...] string { "OP_1", "32-Byte Witness Program" }; o.outputScript.SetFieldTypes (outputType [:]) } else
	if o.outputType == "P2WSH" { outputType := [...] string { "OP_0", "32-Byte Witness Program" }; o.outputScript.SetFieldTypes (outputType [:]) } else
	if o.outputType == "P2WPKH" { outputType := [...] string { "OP_0", "20-Byte Witness Program" }; o.outputScript.SetFieldTypes (outputType [:]) } else
	if o.outputType == "P2SH" { outputType := [...] string { "OP_HASH160", "20-Byte Script Hash", "OP_EQUAL" }; o.outputScript.SetFieldTypes (outputType [:]) } else
	if o.outputType == "P2PKH" { outputType := [...] string { "OP_DUP", "OP_HASH160", "20-Byte Key Hash", "OP_EQUALVERIFY", "OP_CHECKSIG" }; o.outputScript.SetFieldTypes (outputType [:]) } else
	if o.outputType == "P2PK" { outputType := [...] string { GetStackItemType (o.outputScript.GetFields () [0], false, false), "OP_CHECKSIG" }; o.outputScript.SetFieldTypes (outputType [:]) } else
	if o.outputType == "MultiSig" || o.outputType == "OP_RETURN" || o.outputType == "Witness Unknown" || o.outputType == "Non-Standard" {
		hexFields := o.outputScript.GetFields ()
		rawFieldTypes := o.outputScript.GetRawFieldTypes ()
		fieldCount := len (hexFields)
		fieldTypes := make ([] string, fieldCount)
		for f, hexField := range hexFields {
			if rawFieldTypes [f] == 'o' {
				fieldTypes [f] = hexField
				continue
			}

			fieldTypes [f] = GetStackItemType (hexField, false, false)
		}
		o.outputScript.SetFieldTypes (fieldTypes)
	} else {
		fmt.Println ("Unknown output type ", o.outputType)
		outputType := [...] string {}
		o.outputScript.SetFieldTypes (outputType [:])
	}
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

type OutputHtmlData struct {
	WidthCh uint16
	BoxTitle string
	OutputIndex uint32
	OutputType string
	Value template.HTML
	Address string
	OutputScript ScriptHtmlData
}

func (o *Output) GetHtmlData (scriptHtmlId string, boxTitle string, outputIndex uint32, widthCh uint16) OutputHtmlData {

	address := o.address
	if len (address) == 0 { address = "No Address Format" }

	return OutputHtmlData { WidthCh: widthCh, BoxTitle: boxTitle, OutputIndex: outputIndex, OutputType: o.outputType, Value: template.HTML (GetValueHtml (o.value)), Address: address, OutputScript: o.outputScript.GetHtmlData ("Output Script", scriptHtmlId, widthCh - 6, "hex") }
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
