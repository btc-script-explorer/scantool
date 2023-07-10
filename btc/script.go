package btc

import (
	"fmt"
	"strings"
	"encoding/hex"

	"btctx/app"
)

type Script struct {
	rawBytes [] byte
	fields [] string
	fieldTypes string
	parseError bool
}

func NewScript (rawBytes [] byte) Script {

	if rawBytes == nil {
		return Script {}
	}

	valueReader := ValueReader {}

	parseError := false
	hexFields := ""
	fieldTypes := ""

	pos := 0
	bytesRemaining := len (rawBytes)

	// parse the script
	for bytesRemaining > 0 {

		if len (hexFields) > 0 {
			hexFields += "|";
		}

		// is it an opcode?
		if rawBytes [pos] == 0x00 || rawBytes [pos] >= 0x4f {

			hexFields += hex.EncodeToString (rawBytes [pos : pos + 1])
			fieldTypes += "o"

			pos++
			bytesRemaining--
			continue
		}

		// it is a stack item
		fieldTypes += "s"

		nextByte := rawBytes [pos]
		pos++
		bytesRemaining--

		fieldLen := 0
		if nextByte < 0x4c {
			fieldLen = int (nextByte)
		} else {
			// it is a stack item using a push data opcode
			valueSize := 0
			switch nextByte {
				case 0x4c: valueSize = 1; break
				case 0x4d: valueSize = 2; break
				case 0x4e: valueSize = 4; break
			}

			// we must make sure there are enough bytes left, or else there is a parse error
			if bytesRemaining >= valueSize {
				fieldLen = int (valueReader.ReadNumeric (rawBytes [pos : pos + valueSize]))
				pos += valueSize
				bytesRemaining -= valueSize
			} else {
				fieldLen = 1
				pos += bytesRemaining
				bytesRemaining = 0
			}
		}

		if fieldLen > bytesRemaining {
			parseError = true
			fieldLen = int (bytesRemaining)
		}
		hexFields += hex.EncodeToString (rawBytes [pos : pos + fieldLen])

		pos += fieldLen
		bytesRemaining -= fieldLen
	}

	// build the human readable script item list
	scriptFieldCount := len (fieldTypes)
	fields := make ([] string, scriptFieldCount)

	if scriptFieldCount > 0 {
		scriptFieldsHex := strings.Split (hexFields, "|")

		for f := 0; f < scriptFieldCount; f++ {
			if fieldTypes [f : f + 1] == "o" {
				opcodeValue, err := hex.DecodeString (scriptFieldsHex [f])
				if err != nil { fmt.Println (err.Error ()); return Script {} }

				fields [f] = getOpcodeName (opcodeValue [0])
			} else {
				fields [f] = scriptFieldsHex [f]
			}
		}
	}

	return Script { rawBytes: rawBytes, fields: fields, fieldTypes: fieldTypes, parseError: parseError }
}

type ScriptFieldHtmlData struct {
	DisplayText string
	ShowCopyButton bool
	CopyImageUrl string
	CopyText string
}

type ScriptHtmlData struct {
	HtmlId string
	Title string
	Fields [] ScriptFieldHtmlData
}

func (s *Script) GetHtmlData (title string, htmlId string, maxWidthCh uint16) ScriptHtmlData {

	scriptFields := [] ScriptFieldHtmlData (nil)

	if !s.IsEmpty () {
		settings := app.GetSettings ()
		copyImageUrl := "http://" + settings.Website.GetFullUrl () + "/image/clipboard-copy.png"
		fieldCount := len (s.fields); if s.parseError { fieldCount++ }

		scriptFields = make ([] ScriptFieldHtmlData, fieldCount)
		for f, field := range s.fields {
			if uint16 (len (field)) > maxWidthCh {
				scriptFields [f] = ScriptFieldHtmlData { DisplayText: field [0 : maxWidthCh - 2] + "...", ShowCopyButton: true, CopyImageUrl: copyImageUrl, CopyText: field }
			} else {
				scriptFields [f] = ScriptFieldHtmlData { DisplayText: field, ShowCopyButton: false }
			}
		}

		if s.parseError {
			scriptFields [fieldCount - 1] = ScriptFieldHtmlData { DisplayText: "< PARSE ERROR >", ShowCopyButton: false }
		}
	} else {
		scriptFields = make ([] ScriptFieldHtmlData, 1)
		scriptFields [0] = ScriptFieldHtmlData { DisplayText: "Empty", ShowCopyButton: false }
	}

	return ScriptHtmlData { HtmlId: htmlId, Title: title, Fields: scriptFields }
}

