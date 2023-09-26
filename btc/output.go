package btc

import (
	"fmt"
)

type PreviousOutputRequest struct {
	TxId string
	InputIndex uint32
	PreviousTxId string
	PreviousOutputIndex uint32
}

const OUTPUT_TYPE_P2PK = "P2PK"
const OUTPUT_TYPE_MultiSig = "MultiSig"
const OUTPUT_TYPE_P2PKH = "P2PKH"
const OUTPUT_TYPE_P2SH = "P2SH"
const OUTPUT_TYPE_P2WPKH = "P2WPKH"
const OUTPUT_TYPE_P2WSH = "P2WSH"
const OUTPUT_TYPE_TAPROOT = "Taproot"
const OUTPUT_TYPE_OP_RETURN = "OP_RETURN"
const OUTPUT_TYPE_WitnessUnknown = "Witness Unknown"
const OUTPUT_TYPE_NonStandard = "Non-Standard"

type Output struct {
	value uint64
	outputScript Script
	outputType string
	address string
}

func NewOutput (value uint64, script Script, address string) Output {

	// determine the output type
	outputType := ""
	if script.IsTaprootOutput () { outputType = OUTPUT_TYPE_TAPROOT } else
	if script.IsP2wpkhOutput () { outputType = OUTPUT_TYPE_P2WPKH } else
	if script.IsP2wshOutput () { outputType = OUTPUT_TYPE_P2WSH } else
	if script.IsP2shOutput () { outputType = OUTPUT_TYPE_P2SH } else
	if script.IsP2pkhOutput () { outputType = OUTPUT_TYPE_P2PKH } else
	if script.IsMultiSigOutput () { outputType = OUTPUT_TYPE_MultiSig } else
	if script.IsP2pkOutput () { outputType = OUTPUT_TYPE_P2PK } else
	if script.IsNullDataOutput () { outputType = OUTPUT_TYPE_OP_RETURN } else
	if script.IsWitnessUnknownOutput () { outputType = OUTPUT_TYPE_WitnessUnknown } else
	{ outputType = OUTPUT_TYPE_NonStandard }

	o := Output { value: value, outputScript: script, outputType: outputType, address: address }
	o.setFieldTypes ()

	return o
}

func (o *Output) setFieldTypes () {

	if o.outputType == OUTPUT_TYPE_TAPROOT {
//		o.outputScript.SetFieldType (0, "OP_1")
		o.outputScript.SetFieldType (1, "Witness Program (Public Key)")
	} else if o.outputType == OUTPUT_TYPE_P2WSH {
//		o.outputScript.SetFieldType (0, "OP_0")
		o.outputScript.SetFieldType (1, "Witness Program (Script Hash)")
	} else if o.outputType == OUTPUT_TYPE_P2WPKH {
//		o.outputScript.SetFieldType (0, "OP_0")
		o.outputScript.SetFieldType (1, "Witness Program (Public Key Hash)")
	} else if o.outputType == OUTPUT_TYPE_P2SH {
//		o.outputScript.SetFieldType (0, "OP_HASH160")
		o.outputScript.SetFieldType (1, "Script Hash")
//		o.outputScript.SetFieldType (2, "OP_EQUAL")
	} else if o.outputType == OUTPUT_TYPE_P2PKH {
//		o.outputScript.SetFieldType (0, "OP_DUP")
//		o.outputScript.SetFieldType (1, "OP_HASH160")
		o.outputScript.SetFieldType (2, "Public Key Hash")
//		o.outputScript.SetFieldType (3, "OP_EQUALVERIFY")
//		o.outputScript.SetFieldType (4, "OP_CHECKSIG")
	} else if o.outputType != OUTPUT_TYPE_P2PK && o.outputType != OUTPUT_TYPE_MultiSig && o.outputType != OUTPUT_TYPE_OP_RETURN && o.outputType != OUTPUT_TYPE_WitnessUnknown && o.outputType != OUTPUT_TYPE_NonStandard {
		fmt.Println ("Unknown output type ", o.outputType)
	}
}

func (o *Output) GetValue () uint64 {
	return o.value
}

func (o *Output) GetOutputScript () Script {
	return o.outputScript
}

func (o *Output) GetOutputType () string {
	return o.outputType
}

func (o *Output) GetAddress () string {
	if len (o.address) == 0 { return "No Address Format" }
	return o.address
}

