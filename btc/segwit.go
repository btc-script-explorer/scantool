package btc

import (
//	"fmt"
	"encoding/hex"
)

func ShortenField (fieldText string, length uint, dotCount uint) string {

	fieldLength := uint (len (fieldText))
	if fieldLength <= length { return fieldText }

	if dotCount >= length { return fieldText }

	charsToInclude := length - dotCount
	if charsToInclude < 2 { return fieldText }

	charsInEachPart := charsToInclude / 2
	part1End := charsInEachPart + (charsToInclude % 2)
	part2Begin := fieldLength - charsInEachPart

	// build the final result
	shortenedField := fieldText [0 : part1End]
	for i := uint (0); i < dotCount; i++ { shortenedField += "." }
	shortenedField += fieldText [part2Begin :]

	return shortenedField
}

type SegwitField struct {
	rawBytes [] byte
	dataType string
}

func (swf *SegwitField) AsBytes () [] byte {
	return swf.rawBytes
}

// if maxLength is 0, it will be ignored
func (swf *SegwitField) AsHex (maxLength uint) string {
	hexField := hex.EncodeToString (swf.rawBytes)
	if maxLength == 0 || uint (len (hexField)) <= maxLength {
		return hexField
	}

	return ShortenField (hexField, maxLength, 5)
}

func (swf *SegwitField) SetType (dataType string) {
	swf.dataType = dataType
}

func (swf *SegwitField) AsType () string {
	return swf.dataType
}

func (swf *SegwitField) AsText (maxLength uint) string {
	textField := string (swf.rawBytes)
	if maxLength == 0 || uint (len (textField)) <= maxLength {
		return textField
	}

	return ShortenField (textField, maxLength, 5)
}


type Segwit struct {
	fields [] SegwitField
	witnessScript Script
	tapScript Script
	tapScriptIndex int64
}

func NewSegwit (rawFields [] [] byte) Segwit {

	fields := make ([] SegwitField, len (rawFields))
	for f, field := range rawFields {
		fields [f] = SegwitField { rawBytes: field }
	}

	// segwit is not aware of the types of all of its fields, so identification will be deferred until the HTML is rendered
	for f, field := range fields {
		if len (field.rawBytes) == 0 {
			fields [f].dataType = "<<< ZERO-LENGTH FIELD >>>"
		}
	}

	return Segwit { fields: fields, tapScriptIndex: -1 }
}

func (s *Segwit) GetWitnessScript () Script {
	return s.witnessScript
}

func (s *Segwit) GetTapScript () (Script, int64) {
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

func (s *Segwit) parseTapScript () (Script, int64) {

	controlBlockIndex := s.GetControlBlockIndex ()
	if controlBlockIndex == -1 { return Script {}, -1 }

	// now we read the tap script
	tapScriptIndex := int64 (controlBlockIndex) - 1
	tapScriptBytes := s.fields [tapScriptIndex].AsBytes ()

	// the script must be parsable
	tapScript := NewScript (tapScriptBytes)
	if tapScript.IsNil () || tapScript.HasParseError () { return Script {}, -1 }

	return tapScript, tapScriptIndex
}

func (s *Segwit) HasAnnex () bool {
	fieldCount := len (s.fields)
	return fieldCount > 1 && s.fields [fieldCount - 1].AsBytes () [0] == 0x50;
}

func (s *Segwit) GetControlBlockIndex () int {

	minimumFieldCount := 2
	actualFieldCount := len (s.fields)
	controlBlockIndex := actualFieldCount - 1

	if s.HasAnnex () {
		minimumFieldCount++
		controlBlockIndex--
	}

	// if this is really a control block, there will be a minimum number of segwit fields
	if actualFieldCount < minimumFieldCount { return -1 }

	// a valid control must have a valid length
	controlBlockLength := len (s.fields [controlBlockIndex].AsBytes ())
	validControlBlockLength := controlBlockLength >= 33 && (controlBlockLength - 1) % 32 == 0
	if !validControlBlockLength { return -1 }

	return controlBlockIndex
}

