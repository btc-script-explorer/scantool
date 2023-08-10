package btc

import (
	"fmt"
	"encoding/hex"
)

type ScriptField struct {
	rawBytes [] byte
	isOpcode bool
	dataType string
}

func (sf *ScriptField) SetIsOpcode (isOpcode bool) {
	sf.isOpcode = isOpcode
}

func (sf *ScriptField) IsOpcode () bool {
	return sf.isOpcode
}

func (sf *ScriptField) SetBytes (bytes [] byte) {
	sf.rawBytes = bytes
}

func (sf *ScriptField) AsBytes () [] byte {
	return sf.rawBytes
}

func (sf *ScriptField) AsHex () string {
	if sf.isOpcode {
		return getOpcodeName (sf.rawBytes [0])
	}

	return hex.EncodeToString (sf.rawBytes)
}

func (sf *ScriptField) SetType (dataType string) {
	sf.dataType = dataType
}

func (sf *ScriptField) AsType () string {
	if sf.isOpcode {
		return getOpcodeName (sf.rawBytes [0])
	}

	return sf.dataType
}

func (sf *ScriptField) AsText () string {
	if sf.isOpcode {
		return getOpcodeName (sf.rawBytes [0])
	}

	return string (sf.rawBytes)
}

type Script struct {
	rawBytes [] byte
	fields [] ScriptField
	parseError bool
	appearsValid bool
}

func NewScript (rawBytes [] byte) Script {

	if rawBytes == nil { return Script {} }

	scriptLen := len (rawBytes)

	parseError := false
	pos := 0
	bytesRemaining := len (rawBytes)
	fieldMap := make (map [int] ScriptField)

	// parse the script
	fieldCount := 0
	for bytesRemaining > 0 {

		fieldSizeLen := 0
		fieldLen := 0
		isOpcode := false

		nextByte := rawBytes [pos]
		if isValidOpcode (nextByte) {

			// it is an opcode
			isOpcode = true
			fieldLen = 1
		} else {

			// it is a stack item
			fieldSizeLen = 1
			if nextByte < 0x4c {
				// it is a field length
				fieldLen = int (nextByte)
			} else {
				// it is a push data opcode
				pushDataSize := 0
				switch nextByte {
					case 0x4c: pushDataSize = 1
					case 0x4d: pushDataSize = 2
					case 0x4e: pushDataSize = 4
					default:
						fieldSizeLen = 0
						fieldLen = bytesRemaining + 1
				}

				if pushDataSize > 0 {
					// read the size of the field
					fieldSizeBegin := pos + fieldSizeLen
					fieldSizeEnd := fieldSizeBegin + pushDataSize
					if fieldSizeEnd <= scriptLen {
						fieldLen = int (ReadNumeric (rawBytes [fieldSizeBegin : fieldSizeEnd]))
					} else {
						fieldLen = 0 // there are not enough bytes left in the script to read the field size
					}
					fieldSizeLen += pushDataSize
				}
			}
		}

		startPos := pos + fieldSizeLen
		totalLen := fieldSizeLen + fieldLen
		if bytesRemaining >= totalLen {
			fieldMap [fieldCount] = ScriptField { rawBytes: rawBytes [startPos : startPos + fieldLen], isOpcode: isOpcode }
			fieldCount++
			pos += totalLen
			bytesRemaining -= totalLen
		} else {
			// the script contains a parse error
			parseError = true
			if bytesRemaining > fieldSizeLen {
				// there are bytes beyond the field size, so we will take whatever is left
				fieldLen = bytesRemaining - fieldSizeLen
				fieldMap [fieldCount] = ScriptField { rawBytes: rawBytes [startPos : startPos + fieldLen], isOpcode: isOpcode }
				fieldCount++
			}
			pos = scriptLen
			bytesRemaining = 0
		}
	}

	// build the human readable script item list, checking for invalid scripts
	appearsValid := true
	fields := make ([] ScriptField, fieldCount)
	if fieldCount > 0 {
		ifCount := 0
		elseCount := 0
		endIfCount := 0
		opReturnCount := 0
		for f, field := range fieldMap {
			fieldText := field.AsHex ()
			if fieldText == "OP_IF" || fieldText == "OP_NOTIF" { ifCount++ }
			if fieldText == "OP_ELSE" { elseCount++ }
			if fieldText == "OP_ENDIF" { endIfCount++ }
			if fieldText == "OP_RETURN" { opReturnCount++ }
			fields [f] = field
		}
		looksLikeOpReturn := opReturnCount > 0 && ifCount == 0
		mismatchedElse := elseCount > 0 && ifCount == 0
		mismatchedIf := ifCount != endIfCount
		appearsValid = !(looksLikeOpReturn || mismatchedElse || mismatchedIf)
	}

	// finally, determine the data type of each script item
	for f, field := range fields {
		if field.IsOpcode () {
			fields [f].dataType = field.AsHex ()
		} else {
			fields [f].dataType = GetStackItemType (field.AsBytes (), false)
		}
	}

	return Script { rawBytes: rawBytes, fields: fields, parseError: parseError, appearsValid: appearsValid }
}

