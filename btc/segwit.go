package btc

import (
	"fmt"
	"encoding/hex"
)

type Segwit struct {
	fields [] string
	serializedScript Script
}

func (s *Segwit) ParseSerializedScript () bool {
	if s.IsValidP2wpkh () || s.IsValidTaprootKeyPath () { return false }

	// get the index of the serialized script
	scriptIndex := len (s.fields) - 1
	isTaprootScriptPath, tapScriptIndex := s.IsValidTaprootScriptPath ()
	if isTaprootScriptPath { scriptIndex = tapScriptIndex }

	// parse it
	serializedScriptBytes, err := hex.DecodeString (s.fields [scriptIndex])
	if err != nil { fmt.Println (err.Error ()); return false }
	serializedScript := NewScript (serializedScriptBytes)
	if serializedScript.HasParseError () { return false }

	s.serializedScript = serializedScript
	return true
}

func (s *Segwit) GetSerializedScript () Script {
	return s.serializedScript
}

func (s *Segwit) GetFields () [] string {
	return s.fields
}

func (s *Segwit) IsEmpty () bool {
	return len (s.fields) == 0
}

func (s *Segwit) IsValidP2wpkh () bool {
	if len (s.fields) != 2 { return false }

	publicKeyBytes, err := hex.DecodeString (s.fields [0])
	if err != nil { fmt.Println (err.Error ()); return false }

	signatureBytes, err := hex.DecodeString (s.fields [1])
	if err != nil { fmt.Println (err.Error ()); return false }

	return IsValidPublicKey (publicKeyBytes) && IsValidECSignature (signatureBytes)
}

/*
func (s *Segwit) IsValidP2wsh () bool {
}
*/

func (s *Segwit) IsValidTaprootKeyPath () bool {
	exactFieldCount := 1
	if s.HasAnnex () { exactFieldCount++ }

	if len (s.fields) != exactFieldCount { return false }

	signatureBytes, err := hex.DecodeString (s.fields [0])
	if err != nil { fmt.Println (err.Error ()); return false }

	return IsValidSchnorrSignature (signatureBytes)
}

// second return value is index of tap script or -1
func (s *Segwit) IsValidTaprootScriptPath () (bool, int) {
	minimumFieldCount := 2
	actualFieldCount := len (s.fields)
	controlBlockIndex := actualFieldCount - 1

	if s.HasAnnex () {
		minimumFieldCount++
		controlBlockIndex--
	}

	// there must be the minimum number of fields
	if actualFieldCount < minimumFieldCount { return false, -1 }

	// the control block must be of a valid length
	// we subtract 2 and mod with 64 because these are hex text fields with 2 characters representing each byte
	controlBlockLengthValid := (len (s.fields [controlBlockIndex]) - 2) % 64 == 0
	if !controlBlockLengthValid { return false, -1 }

	// the tap script must be parsable
	tapScriptIndex := controlBlockIndex - 1

	tapScriptBytes, err := hex.DecodeString (s.fields [tapScriptIndex])
	if err != nil { fmt.Println (err.Error ()); return false, -1 }

	tapScript := NewScript (tapScriptBytes)
	if tapScript.HasParseError () { return false, -1 }

	return true, tapScriptIndex
}

func (s *Segwit) HasAnnex () bool {
	fieldCount := len (s.fields)
	return fieldCount > 1 && s.fields [fieldCount - 1] [0:2] == "50";
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
