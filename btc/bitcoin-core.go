package btc

import (
	"fmt"
	"io"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"btctx/app"
)

type BitcoinCore struct {
	version string
}

func NewBitcoinCore () BitcoinCore {
	return BitcoinCore { version: "0.0.0" }
}

func (bc *BitcoinCore) GetType () string {
	return "Bitcoin Core"
}

func (bc *BitcoinCore) GetVersion () string {
	return bc.version
}

func (bc *BitcoinCore) GetTx (txId [32] byte) Tx {
	rawTx := bc.getRawTransaction (hex.EncodeToString (txId [:]))
	return bc.parseTx (rawTx)
}

func (bc *BitcoinCore) GetPreviousOutput (txId [32] byte, outputIndex uint32) Output {
	rawTx := bc.getRawTransaction (hex.EncodeToString (txId [:]))
	vout := rawTx ["vout"].([] interface {})
	outputJson := vout [outputIndex].(map [string] interface {})
	return bc.parseOutput(outputJson) 
}

func (bc *BitcoinCore) parseTx (rawTx map [string] interface {}) Tx {

	version := uint32 (rawTx ["version"].(float64))
	lockTime := uint32 (rawTx ["locktime"].(float64))
	blockTime := int64 (rawTx ["blocktime"].(float64))

	hashBytes, err := hex.DecodeString (rawTx ["txid"].(string))
	if err != nil { fmt.Println (err.Error ()); return Tx {} }

	blockHash, err := hex.DecodeString (rawTx ["blockhash"].(string))
	if err != nil { fmt.Println (err.Error ()); return Tx {} }

	// getting this from the raw tx hex string because there isn't another easy way to get it from the Bitcoin Core JSON response
	bip141 := rawTx ["hex"].(string) [8:10] == "00"

	// inputs
	vin := rawTx ["vin"].([] interface {})
	inputCount := len (vin)
	inputs := make ([] Input, inputCount)
	for i := 0; i < int (inputCount); i++ {
		inputJson := vin [i].(map [string] interface {})
		inputs [i] = bc.parseInput (inputJson)
	}

	coinbase := inputs [0].IsCoinbase ()

	// outputs
	vout := rawTx ["vout"].([] interface {})
	outputCount := len (vout)
	outputs := make ([] Output, outputCount)
	for o := 0; o < int (outputCount); o++ {
		outputJson := vout [o].(map [string] interface {})
		outputs [o] = bc.parseOutput (outputJson)
	}

	return Tx {	id: [32] byte (hashBytes),
				blockHash: [32] byte (blockHash),
				blockTime: blockTime,
				version: version,
				coinbase: coinbase,
				bip141: bip141,
				inputs: inputs,
				outputs: outputs,
				lockTime: lockTime }
}

func (bc *BitcoinCore) parseInput (inputJson map [string] interface {}) Input {
	coinbase := inputJson ["coinbase"] != nil
	sequence := uint32 (inputJson ["sequence"].(float64))

	// input script
	var inputScriptBytes [] byte
	var err error
	if coinbase {
		inputScriptBytes, err = hex.DecodeString (inputJson ["coinbase"].(string))
	} else {
		inputScript := inputJson ["scriptSig"].(map [string] interface {})
		inputScriptBytes, err = hex.DecodeString (inputScript ["hex"].(string))
	}
	if err != nil { fmt.Println (err.Error ()); return Input {} }

	script := Script {}
	if len (inputScriptBytes) > 0 {
		script = NewScript (inputScriptBytes)
	}

	//segwit
	segwit := Segwit {}
	if inputJson ["txinwitness"] != nil {
		segwit = bc.parseSegwit (inputJson ["txinwitness"].([] interface {}))
	}

	txType := ""

	// previous output
	if coinbase {
		txType = "COINBASE"
		return Input { coinbase: coinbase, spendType: txType, inputScript: script, segwit: segwit, sequence: sequence }
	}

	previousOutputIndex := uint32 (inputJson ["vout"].(float64))
	previousOutputHash, err := hex.DecodeString (inputJson ["txid"].(string))
	if err != nil { fmt.Println (err.Error ()); return Input {} }

	// attempt to determine the spend type without yet knowing the output type
	// there are public keys that can be parsed into scripts (albeit nonsensical scripts)
	// therefore, we can't rely on the parsability of the last field of the input script alone
	// inputs with a previous p2sh output type have redeem scripts
	redeemScript := NewScript (script.GetSerializedScript ())
//fmt.Println (redeemScript)
	if !redeemScript.HasParseError () {
		if redeemScript.IsP2shP2wshInput () {
			txType = "P2SH-P2WSH"
		} else if redeemScript.IsP2shP2wpkhInput () {
			txType = "P2SH-P2WPKH"
		}
	} else {
		// not a redeem script
		redeemScript = Script {}
	}

	// p2sh-p2wsh and p2wsh inputs have a witness script
	// Taproot Script Path inputs have a tap script
	// Taproot and p2wsh inputs must have an empty input script
	if script.IsP2shP2wshInput () || script.IsEmpty () {
		if segwit.ParseSerializedScript () {
			if script.IsP2shP2wshInput () {
				txType = "P2SH-P2WSH"
			} else {
				txType = "Taproot Script Path"
			}
		}
	}

	return Input {
		coinbase: coinbase,
		previousOutputTxId: [32] byte (previousOutputHash),
		previousOutputIndex: previousOutputIndex,
		spendType: txType,
		inputScript: script,
		redeemScript: redeemScript,
		segwit: segwit,
		sequence: sequence }
}