// used only for testing
func (s *Script) PrintToScreen () {
	fmt.Println ("\n**************************************")
	fmt.Println (len (s.fields), " fields in script")
	for _, f := range s.fields {
		fmt.Println (f.AsHex ())
	}
	fmt.Println ("**************************************\n")
}

func (s *Script) AsBytes () [] byte {
	return s.rawBytes
}

func (s *Script) GetParsedFieldCount () int {
	return len (s.fields)
}

func (s *Script) SetFieldType (fieldIndex int, fieldType string) {

	fieldCount := len (s.fields)
	if fieldCount <= fieldIndex { fmt.Println ("Setting type of field ", fieldIndex, " in script that only has ", fieldCount, " fields."); return }

	s.fields [fieldIndex].dataType = fieldType
}

func (s *Script) HasParseError () bool {
	return s.parseError
}

func (s *Script) AppearsValid () bool {
	return s.appearsValid
}

func (s *Script) GetFieldCount () uint16 {
	return uint16 (len (s.fields))
}

func (s *Script) GetFields () [] ScriptField {
	return s.fields
}

// used only for testing
func (s *Script) GetFieldsAsHex () [] string {
	hexFields := make ([] string, len (s.fields))
	for f, field := range s.fields {
		hexFields [f] = field.AsHex ()
	}
	return hexFields
}

/*
func (s *Script) GetRawFieldTypes () string {
	return s.rawFieldTypes
}
*/

// used only in test mode
func (s *Script) AsHex () string {
	return hex.EncodeToString (s.rawBytes)
}

func (s *Script) GetSerializedScript () Script {

	if s.IsNil () || s.IsEmpty () {
		return Script {}
	}

	// if the serialized script is not a stack item, then this is not a serialized script
	serializedScriptIndex := len (s.fields) - 1
	if s.fields [serializedScriptIndex].IsOpcode () {
		return Script {}
	}

	// parse it
	serializedScriptBytes := s.fields [serializedScriptIndex].AsBytes ()
	possibleScript := NewScript (serializedScriptBytes)
	if possibleScript.HasParseError () {
		return Script {}
	}

	// it parses, but it is not valid if it contains OP_INVALIDOPCODE
	for _, field := range possibleScript.fields {
		if field.AsHex () == "OP_INVALIDOPCODE" {
			return Script {}
		}
	}

	// it could be an serialized script
	return possibleScript
}

func (s *Script) IsNil () bool {
	return s.rawBytes == nil
}

func (s *Script) IsEmpty () bool {
	return len (s.fields) == 0
}

// identification of the 7 standard redeemable output types
func (s *Script) IsP2pkOutput () bool {
	scriptLen := len (s.rawBytes)
	if scriptLen != 35 && scriptLen != 67 { return false }

	pubkeyLen := int (s.rawBytes [0])
	if scriptLen <= pubkeyLen { return false }

	return IsValidECPublicKey (s.rawBytes [1 : 1 + pubkeyLen]) && s.rawBytes [scriptLen - 1] == 0xac
}

