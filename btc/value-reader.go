package btc

import (
	"encoding/hex"
	"strconv"
)

func ReadNumeric (rawBytes [] byte) uint64 {
	var val uint64
	for i := len (rawBytes) - 1; i >= 0; i-- {
		val |= uint64 (rawBytes [i])
		if i > 0 {
			val <<= 8
		}
	}
	return val
}

func ReadVarInt (rawBytes [] byte) (uint64, int) {

	byteCount := 1
	firstByte := ReadNumeric (rawBytes [0:1])
	if firstByte <= 0xfc {
		return firstByte, byteCount
	}

	switch firstByte {
		case 0xfd: byteCount += 2; break
		case 0xfe: byteCount += 4; break
		case 0xff: byteCount += 8; break
	}
	
	return ReadNumeric (rawBytes [1 : firstByte]), byteCount
}

func ReverseBytes (rawBytes [] byte) [] byte {

	byteCount := len (rawBytes)
	indexLimit := byteCount - 1
	reversed := make ([] byte, byteCount)
	for b := 0; b < byteCount; b++ {
		reversed [indexLimit - b] = rawBytes [b]
	}
	return reversed
}


func IsValidUncompressedPublicKey (field [] byte) bool {
	return len (field) == 65 && field [0] == 0x04
}

func IsValidCompressedPublicKey (field [] byte) bool {
	return len (field) == 33 && (field [0] == 0x02 || field [0] == 0x03)
}

func IsValidECPublicKey (field [] byte) bool {
	return IsValidCompressedPublicKey (field) || IsValidUncompressedPublicKey (field)
}

func IsValidECSignature (field [] byte) bool {

	fieldLen := len (field)
	if fieldLen < 4 { return false }

	// first byte
	if field [0] != 0x30 { return false }

	// overall length
	signatureLen := int (field [1])
	if fieldLen < signatureLen { return false }
	if field [2] != 0x02 { return false }

	// r
	rLen := int (field [3])
	if fieldLen < rLen + 6 { return false }
	if field [rLen + 4] != 0x02 { return false }

	// s
	sLen := int (field [rLen + 5])
	if rLen + sLen + 4 != signatureLen { return false }

	// sighash byte
	lastByte := field [fieldLen - 1]
	return lastByte == 0x01 || lastByte == 0x02 || lastByte == 0x03 || lastByte == 0x81 || lastByte == 0x82 || lastByte == 0x83
}

func IsValidSchnorrPublicKey (field [] byte) bool {

	// we don't have much to go on here other than the length
	return len (field) == 32
}

func IsValidSchnorrSignature (field [] byte) bool {

	fieldLen := len (field)
	if fieldLen == 64 { return true }
	if fieldLen != 65 { return false }

	lastByte := field [fieldLen - 1]
	return lastByte == 0x01 || lastByte == 0x02 || lastByte == 0x03 || lastByte == 0x81 || lastByte == 0x82 || lastByte == 0x83
}

func GetValueHtml (satoshis uint64) string {
	satoshisStr := strconv.FormatUint (satoshis, 10)
	digitCount := len (satoshisStr)
	if digitCount > 8 {
		btcDigits := digitCount - 8
		satoshisStr = "<span style=\"font-weight:bold;\">" + satoshisStr [0 : btcDigits] + "</span>" + satoshisStr [btcDigits :]
	}
	return satoshisStr
}

func GetStackItemType (fieldText string, bip141 bool, taproot bool) string {

	fieldBytes, _ := hex.DecodeString (fieldText)

	if IsValidECSignature (fieldBytes) { return "Signature (EC)" }
	if IsValidUncompressedPublicKey (fieldBytes) { return "Uncompressed Public Key" }
	if IsValidCompressedPublicKey (fieldBytes) { return "Compressed Public Key" }

	fieldLen := len (fieldBytes)
	s := ""; if fieldLen != 1 { s = "s" }
	return "Data (" + strconv.Itoa (fieldLen) + " Byte" + s + ")"
}