func (s *Script) GetTextHtmlData (title string, htmlId string, maxWidthCh uint16) ScriptHtmlData {

	scriptFields := make ([] ScriptFieldHtmlData, 1)

	if !s.IsEmpty () {
		scriptText := GetHexFieldHtml (string (s.rawBytes [:]), int (maxWidthCh))
		if uint16 (len (scriptText)) > maxWidthCh {
			settings := app.GetSettings ()
			copyImageUrl := "http://" + settings.Website.GetFullUrl () + "/image/clipboard-copy.png"
			scriptFields [0] = ScriptFieldHtmlData { DisplayText: scriptText [0 : maxWidthCh - 2] + "...", ShowCopyButton: true, CopyImageUrl: copyImageUrl, CopyText: scriptText }
		} else {
			scriptFields [0] = ScriptFieldHtmlData { DisplayText: scriptText, ShowCopyButton: false }
		}
	} else {
		scriptFields [0] = ScriptFieldHtmlData { DisplayText: "Empty", ShowCopyButton: false }
	}

	return ScriptHtmlData { HtmlId: htmlId, Title: title, Fields: scriptFields }
}

func (s *Script) GetFields () [] string {
	return s.fields
}

func (s *Script) HasParseError () bool {
	return s.parseError
}

func (s *Script) GetHex () string {
	return hex.EncodeToString (s.rawBytes)
}

func (s *Script) GetSerializedScript () Script {

	if s.IsNil () || s.IsEmpty () {
		return Script {}
	}

	serializedScriptIndex := len (s.GetFields ()) - 1

	// in the rare case of a zero-length serialized script, we must check for OP_0 and return an empty script
	serializedScriptHex := s.GetFields () [serializedScriptIndex]
	if serializedScriptHex == "OP_0" {
		return NewScript (make ([] byte, 0))
	}

	// by this point, the serialized script must be interpretted as a stack item, not an opcode
	if s.fieldTypes [serializedScriptIndex] != 's' {
		return Script {}
	}

	serializedScriptBytes, err := hex.DecodeString (serializedScriptHex)
	if err != nil { fmt.Println (err.Error ()); return Script {} }

	possibleScript := NewScript (serializedScriptBytes)
	if possibleScript.HasParseError () {
		return Script {}
	}

	return possibleScript
}

func (s *Script) IsNil () bool {
	return s.rawBytes == nil
}

func (s *Script) IsEmpty () bool {
	return ! s.IsNil () && len (s.rawBytes) == 0
}

// identification of 7 standard output types
func (s *Script) IsP2pkOutput () bool { scriptLen := len (s.rawBytes); return (scriptLen == 35 || scriptLen == 67) && IsValidPublicKey (s.rawBytes [1 : 1 + s.rawBytes [0]]) && s.rawBytes [scriptLen - 1] == 0xac }
func (s *Script) IsMultiSigOutput () bool { scriptLen := len (s.rawBytes); return scriptLen >= 3 && int (s.rawBytes [scriptLen - 2]) == len (s.GetFields ()) - 3 && s.rawBytes [scriptLen - 1] == 0xae }
func (s *Script) IsP2pkhOutput () bool { return len (s.rawBytes) == 25 && s.rawBytes [0] == 0x76 && s.rawBytes [1] == 0xa9 && s.rawBytes [2] == 0x14 && s.rawBytes [23] == 0x88 && s.rawBytes [24] == 0xac }
func (s *Script) IsP2shOutput () bool { return len (s.rawBytes) == 23 && s.rawBytes [0] == 0xa9 && s.rawBytes [1] == 0x14 && s.rawBytes [22] == 0x87 }
func (s *Script) IsP2wpkhOutput () bool { return len (s.rawBytes) == 22 && s.rawBytes [0] == 0x00 && s.rawBytes [1] == 0x14 }
func (s *Script) IsP2wshOutput () bool { return len (s.rawBytes) == 34 && s.rawBytes [0] == 0x00 && s.rawBytes [1] == 0x20 }
func (s *Script) IsTaprootOutput () bool { return len (s.rawBytes) == 34 && s.rawBytes [0] == 0x51 && s.rawBytes [1] == 0x20 }

func (s *Script) IsP2shP2wpkhRedeemScript () bool { return s.IsP2wpkhOutput () }
func (s *Script) IsP2shP2wshRedeemScript () bool { return s.IsP2wshOutput () }

// OP_RETURN required to be first opcode here, might be slightly different than Bitcoin Core
func (s *Script) IsNullDataOutput () bool { return len (s.rawBytes) >= 1 && s.rawBytes [0] == 0x6a }
func (s *Script) IsNonstandardOutput () bool { return !s.HasParseError () && !s.IsTaprootOutput () && !s.IsP2wpkhOutput () && !s.IsP2wshOutput () && !s.IsP2shOutput () && !s.IsP2pkhOutput () && !s.IsMultiSigOutput () && !s.IsP2pkOutput () && !s.IsNullDataOutput () && !s.IsWitnessUnknownOutput () }

func (s *Script) IsWitnessUnknownOutput () bool {
	exactlyTwoFields := len (s.GetFields ()) == 2
	if !exactlyTwoFields { return false }

	firstByteIsValidWitnessVersion := s.rawBytes [0] == 0x00 || (s.rawBytes [0] >= 0x51 && s.rawBytes [0] <= 0x60)
	if !firstByteIsValidWitnessVersion { return false }

	validVersion0 := s.IsP2wpkhOutput () || s.IsP2wshOutput ()
	if validVersion0 { return false }

	validVersion1 := s.IsTaprootOutput ()
	if validVersion1 { return false }

	return true
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