func (s *Script) IsValidP2pkInput () bool {

	// it must contain a single signature
	return len (s.fields) == 1 && IsValidECSignature (s.fields [0].AsBytes ())
}

func (s *Script) IsValidMultiSigInput () bool {

	// a standard multisig input must have at least 1 field, the extra stack item for the checkmultisig bug
	fieldCount := len (s.fields)
	if fieldCount < 1 { return false }

	// the extra stack item can be anything, so we ignore it
	// all remaining fields must be valid signatures of OP_DUP
	for f := 1; f < fieldCount; f++ {
		if !IsValidECSignature (s.fields [f].AsBytes ()) && s.fields [f].AsHex () != "OP_DUP" {
			return false
		}
	}

	return true
}

func (s *Script) IsValidP2pkhInput () bool {

	// it must contain a signature followed by a public key
	return len (s.fields) == 2 && IsValidECSignature (s.fields [0].AsBytes ()) && IsValidECPublicKey (s.fields [1].AsBytes ())
}

func (s *Script) IsMultiSigOutput () bool {

	// a standard multisig output must have at least 3 fields
	fieldCount := len (s.fields)
	if fieldCount < 3 { return false }

	sigCountIndex := 0
	pubKeyCountIndex := fieldCount - 2

	// everything but the sizes and opcode must be public keys
	for i := sigCountIndex + 1; i < pubKeyCountIndex; i++ {
		if !IsValidECPublicKey (s.fields [i].AsBytes ()) { return false }
	}

	// verify the number of public keys
	pubKeyCount := s.fields [pubKeyCountIndex].rawBytes [0]
	if isValidOpcode (pubKeyCount) && pubKeyCount >= 0x51 && pubKeyCount <= 0x60 { pubKeyCount -= 0x50 }
	if int (pubKeyCount) > fieldCount - 3 { return false }

	// verify the number of signatures
	sigCount := s.fields [sigCountIndex].rawBytes [0]
	if isValidOpcode (sigCount) && sigCount >= 0x51 && sigCount <= 0x60 { sigCount -= 0x50 }
	if sigCount > pubKeyCount { return false }

	// the last field must be OP_CHECKMULTISIG
	lastFieldIndex := fieldCount - 1
	return s.fields [lastFieldIndex].IsOpcode () && s.fields [lastFieldIndex].AsHex () == "OP_CHECKMULTISIG"
}

func (s *Script) IsP2pkhOutput () bool { return len (s.rawBytes) == 25 && s.rawBytes [0] == 0x76 && s.rawBytes [1] == 0xa9 && s.rawBytes [2] == 0x14 && s.rawBytes [23] == 0x88 && s.rawBytes [24] == 0xac }
func (s *Script) IsP2shOutput () bool { return len (s.rawBytes) == 23 && s.rawBytes [0] == 0xa9 && s.rawBytes [1] == 0x14 && s.rawBytes [22] == 0x87 }
func (s *Script) IsP2wpkhOutput () bool { return len (s.rawBytes) == 22 && s.rawBytes [0] == 0x00 && s.rawBytes [1] == 0x14 }
func (s *Script) IsP2wshOutput () bool { return len (s.rawBytes) == 34 && s.rawBytes [0] == 0x00 && s.rawBytes [1] == 0x20 }
func (s *Script) IsTaprootOutput () bool { return len (s.rawBytes) == 34 && s.rawBytes [0] == 0x51 && s.rawBytes [1] == 0x20 }

// identification of the 2 p2sh-wrapped spend types
func (s *Script) IsP2shP2wpkhRedeemScript () bool { return s.IsP2wpkhOutput () }
func (s *Script) IsP2shP2wshRedeemScript () bool { return s.IsP2wshOutput () }

// OP_RETURN required to be first opcode
func (s *Script) IsNullDataOutput () bool { return len (s.rawBytes) >= 1 && s.rawBytes [0] == 0x6a }

