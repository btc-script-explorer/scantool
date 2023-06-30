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

	valueReader := ValueReader {}

	publicKeyBytes, err := hex.DecodeString (s.fields [0])
	if err != nil { fmt.Println (err.Error ()); return false }

	signatureBytes, err := hex.DecodeString (s.fields [1])
	if err != nil { fmt.Println (err.Error ()); return false }

	return valueReader.IsValidPublicKey (publicKeyBytes) && valueReader.IsValidECSignature (signatureBytes)
}

/*
func (s *Segwit) IsValidP2wsh () bool {
}
*/

func (s *Segwit) IsValidTaprootKeyPath () bool {
	exactFieldCount := 1
	if s.HasAnnex () { exactFieldCount++ }

	if len (s.fields) != exactFieldCount { return false }

	valueReader := ValueReader {}
	signatureBytes, err := hex.DecodeString (s.fields [0])
	if err != nil { fmt.Println (err.Error ()); return false }

	return valueReader.IsValidSchnorrSignature (signatureBytes)
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

