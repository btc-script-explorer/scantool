package btc

import (
	"fmt"
	"encoding/hex"
	"strconv"

	"btctx/app"
)

type Segwit struct {
	hexFields [] string
	witnessScript Script
	tapScript Script
	tapScriptIndex int64
}

func NewSegwit (hexFields [] string) Segwit {
	segwit := Segwit { hexFields: hexFields, tapScriptIndex: -1 }
	segwit.parseWitnessScript ()
	segwit.parseTapscript ()
	return segwit
}

func (s *Segwit) GetWitnessScript () Script {
	return s.witnessScript
}

func (s *Segwit) GetTapScript () (Script, int64) {
	return s.tapScript, s.tapScriptIndex
}

func (s *Segwit) GetFields () [] string {
	return s.hexFields
}

func (s *Segwit) IsNil () bool {
	return s.hexFields == nil
}

func (s *Segwit) IsEmpty () bool {
	return len (s.hexFields) == 0
}

type SegwitFieldHtmlData struct {
	DisplayText string
	ShowCopyButton bool
	CopyImageUrl string
	CopyText string
}

type SegwitHtmlData struct {
	HtmlId string
	Fields [] SegwitFieldHtmlData
	WitnessScript ScriptHtmlData
	TapScript ScriptHtmlData
	IsNil bool
}

func (s *Segwit) GetHtmlData (inputIndex uint32, maxWidthCh uint16) SegwitHtmlData {

	if s.IsNil () {
		return SegwitHtmlData { IsNil: true}
	}

	htmlId := "segwit-" + strconv.FormatUint (uint64 (inputIndex), 10)

	fields := [] SegwitFieldHtmlData (nil)
	if !s.IsEmpty () {
		settings := app.GetSettings ()
		copyImageUrl := "http://" + settings.Website.GetFullUrl () + "/image/clipboard-copy.png"

		fields = make ([] SegwitFieldHtmlData, len (s.hexFields))
		for f, field := range s.hexFields {
			if uint16 (len (field)) > maxWidthCh {
				fields [f] = SegwitFieldHtmlData { DisplayText: field [0 : maxWidthCh - 2] + "...", ShowCopyButton: true, CopyImageUrl: copyImageUrl, CopyText: field }
			} else {
				fields [f] = SegwitFieldHtmlData { DisplayText: field, ShowCopyButton: false }
			}
		}
	} else {
		fields = make ([] SegwitFieldHtmlData, 1)
		fields [0] = SegwitFieldHtmlData { DisplayText: "Empty", ShowCopyButton: false }
	}

	htmlData := SegwitHtmlData { HtmlId: htmlId, Fields: fields, IsNil: false }

	if !s.witnessScript.IsNil () {
		htmlData.WitnessScript = s.witnessScript.GetHtmlData ("Witness Script", htmlId + "-witness-script", maxWidthCh - 6, "hex")
	}

	if !s.tapScript.IsNil () {
		htmlData.TapScript = s.tapScript.GetHtmlData ("Tap Script", htmlId + "-tap-script", maxWidthCh - 6, "hex")
	}

//fmt.Println (htmlData)
	return htmlData
}

func (s *Segwit) IsValidP2wpkh () bool {
	if len (s.hexFields) != 2 { return false }

	signatureBytes, err := hex.DecodeString (s.hexFields [0])
	if err != nil { fmt.Println (err.Error ()); return false }

	publicKeyBytes, err := hex.DecodeString (s.hexFields [1])
	if err != nil { fmt.Println (err.Error ()); return false }

	return IsValidECPublicKey (publicKeyBytes) && IsValidECSignature (signatureBytes)
}

/*
func (s *Segwit) IsValidP2wsh () bool {
}
*/

func (s *Segwit) IsValidTaprootKeyPath () bool {
	exactFieldCount := 1
	if s.HasAnnex () { exactFieldCount++ }

	if len (s.hexFields) != exactFieldCount { return false }

	signatureBytes, err := hex.DecodeString (s.hexFields [0])
	if err != nil { fmt.Println (err.Error ()); return false }

	return IsValidSchnorrSignature (signatureBytes)
}

