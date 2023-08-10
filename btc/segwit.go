package btc

import (
	"fmt"
	"encoding/hex"
)

type SegwitField struct {
	rawBytes [] byte
	dataType string
}

func (swf *SegwitField) AsBytes () [] byte {
	return swf.rawBytes
}

// if maxLength is 0, it will be ignored
func (swf *SegwitField) AsHex () string {
	return hex.EncodeToString (swf.rawBytes)
}

func (swf *SegwitField) SetType (dataType string) {
	swf.dataType = dataType
}

func (swf *SegwitField) AsType () string {
	return swf.dataType
}

func (swf *SegwitField) AsText () string {
	return string (swf.rawBytes)
}


type Segwit struct {
	fields [] SegwitField
	witnessScript Script
	tapScript Script
	tapScriptIndex uint32
}

func NewSegwit (rawFields [] [] byte) Segwit {

	fields := make ([] SegwitField, len (rawFields))
	for f, field := range rawFields {
		fields [f] = SegwitField { rawBytes: field }
	}

	// segwit is not aware of the types of all of its fields
	for f, field := range fields {
		if len (field.rawBytes) == 0 {
			fields [f].dataType = "<<< ZERO-LENGTH FIELD >>>"
		}
	}

	return Segwit { fields: fields, tapScriptIndex: INVALID_CB_INDEX }
}

func (s *Segwit) GetWitnessScript () Script {
	return s.witnessScript
}

func (s *Segwit) GetTapScript () (Script, uint32) {
	return s.tapScript, s.tapScriptIndex
}

func (s *Segwit) GetFieldCount () uint16 {
	return uint16 (len (s.fields))
}

func (s *Segwit) GetFields () [] SegwitField {
	return s.fields
}

func (s *Segwit) IsNil () bool {
	return s.fields == nil
}

func (s *Segwit) IsEmpty () bool {
	return s.IsNil () || len (s.fields) == 0
}

func (s *Segwit) IsValidP2wpkh () bool {
	if len (s.fields) < 2 { return false }

	// we must count only non-empty fields
	nonEmptyFieldCount := 0
	for f := 0; f < len (s.fields); f++ {
		fieldBytes := s.fields [f].AsBytes ()
		if len (fieldBytes) > 0 {
			if nonEmptyFieldCount == 0 {
				// the first non-empty field must be a Signature
				if !IsValidECSignature (fieldBytes) {
					return false
				}
			} else if nonEmptyFieldCount == 1 {
				// the first non-empty field must be a public key
				if !IsValidECPublicKey (fieldBytes) {
					return false
				}
			}

			nonEmptyFieldCount++
		}
	}
	if nonEmptyFieldCount != 2 { return false }

	return true
}

func (s *Segwit) IsValidTaprootKeyPath () bool {
	exactFieldCount := 1
	if s.HasAnnex () { exactFieldCount++ }

	// we must count only non-empty fields
	nonEmptyFieldCount := 0
	for f := 0; f < len (s.fields); f++ {
		fieldBytes := s.fields [f].AsBytes ()
		if len (fieldBytes) > 0 {
			if nonEmptyFieldCount == 0 {
				// the first non-empty field must be a Schnorr Signature
				if !IsValidSchnorrSignature (fieldBytes) {
					return false
				}
			}

			nonEmptyFieldCount++
		}
	}
	if nonEmptyFieldCount != exactFieldCount { return false }

	return true
}

/*
func (s *Segwit) IsValidP2wsh (setWitnessScript bool) bool {
	witnessScript := s.parseWitnessScript ()

	if setWitnessScript {
		s.witnessScript = witnessScript
	}

	return !witnessScript.IsNil ()
}
*/

