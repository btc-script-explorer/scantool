package btc

import (
	"fmt"
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
	previousOutputTxId string
	previousOutputIndex uint32
	coinbase bool
	spendType string
	inputScript Script
	redeemScript Script
	segwit Segwit
	sequence uint32

	multisigRedeemScript bool
	multisigWitnessScript bool
	ordinalTapScript bool
}

func NewInput (coinbase bool, previousOutputTxId string, previousOutputIndex uint32, inputScript Script, segwit Segwit, sequence uint32) Input {

	i := Input { coinbase: coinbase, previousOutputTxId: previousOutputTxId, previousOutputIndex: previousOutputIndex, inputScript: inputScript, segwit: segwit, sequence: sequence }

	if i.coinbase {
		i.spendType = "COINBASE"
	} else {
		nodeClient := GetNodeClient ()
		i.spendType = ""

		inputScriptHasFields := !inputScript.IsEmpty ()

		isP2shWrappedType := !i.segwit.IsEmpty () && inputScriptHasFields
		isWitnessType := !i.segwit.IsNil () && !inputScriptHasFields

		// a form of duck typing is used here in order to identify the spend type with no knowledge of the previous output type
		// messages are printed to the screen when there are potential misidentifications

		// it is impossible to know for sure that a redemption is not non-standard without seeing the previous output script
		// spend types for legacy output types use the same name as the output type, so there is no reason to identify them at this point
		// therefore, we defer identification of these until the previous output is retrieved by the client

		if isP2shWrappedType {

			// the only two spend types that have segwits fields along with a non-empty input script are the p2sh-wrapped spend types
			// these are the easiest to identify by their inputs
			i.redeemScript = i.inputScript.GetSerializedScript ()
			if !i.redeemScript.IsNil () {
				if i.redeemScript.IsP2shP2wpkhRedeemScript () {
					i.spendType = SPEND_TYPE_P2SH_P2WPKH
				} else if i.redeemScript.IsP2shP2wshRedeemScript () {
					i.spendType = SPEND_TYPE_P2SH_P2WSH
					i.segwit.SetWitnessScript (i.segwit.parseWitnessScript ())
				} else { fmt.Println ("Segwit and Input Script exist, but redeem script is not a p2sh-wrapped script.") }
			} else { fmt.Println ("Segwit and Input Script exist, but no redeem script.") }

			i.inputScript.SetFieldType (i.inputScript.GetParsedFieldCount () - 1, "SERIALIZED REDEEM SCRIPT")

			// input script
			if i.inputScript.IsNil () { fmt.Println (i.spendType + " input with no input script.") }
			parsedFieldCount := i.inputScript.GetParsedFieldCount ()
			if parsedFieldCount != 1 { fmt.Println (i.spendType + " input script has wrong field count. Found ", parsedFieldCount, ", expected 1.") }

			// redeem script should always exist for these types
			if i.redeemScript.IsNil () { fmt.Println (i.spendType + " input with no redeem script.") }
			parsedFieldCount = i.redeemScript.GetParsedFieldCount ()
			if parsedFieldCount != 2 { fmt.Println (i.spendType + " redeem script has wrong field count. Found ", parsedFieldCount, ", expected 2.") }

			i.redeemScript.SetFieldType (0, "OP_0")
			if i.spendType == SPEND_TYPE_P2SH_P2WPKH { i.redeemScript.SetFieldType (1, "Witness Program (Public Key Hash)") } else 
			if i.spendType == SPEND_TYPE_P2SH_P2WSH { i.redeemScript.SetFieldType (1, "Witness Program (Script Hash)") }
		}

		if isWitnessType {

			// it looks like one of the witness types
			possibleWitnessScript := i.segwit.parseWitnessScript ()
			possibleTapScript, possibleTapScriptIndex := i.segwit.parseTapScript ()

			possibleSpendTypeCount := 0

st := ""
			possibleP2wpkh := i.segwit.IsValidP2wpkh ()
			if possibleP2wpkh { st += OUTPUT_TYPE_P2WPKH + ", "; possibleSpendTypeCount++ }

			possibleTaprootKeyPath := i.segwit.IsValidTaprootKeyPath ()
			if possibleTaprootKeyPath { st += SPEND_TYPE_P2TR_Key + ", "; possibleSpendTypeCount++ }

			possibleP2wsh := !possibleWitnessScript.IsNil () && !possibleWitnessScript.HasParseError () && possibleWitnessScript.AppearsValid ()
			if possibleP2wsh { st += OUTPUT_TYPE_P2WSH + ", "; possibleSpendTypeCount++ }

			possibleTaprootScriptPath := !possibleTapScript.IsNil () && !possibleTapScript.HasParseError () && possibleTapScript.AppearsValid ()
			if possibleTaprootScriptPath { st += SPEND_TYPE_P2TR_Script + ", "; possibleSpendTypeCount++ }

			// set the spend type
			if possibleSpendTypeCount > 1 {

				// duck typing of the input data has resulted in an ambiguous identification of the spend type
				// get the previous output for exact identification

				i.spendType = SPEND_TYPE_NonStandard

				previousOutput := nodeClient.GetPreviousOutput (i.GetPreviousOutputTxId (), i.GetPreviousOutputIndex ())
				correctOutputType := previousOutput.GetOutputType ()

				switch correctOutputType {
					case OUTPUT_TYPE_TAPROOT:
						if possibleTaprootKeyPath {
							i.spendType = SPEND_TYPE_P2TR_Key
						} else if possibleTaprootScriptPath {
							i.spendType = SPEND_TYPE_P2TR_Script
							i.segwit.SetTapScript (possibleTapScript, possibleTapScriptIndex)
						}
						break
					case OUTPUT_TYPE_P2WPKH:
						i.spendType = correctOutputType
						break
					case OUTPUT_TYPE_P2WSH:
						i.spendType = correctOutputType
						i.segwit.SetWitnessScript (possibleWitnessScript)
						break
					default:
						// the output type did not turn out to be a witness type at all, it must be one of the legacy types
						i.spendType = ""
						fmt.Println ("previous output type \"" + correctOutputType + "\" incorrectly paired with witness spend type.")
						break
				}
if i.spendType == SPEND_TYPE_P2TR_Script { fmt.Println (previousOutputTxId, previousOutputIndex, st, i.spendType) }
			} else {

				// there was only one possible spend type, no need to check the previous output
				if possibleP2wpkh {
					i.spendType = OUTPUT_TYPE_P2WPKH
				} else if possibleTaprootKeyPath {
					i.spendType = SPEND_TYPE_P2TR_Key
				} else if possibleP2wsh {
					i.spendType = OUTPUT_TYPE_P2WSH
					if possibleWitnessScript.IsEmpty () { fmt.Printf ("Input that redeems %s:%d has %s spend type with empty witness script.\n", previousOutputTxId, previousOutputIndex, i.spendType) }
					i.segwit.SetWitnessScript (possibleWitnessScript)
				} else if possibleTaprootScriptPath {
					i.spendType = SPEND_TYPE_P2TR_Script
					if possibleTapScript.IsEmpty () { fmt.Printf ("Input that redeems %s:%d has %s spend type with empty tap script.\n", previousOutputTxId, previousOutputIndex, i.spendType) }
					i.segwit.SetTapScript (possibleTapScript, possibleTapScriptIndex)
				}
			}
		}

		// set any segwit field types that aren't already set
		if !i.segwit.IsEmpty () {
			for f, field := range i.segwit.fields {
				if len (field.AsType ()) == 0 {
					i.segwit.fields [f].SetType (GetStackItemType (field.AsBytes (), i.spendType == SPEND_TYPE_P2TR_Key || i.spendType == SPEND_TYPE_P2TR_Script))
				}
			}
		}

		// if the spend type is empty, it must redeem a legacy output type
		// therefore, we must check for the possibility of a P2SH output
		if len (i.spendType) == 0 {
			redeemScript := i.inputScript.GetSerializedScript ()
			if !redeemScript.HasParseError () {

				previousOutput := nodeClient.GetPreviousOutput (i.GetPreviousOutputTxId (), i.GetPreviousOutputIndex ())
				correctOutputType := previousOutput.GetOutputType ()
				if correctOutputType == OUTPUT_TYPE_P2SH {
					i.spendType = OUTPUT_TYPE_P2SH
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
			}
		}

		i.multisigRedeemScript = i.spendType == OUTPUT_TYPE_P2SH && i.redeemScript.IsMultiSigOutput ()
		i.multisigWitnessScript = (i.spendType == OUTPUT_TYPE_P2WSH || i.spendType == SPEND_TYPE_P2SH_P2WSH) && i.segwit.witnessScript.IsMultiSigOutput ()
		i.ordinalTapScript = i.spendType == SPEND_TYPE_P2TR_Script && i.segwit.tapScript.IsOrdinal ()
	}

	return i
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

func (i *Input) HasMultisigRedeemScript () bool {
	return i.multisigRedeemScript
}

func (i *Input) HasMultisigWitnessScript () bool {
	return i.multisigWitnessScript
}

func (i *Input) HasOrdinalTapScript () bool {
	return i.ordinalTapScript
}

func (i *Input) IsCoinbase () bool {
	return i.coinbase
}

func (i *Input) GetPreviousOutputTxId () string {
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