func (s *Script) IsNonstandardOutput () bool { return !s.IsTaprootOutput () && !s.IsP2wpkhOutput () && !s.IsP2wshOutput () && !s.IsP2shOutput () && !s.IsP2pkhOutput () && !s.IsMultiSigOutput () && !s.IsP2pkOutput () && !s.IsNullDataOutput () && !s.IsWitnessUnknownOutput () }

func (s *Script) IsWitnessUnknownOutput () bool {
	exactlyTwoFields := len (s.fields) == 2
	if !exactlyTwoFields { return false }

	firstByteIsValidWitnessVersion := s.rawBytes [0] == 0x00 || (s.rawBytes [0] >= 0x51 && s.rawBytes [0] <= 0x60)
	if !firstByteIsValidWitnessVersion { return false }

	validVersion0 := s.IsP2wpkhOutput () || s.IsP2wshOutput ()
	if validVersion0 { return false }

	validVersion1 := s.IsTaprootOutput ()
	if validVersion1 { return false }

	return true
}

func (s *Script) IsOrdinal () bool {

	fieldCount := len (s.fields)
	if fieldCount < 10 { return false }

	if !IsValidSchnorrPublicKey (s.fields [0].AsBytes ()) { return false }
	if s.fields [1].AsHex () != "OP_CHECKSIG" { return false }

	ordBegin := 2
	if s.fields [3].AsHex () == "OP_DROP" { ordBegin = 4 }

	if s.fields [ordBegin].AsHex () != "OP_0" { return false }
	if s.fields [ordBegin + 1].AsHex () != "OP_IF" { return false }
	if s.fields [ordBegin + 2].AsText () != "ord" { return false }
	if s.fields [ordBegin + 3].AsBytes () [0] != 0x01 { return false }

	if s.fields [ordBegin + 5].AsHex () != "OP_0" { return false }

	if s.fields [fieldCount - 1].AsHex () != "OP_ENDIF" { return false }

	return true
}

func isValidOpcode (b byte) bool {
	return (b == 0x00 || b >= 0x4f) && getOpcodeName (b) != "OP_INVALIDOPCODE"
}