func (bc *BitcoinCore) parseOutput (outputJson map [string] interface {}) Output {

	value := uint64 (outputJson ["value"].(float64) * 100000000)

	// output script
	outputScript := outputJson ["scriptPubKey"].(map [string] interface {})
	outputScriptBytes, err := hex.DecodeString (outputScript ["hex"].(string))
	if err != nil { fmt.Println (err.Error ()); return Output {} }

	script := Script {}
	if len (outputScriptBytes) > 0 {
		script = NewScript (outputScriptBytes)
	}

	// address
	address := ""
	if outputScript ["address"] != nil {
		address = outputScript ["address"].(string)
	}

	outputType := ""
	if script.IsTaprootOutput () { outputType = "Taproot"
	} else if script.IsP2wpkhOutput () { outputType = "P2WPKH"
	} else if script.IsP2wshOutput () { outputType = "P2WSH"
	} else if script.IsP2shOutput () { outputType = "P2SH"
	} else if script.IsP2pkhOutput () { outputType = "P2PKH"
	} else if script.IsMultiSigOutput () { outputType = "MultiSig"
	} else if script.IsP2pkOutput () { outputType = "P2PK"
	} else if script.IsNullDataOutput () { outputType = "OP_RETURN"
	} else if script.IsWitnessUnknownOutput () { outputType = "Witness Unknown"
	} else if script.IsNonstandardOutput () { outputType = "Non-Standard" }

	return Output { value: value,
					outputScript: script,
					outputType: outputType,
					address: address }
}

func (bc *BitcoinCore) parseSegwit (segwitFieldsHex [] interface {}) Segwit {

	fieldCount := len (segwitFieldsHex)

	fields := make ([] string, fieldCount)
	for s := 0; s < fieldCount; s++ {
		fields [s] = segwitFieldsHex [s].(string)
	}

	return Segwit { fields: fields }
}

func (bc *BitcoinCore) getRawTransaction (txId string) map [string] interface {} {

	jsonResult := bc.getJson ("getrawtransaction", [] interface {} { txId, true })

	var rawResponse map [string] interface {}
	err := json.Unmarshal (jsonResult, &rawResponse)
	if err != nil { fmt.Println (err.Error ()) }

	// TODO: check for error from node in json

	return rawResponse ["result"].(map [string] interface {})
}

func (bc *BitcoinCore) getJson (function string, params [] interface {}) [] byte {

	settings := app.GetSettings ()

	type jsonRequestObject struct {
		Jsonrpc string `json:"jsonrpc"`
		Method string `json:"method"`
		Params [] interface {} `json:"params"`
	}

	// create the JSON request
	requestObject := jsonRequestObject { Jsonrpc: "2.0", Method: function, Params: params }
	requestJsonBytes, err := json.Marshal (requestObject)
	if err != nil { fmt.Println (err.Error ()) }

	// create the HTTP request
	requestUrl := "http://" + settings.Node.GetFullUrl () + "/"
	req, err := http.NewRequest (http.MethodPost, requestUrl, bytes.NewBuffer (requestJsonBytes))
	req.SetBasicAuth (settings.Node.GetUsername (), settings.Node.GetPassword ())

	// get the HTTP response
	client := &http.Client {}
	response, err := client.Do (req)
	if response == nil { fmt.Println ("Node did not respond.") }
	if err != nil { fmt.Println (err.Error ()) }

	// return the JSON response
	json, err := io.ReadAll (response.Body)
	if err != nil { fmt.Println (err.Error ()) }

	return json
}