func (s *Segwit) parseWitnessScript () {

	// if there are no segwit fields, then there is no witness script
	fieldCount := len (s.hexFields)
	if fieldCount < 1 { return }

	// read the witness script
	witnessScriptIndex := fieldCount - 1
	witnessScriptBytes, err := hex.DecodeString (s.hexFields [witnessScriptIndex])
	if err != nil { fmt.Println (err.Error ()); return }

	// the script must be parsable
	witnessScript := NewScript (witnessScriptBytes)
	if witnessScript.HasParseError () { return }

	s.witnessScript = witnessScript
}

func (s *Segwit) parseTapscript () {

	// gather some data to determine whether this could be a Taproot segwit block
	minimumFieldCount := 2
	actualFieldCount := len (s.hexFields)
	controlBlockIndex := actualFieldCount - 1

	if s.HasAnnex () {
		minimumFieldCount++
		controlBlockIndex--
	}

	// if there is not the required minimum number of segwit fields, then there is no tap script
	if actualFieldCount < minimumFieldCount { return }

	// the control block must be of a valid length
	// we use (len - 2) % 64 because these are hex strings with 2 characters representing each byte in the segwit field
	controlBlockLengthValid := (len (s.hexFields [controlBlockIndex]) - 2) % 64 == 0
	if !controlBlockLengthValid { return }

	// now we read the tap script
	tapScriptIndex := int64 (controlBlockIndex) - 1
	tapScriptBytes, err := hex.DecodeString (s.hexFields [tapScriptIndex])
	if err != nil { fmt.Println (err.Error ()); return }

	// the script must be parsable
	tapScript := NewScript (tapScriptBytes)
	if tapScript.HasParseError () { return }

	s.tapScript = tapScript
	s.tapScriptIndex = tapScriptIndex
}

func (s *Segwit) HasAnnex () bool {
	fieldCount := len (s.hexFields)
	return fieldCount > 1 && s.hexFields [fieldCount - 1] [0:2] == "50";
}

/*
// This function can be used to read a raw transaction as a byte array.
// This method has been abandoned because it does not include bitcoin addresses.
// However, it is still included here, commented out, in case it becomes more
// convenient to read transactions this way if/when other bitcoin node types are supported.
func NewSegwit (raw_bytes [] byte) (Segwit, int) {

	value_reader := ValueReader {}

	pos := 0

	field_count, byte_count := value_reader.ReadVarInt (raw_bytes [pos:])
	pos += byte_count

//fmt.Println ("Segwit fields = ", field_count)

	raw_fields := make ([] [] byte, field_count)
	if field_count > 0 {
		for s := 0; s < int (field_count); s++ {
			field_len, byte_count := value_reader.ReadVarInt (raw_bytes [pos:])
			pos += int (byte_count)

			raw_fields [s] = make ([] byte, field_len)
			copy (raw_fields [s], raw_bytes [pos : pos + int (field_len)])
//			fields [s] = hex.EncodeToString (raw_bytes [pos : pos + int (field_len)])
//fmt.Println (fields [s])
			pos += int (field_len)
//fmt.Println (s + 1, ": Field len = ", field_len, ", pos = ", pos)
		}
	}

	return Segwit { raw_fields: raw_fields, has_witness_script: false, has_tap_script: false }, pos
}

func (s *Segwit) ParseSerializedScript () bool {
	if s.IsValidP2wpkh () || s.IsValidTaprootKeyPath () { return false }

	// get the index of the serialized script
	script_index := len (s.raw_fields) - 1
	is_taproot_script_path, tap_script_index := s.IsValidTaprootScriptPath ()
	if is_taproot_script_path { script_index = tap_script_index }

	// parse it
	serialized_script, _ := NewScript (s.raw_fields [script_index])
	if serialized_script.parse_error { return false }

	s.has_witness_script = !is_taproot_script_path
	s.has_tap_script = is_taproot_script_path
	s.serialized_script = serialized_script
	return true
}
*/
