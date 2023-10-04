package btc

import (
//	"fmt"
)

//const SPEND_TYPE_P2PK = "P2PK"
//const SPEND_TYPE_MultiSig = "MultiSig"
//const SPEND_TYPE_P2PKH = "P2PKH"
//const SPEND_TYPE_P2SH = "P2SH"
const SPEND_TYPE_P2SH_P2WPKH = "P2SH-P2WPKH"
const SPEND_TYPE_P2SH_P2WSH = "P2SH-P2WSH"
//const SPEND_TYPE_P2WPKH = "P2WPKH"
//const SPEND_TYPE_P2WSH = "P2WSH"
const SPEND_TYPE_P2TR_Key = "Taproot Key Path"
const SPEND_TYPE_P2TR_Script = "Taproot Script Path"
const SPEND_TYPE_NonStandard = "Non-Standard"

type Input struct {
	coinbase bool
	previousOutputTxId string
	previousOutputIndex uint16
	inputScript Script
	sequence uint32

	redeemScript Script
	segwit Segwit

	previousOutput Output
	spendType string
}

func NewInput (coinbase bool, previousOutputTxId string, previousOutputIndex uint16, inputScript Script, segwit Segwit, sequence uint32, previousOutput Output) Input {

	i := Input { coinbase: coinbase, previousOutputTxId: previousOutputTxId, previousOutputIndex: previousOutputIndex, inputScript: inputScript, segwit: segwit, sequence: sequence }

	if i.coinbase {
		i.spendType = "COINBASE"
	} else {
		i.SetPreviousOutput (previousOutput)
	}

	return i
}

func (i *Input) SetPreviousOutput (previousOutput Output) {

	// reset everything
	i.spendType = ""
	i.segwit.DeleteWitnessScript ()
	i.segwit.DeleteTapScript ()
	for f, _ := range i.segwit.fields {
		i.segwit.fields [f].SetType ("")
	}

	// re-evaluate the input based on the new previous output
	i.previousOutput = previousOutput
	previousOutputType := previousOutput.GetOutputType ()
	if previousOutputType == OUTPUT_TYPE_P2PK || previousOutputType == OUTPUT_TYPE_MultiSig || previousOutputType == OUTPUT_TYPE_P2PKH || previousOutputType == OUTPUT_TYPE_P2WPKH {
		i.spendType = previousOutputType
	} else {

		i.spendType = SPEND_TYPE_NonStandard

		switch previousOutputType {

			case OUTPUT_TYPE_P2SH:

				isP2shWrappedType := !i.segwit.IsEmpty () && !i.inputScript.IsEmpty ()
				if isP2shWrappedType {

					// the only two spend types that have segwit fields and also have a non-empty input script are the p2sh-wrapped spend types
					i.redeemScript = i.inputScript.GetSerializedScript ()
					if !i.redeemScript.IsNil () {
						if i.redeemScript.IsP2shP2wpkhRedeemScript () {
							i.spendType = SPEND_TYPE_P2SH_P2WPKH
						} else if i.redeemScript.IsP2shP2wshRedeemScript () {
							i.spendType = SPEND_TYPE_P2SH_P2WSH
							i.segwit.SetWitnessScript (i.segwit.parseWitnessScript ())
						}
					}

					i.inputScript.SetFieldType (i.inputScript.GetParsedFieldCount () - 1, "SERIALIZED REDEEM SCRIPT")

					i.redeemScript.SetFieldType (0, "OP_0")
					if i.spendType == SPEND_TYPE_P2SH_P2WPKH { i.redeemScript.SetFieldType (1, "Witness Program (Public Key Hash)") } else 
					if i.spendType == SPEND_TYPE_P2SH_P2WSH { i.redeemScript.SetFieldType (1, "Witness Program (Script Hash)") }

				} else {

					i.spendType = OUTPUT_TYPE_P2SH
					redeemScript := i.inputScript.GetSerializedScript ()
					i.SetRedeemScript (redeemScript)
					if !i.inputScript.IsEmpty () {
						i.inputScript.SetFieldType (i.inputScript.GetParsedFieldCount () - 1, "SERIALIZED REDEEM SCRIPT")

						// check for a zero-length redeem script
						inputScriptFields := i.inputScript.GetFields ()
						serializedScriptIndex := len (inputScriptFields) - 1
						serializedScriptBytes := inputScriptFields [serializedScriptIndex].AsBytes ()
						if len (serializedScriptBytes) == 1 && serializedScriptBytes [0] == 0x00 {
							inputScriptFields [serializedScriptIndex].SetIsOpcode (false)
							inputScriptFields [serializedScriptIndex].SetBytes ([] byte {})
							i.SetRedeemScript (NewScript ([] byte {}))
						}
					}
				}

			case OUTPUT_TYPE_P2WSH:

				if i.segwit.IsValidP2wsh () {
					i.spendType = OUTPUT_TYPE_P2WSH
					i.segwit.SetWitnessScript (i.segwit.parseWitnessScript ())
				}

			case OUTPUT_TYPE_TAPROOT:

				if i.segwit.IsValidTaprootKeyPath () {
					i.spendType = SPEND_TYPE_P2TR_Key
				} else if i.segwit.IsValidTaprootScriptPath () {
					i.spendType = SPEND_TYPE_P2TR_Script
					i.segwit.SetTapScript (i.segwit.parseTapScript ())
				}

		}
	}

	// set the segwit field types
	for f, field := range i.segwit.fields {
		if len (field.AsType ()) == 0 {
			i.segwit.fields [f].SetType (GetStackItemType (field.AsBytes (), i.spendType == SPEND_TYPE_P2TR_Key || i.spendType == SPEND_TYPE_P2TR_Script))
		}
	}
}

func (i *Input) SetRedeemScript (redeemScript Script) {
	i.redeemScript = redeemScript
}

func (i *Input) GetInputScript () Script {
	return i.inputScript
}

func (i *Input) HasRedeemScript () bool {
	return !i.redeemScript.IsNil ()
}

func (i *Input) GetRedeemScript () Script {
	return i.redeemScript
}

func (i *Input) HasSegwitFields () bool {
	return !i.segwit.IsNil () && !i.segwit.IsEmpty ()
}

func (i *Input) GetSegwit () Segwit {
	return i.segwit
}

func (i *Input) IsCoinbase () bool {
	return i.coinbase
}

func (i *Input) GetPreviousOutputTxId () string {
	return i.previousOutputTxId
}

func (i *Input) GetPreviousOutputIndex () uint16 {
	return i.previousOutputIndex
}

func (i *Input) GetPreviousOutput () Output {
	return i.previousOutput
}

func (i *Input) GetSpendType () string {
	return i.spendType
}

func (i *Input) GetSequence () uint32 {
	return i.sequence
}