func (s *Segwit) parseWitnessScript () Script {

	// if there are no segwit fields, then there is no witness script
	fieldCount := len (s.fields)
	if fieldCount < 1 { return Script {} }

	// read the witness script
	witnessScriptIndex := fieldCount - 1
	witnessScriptBytes := s.fields [witnessScriptIndex].AsBytes ()

	// the script must be parsable
	witnessScript := NewScript (witnessScriptBytes)
	if witnessScript.HasParseError () { return Script {} }

	return witnessScript
}

func (s *Segwit) SetWitnessScript (ws Script) {
	s.witnessScript = ws
	s.fields [len (s.fields) - 1].SetType ("<<< SERIALIZED WITNESS SCRIPT >>>")
}

/*
func (s *Segwit) IsValidTaprootScriptPath (setTapScript bool) bool {

	tapScript, tapScriptIndex := s.parseTapScript ()

	if setTapScript {
		s.tapScript = tapScript
		s.tapScriptIndex = tapScriptIndex
	}

	return tapScriptIndex != -1
}
*/

func (s *Segwit) parseTapScript () (Script, uint32) {

	controlBlockIndex := s.GetControlBlockIndex ()
	if controlBlockIndex == INVALID_CB_INDEX { return Script {}, INVALID_CB_INDEX }

	// now we read the tap script
	tapScriptIndex := uint32 (controlBlockIndex) - 1
	tapScriptBytes := s.fields [tapScriptIndex].AsBytes ()

	// the script must be parsable
	tapScript := NewScript (tapScriptBytes)
	if tapScript.IsNil () || tapScript.HasParseError () { return Script {}, INVALID_CB_INDEX }

	return tapScript, tapScriptIndex
}

func (s *Segwit) SetTapScript (ts Script, i uint32) {
	s.tapScript = ts
	s.tapScriptIndex = i

	cbIndex := s.GetControlBlockIndex ()
	cbLeafCount := 0
	if cbIndex != INVALID_CB_INDEX {
		cbLeafCount = (len (s.fields [cbIndex].AsBytes ()) - 1) / 32
	} else {
		fmt.Println ("Segwit has tap script but no control block.")
	}

	// set the field types for the Taproot Segwit fields
	if s.HasAnnex () {
		annexIndex := len (s.fields) - 1
		s.fields [annexIndex].SetType (fmt.Sprintf ("Annex (%d Bytes)", len (s.fields [annexIndex].AsBytes ())))
	}

	s.fields [s.tapScriptIndex].SetType ("<<< SERIALIZED TAP SCRIPT >>>")

	leafCountLabel := "TapLea"
	if cbLeafCount == 1 { leafCountLabel += "f" } else { leafCountLabel += "ves" }
	s.fields [cbIndex].SetType (fmt.Sprintf ("Control Block (%d %s)", cbLeafCount, leafCountLabel))

	// set the field types for the Tap Script
	tapScriptFields := s.tapScript.GetFields ()
	for f, field := range tapScriptFields {
		if !field.IsOpcode () {
			tapScriptFields [f].SetType (GetStackItemType (field.AsBytes (), true))
		}
	}

}

func (s *Segwit) HasAnnex () bool {
	fieldCount := len (s.fields)
	return fieldCount > 1 && s.fields [fieldCount - 1].AsBytes () [0] == 0x50;
}

const INVALID_CB_INDEX = 0xffffffff
func (s *Segwit) GetControlBlockIndex () uint32 {

	minimumFieldCount := 2
	actualFieldCount := len (s.fields)
	controlBlockIndex := actualFieldCount - 1

	if s.HasAnnex () {
		minimumFieldCount++
		controlBlockIndex--
	}

	// if this is really a control block, there will be a minimum number of segwit fields
	if actualFieldCount < minimumFieldCount { return INVALID_CB_INDEX }

	// a valid control must have a valid length
	controlBlockLength := len (s.fields [controlBlockIndex].AsBytes ())
	validControlBlockLength := controlBlockLength >= 33 && (controlBlockLength - 1) % 32 == 0
	if !validControlBlockLength { return INVALID_CB_INDEX }

	return uint32 (controlBlockIndex)
}

