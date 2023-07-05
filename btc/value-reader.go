package btc

import (
	"strconv"
)

type ValueReader struct {
}

func (vr *ValueReader) ReadNumeric (rawBytes [] byte) uint64 {
	var val uint64
	for i := len (rawBytes) - 1; i >= 0; i-- {
		val |= uint64 (rawBytes [i])
		if i > 0 {
			val <<= 8
		}
	}
	return val
}

func (vr *ValueReader) ReadVarInt (rawBytes [] byte) (uint64, int) {

	byteCount := 1
	firstByte := vr.ReadNumeric (rawBytes [0:1])
	if firstByte <= 0xfc {
		return firstByte, byteCount
	}

	switch firstByte {
		case 0xfd: byteCount += 2; break
		case 0xfe: byteCount += 4; break
		case 0xff: byteCount += 8; break
	}
	
	return vr.ReadNumeric (rawBytes [1 : firstByte]), byteCount
}

func (vr *ValueReader) ReverseBytes (rawBytes [] byte) [] byte {

	byteCount := len (rawBytes)
	indexLimit := byteCount - 1
	reversed := make ([] byte, byteCount)
	for b := 0; b < byteCount; b++ {
		reversed [indexLimit - b] = rawBytes [b]
	}
	return reversed
}


func IsValidPublicKey (field [] byte) bool {
	fieldLen := len (field)
	if fieldLen != 33 && fieldLen != 65 {
		return false
	}

	firstByte := field [0]
	if fieldLen == 33 && firstByte != 0x02 && firstByte != 0x03 {
		return false
	}
	if fieldLen == 65 && firstByte != 0x04 {
		return false
	}

	return true
}

func IsValidECSignature (field [] byte) bool {
	fieldLen := len (field)

	if fieldLen < 55 || fieldLen > 78 {
		return false
	}

	if field [0] != 0x30 {
		return false
	}

	lastByte := field [fieldLen - 1]
	validSighashByte := lastByte == 0x01 || lastByte == 0x02 || lastByte == 0x03 || lastByte == 0x81 || lastByte == 0x82 || lastByte == 0x83
	return validSighashByte
}

func IsValidSchnorrSignature (field [] byte) bool {
	fieldLen := len (field)

	if fieldLen == 64 { return true }
	if fieldLen != 65 { return false }

	lastByte := field [fieldLen - 1]
	validSighashByte := lastByte == 0x01 || lastByte == 0x02 || lastByte == 0x03 || lastByte == 0x81 || lastByte == 0x82 || lastByte == 0x83
	return validSighashByte
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