// https://github.com/bitcoin/bitcoin/blob/master/src/script/script.h
func getOpcodeName (val byte) string {
	switch val {
		// push value
		case 0x00: return "OP_0"
		case 0x4c: return "OP_PUSHDATA1"
		case 0x4d: return "OP_PUSHDATA2"
		case 0x4e: return "OP_PUSHDATA4"
		case 0x4f: return "OP_1NEGATE"
		case 0x50: return "OP_RESERVED"
		case 0x51: return "OP_1"
		case 0x52: return "OP_2"
		case 0x53: return "OP_3"
		case 0x54: return "OP_4"
		case 0x55: return "OP_5"
		case 0x56: return "OP_6"
		case 0x57: return "OP_7"
		case 0x58: return "OP_8"
		case 0x59: return "OP_9"
		case 0x5a: return "OP_10"
		case 0x5b: return "OP_11"
		case 0x5c: return "OP_12"
		case 0x5d: return "OP_13"
		case 0x5e: return "OP_14"
		case 0x5f: return "OP_15"
		case 0x60: return "OP_16"

		// control
		case 0x61: return "OP_NOP"
		case 0x62: return "OP_VER"
		case 0x63: return "OP_IF"
		case 0x64: return "OP_NOTIF"
		case 0x65: return "OP_VERIF"
		case 0x66: return "OP_VERNOTIF"
		case 0x67: return "OP_ELSE"
		case 0x68: return "OP_ENDIF"
		case 0x69: return "OP_VERIFY"
		case 0x6a: return "OP_RETURN"

		// stack ops
		case 0x6b: return "OP_TOALTSTACK"
		case 0x6c: return "OP_FROMALTSTACK"
		case 0x6d: return "OP_2DROP"
		case 0x6e: return "OP_2DUP"
		case 0x6f: return "OP_3DUP"
		case 0x70: return "OP_2OVER"
		case 0x71: return "OP_2ROT"
		case 0x72: return "OP_2SWAP"
		case 0x73: return "OP_IFDUP"
		case 0x74: return "OP_DEPTH"
		case 0x75: return "OP_DROP"
		case 0x76: return "OP_DUP"
		case 0x77: return "OP_NIP"
		case 0x78: return "OP_OVER"
		case 0x79: return "OP_PICK"
		case 0x7a: return "OP_ROLL"
		case 0x7b: return "OP_ROT"
		case 0x7c: return "OP_SWAP"
		case 0x7d: return "OP_TUCK"

		// splice ops
		case 0x7e: return "OP_CAT"
		case 0x7f: return "OP_SUBSTR"
		case 0x80: return "OP_LEFT"
		case 0x81: return "OP_RIGHT"
		case 0x82: return "OP_SIZE"

		// bit logic
		case 0x83: return "OP_INVERT"
		case 0x84: return "OP_AND"
		case 0x85: return "OP_OR"
		case 0x86: return "OP_XOR"
		case 0x87: return "OP_EQUAL"
		case 0x88: return "OP_EQUALVERIFY"
		case 0x89: return "OP_RESERVED1"
		case 0x8a: return "OP_RESERVED2"

		// numeric
		case 0x8b: return "OP_1ADD"
		case 0x8c: return "OP_1SUB"
		case 0x8d: return "OP_2MUL"
		case 0x8e: return "OP_2DIV"
		case 0x8f: return "OP_NEGATE"
		case 0x90: return "OP_ABS"
		case 0x91: return "OP_NOT"
		case 0x92: return "OP_0NOTEQUAL"

		case 0x93: return "OP_ADD"
		case 0x94: return "OP_SUB"
		case 0x95: return "OP_MUL"
		case 0x96: return "OP_DIV"
		case 0x97: return "OP_MOD"
		case 0x98: return "OP_LSHIFT"
		case 0x99: return "OP_RSHIFT"

		case 0x9a: return "OP_BOOLAND"
		case 0x9b: return "OP_BOOLOR"
		case 0x9c: return "OP_NUMEQUAL"
		case 0x9d: return "OP_NUMEQUALVERIFY"
		case 0x9e: return "OP_NUMNOTEQUAL"
		case 0x9f: return "OP_LESSTHAN"
		case 0xa0: return "OP_GREATERTHAN"
		case 0xa1: return "OP_LESSTHANOREQUAL"
		case 0xa2: return "OP_GREATERTHANOREQUAL"
		case 0xa3: return "OP_MIN"
		case 0xa4: return "OP_MAX"

		case 0xa5: return "OP_WITHIN"

		// crypto
		case 0xa6: return "OP_RIPEMD160"
		case 0xa7: return "OP_SHA1"
		case 0xa8: return "OP_SHA256"
		case 0xa9: return "OP_HASH160"
		case 0xaa: return "OP_HASH256"
		case 0xab: return "OP_CODESEPARATOR"
		case 0xac: return "OP_CHECKSIG"
		case 0xad: return "OP_CHECKSIGVERIFY"
		case 0xae: return "OP_CHECKMULTISIG"
		case 0xaf: return "OP_CHECKMULTISIGVERIFY"

		// expansion
		case 0xb0: return "OP_NOP1"
		case 0xb1: return "OP_CHECKLOCKTIMEVERIFY"
		case 0xb2: return "OP_CHECKSEQUENCEVERIFY"
		case 0xb3: return "OP_NOP4"
		case 0xb4: return "OP_NOP5"
		case 0xb5: return "OP_NOP6"
		case 0xb6: return "OP_NOP7"
		case 0xb7: return "OP_NOP8"
		case 0xb8: return "OP_NOP9"
		case 0xb9: return "OP_NOP10"

		// Opcode added by BIP 342 (Tapscript)
		case 0xba: return "OP_CHECKSIGADD"

		case 0xff: return "OP_INVALIDOPCODE"
	}

	return "OP_INVALIDOPCODE"
}

