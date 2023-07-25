package btc

import (
	"fmt"
	"encoding/hex"

	"btctx/app"
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

func (s *Segwit) GetFields () [] SegwitField {
	return s.fields
}

func (s *Segwit) IsNil () bool {
	return s.fields == nil
}

func (s *Segwit) IsEmpty () bool {
	return s.IsNil () || len (s.fields) == 0
}

type SegwitHtmlData struct {
	FieldSet FieldSetHtmlData
	WitnessScript ScriptHtmlData
	TapScript ScriptHtmlData
	IsEmpty bool
}

func (s *Segwit) GetHtmlData (inputIndex uint32, displayTypeClassPrefix string, usingSchnorrSignatures bool) SegwitHtmlData {

	if s.IsNil () {
		return SegwitHtmlData { IsEmpty: true}
	}

	htmlId := fmt.Sprintf ("input-%d-segwit", inputIndex)
	const maxCharWidth = uint (89)

	var hexFieldsHtml [] FieldHtmlData
	var textFieldsHtml [] FieldHtmlData
	var typeFieldsHtml [] FieldHtmlData

	if !s.IsEmpty () {

		fieldCount := len (s.fields);

		// segwit does not know what some of its types are, so it identifies data types when the HTML is rendered
		if !s.witnessScript.IsNil () { s.fields [fieldCount - 1].dataType = "<<< SERIALIZED WITNESS SCRIPT >>>" }
		if !s.tapScript.IsNil () {

			cbIndex := s.getControlBlockIndex ()
			cbLeafCount := 0
			if cbIndex != -1 {
				cbLeafCount = (len (s.fields [cbIndex].rawBytes) - 1) / 32
			} else {
				fmt.Println ("Segwit has tap script but no control block.")
			}

			// set the field types for the Taproot Segwit fields
			if s.HasAnnex () {
				annexIndex := len (s.fields) - 1
				s.fields [annexIndex].dataType = fmt.Sprintf ("Annex (%d Bytes)", len (s.fields [annexIndex].rawBytes))
			}

			s.fields [s.tapScriptIndex].dataType = "<<< SERIALIZED TAP SCRIPT >>>"

			leafCountLabel := "TapLea"
			if cbLeafCount == 1 { leafCountLabel += "f" } else { leafCountLabel += "ves" }
			s.fields [cbIndex].dataType = fmt.Sprintf ("Control Block (%d %s)", cbLeafCount, leafCountLabel)

			// set the field types for the Tap Script
			for i, field := range s.tapScript.fields {
				if !field.IsOpcode () {
					s.tapScript.fields [i].dataType = GetStackItemType (field.AsBytes (), true)
				}
			}
		}

		hexFieldsHtml = make ([] FieldHtmlData, fieldCount)
		textFieldsHtml = make ([] FieldHtmlData, fieldCount)
		typeFieldsHtml = make ([] FieldHtmlData, fieldCount)

		for f, field := range s.fields {

			// set any field types that aren't already set
			if len (s.fields [f].dataType) == 0 {
				s.fields [f].dataType = GetStackItemType (field.AsBytes (), usingSchnorrSignatures)
			}

			// hex strings
			entireHexField := field.AsHex (0)
			hexFieldsHtml [f] = FieldHtmlData { DisplayText: field.AsHex (maxCharWidth), ShowCopyButton: false }
			if hexFieldsHtml [f].DisplayText != entireHexField {
				hexFieldsHtml [f].ShowCopyButton = true
				hexFieldsHtml [f].CopyText = entireHexField
			}

			// text strings
			entireTextField := field.AsText (0)
			textFieldsHtml [f] = FieldHtmlData { DisplayText: field.AsText (maxCharWidth), ShowCopyButton: false }
			if textFieldsHtml [f].DisplayText != entireTextField {
				textFieldsHtml [f].ShowCopyButton = true
				textFieldsHtml [f].CopyText = entireTextField
			}

			// field types
			typeFieldsHtml [f] = FieldHtmlData { DisplayText: s.fields [f].dataType, ShowCopyButton: false }
		}
	}

	settings := app.GetSettings ()
	copyImageUrl := "http://" + settings.Website.GetFullUrl () + "/image/clipboard-copy.png"

	fieldSet := FieldSetHtmlData { HtmlId: htmlId, DisplayTypeClassPrefix: displayTypeClassPrefix, CharWidth: maxCharWidth, HexFields: hexFieldsHtml, TextFields: textFieldsHtml, TypeFields: typeFieldsHtml, CopyImageUrl: copyImageUrl }
	htmlData := SegwitHtmlData { FieldSet: fieldSet, IsEmpty: s.IsEmpty () }

	htmlData.WitnessScript = s.witnessScript.GetHtmlData (htmlId + "-witness-script", displayTypeClassPrefix)
	htmlData.TapScript = s.tapScript.GetHtmlData (htmlId + "-tap-script", displayTypeClassPrefix)

	return htmlData
}

func (s *Segwit) IsValidP2wpkh () bool {
	if len (s.fields) != 2 { return false }

	signatureBytes := s.fields [0].AsBytes ()
	publicKeyBytes := s.fields [1].AsBytes ()
	return IsValidECPublicKey (publicKeyBytes) && IsValidECSignature (signatureBytes)
}

func (s *Segwit) IsValidTaprootKeyPath () bool {
	exactFieldCount := 1
	if s.HasAnnex () { exactFieldCount++ }

	if len (s.fields) != exactFieldCount { return false }

	signatureBytes := s.fields [0].AsBytes ()
	return IsValidSchnorrSignature (signatureBytes)
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

	controlBlockIndex := s.getControlBlockIndex ()
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

func (s *Segwit) getControlBlockIndex () int {

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

