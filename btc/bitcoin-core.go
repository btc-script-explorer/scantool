package btc

import (
	"fmt"
	"io"
	"bytes"
	"strings"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/btc-script-explorer/scantool/app"
)

type BitcoinCore struct {
	version string
}

func NewBitcoinCore () BitcoinCore {
	
	bc := BitcoinCore {}
	bc.version = bc.getVersionStr ()
	return bc
}

func (bc *BitcoinCore) GetNodeType () string {
	return "Bitcoin Core"
}

func (bc *BitcoinCore) GetVersionString () string {
	return bc.GetNodeType () + " " + bc.version
}

func (bc *BitcoinCore) GetTx (txId string) Tx {
	rawTx := bc.getRawTransaction (txId)
	if rawTx == nil { return Tx {} }
	return bc.parseTx (rawTx, true)
}

func (bc *BitcoinCore) GetPreviousOutput (txId string, outputIndex uint32) Output {
	rawTx := bc.getRawTransaction (txId)
	vout := rawTx ["vout"].([] interface {})
	outputJson := vout [outputIndex].(map [string] interface {})
	return bc.parseOutput(outputJson) 
}

func (bc *BitcoinCore) GetBlock (blockHash string, verbose bool) Block {

	block := bc.getBlock (blockHash, verbose)
	if block == nil { return Block {} }

	txs := make ([] Tx, 0)

	if verbose {
		txArray := block ["tx"].([] interface {})
		txs = make ([] Tx, len (txArray))
		for t, tx := range txArray {
			parsedTx := bc.parseTx (tx.(map [string] interface {}), false)
			if parsedTx.IsNil () {
				return Block {} // should never happen
			}
			txs [t] = parsedTx
		}
	}

	previousHash := ""
	nextHash := ""
	if block ["previousblockhash"] != nil { previousHash = block ["previousblockhash"].(string) }
	if block ["nextblockhash"] != nil { nextHash = block ["nextblockhash"].(string) }

	return NewBlock (block ["hash"].(string), previousHash, nextHash, uint32 (block ["height"].(float64)), int64 (block ["time"].(float64)), txs)
}

func (bc *BitcoinCore) GetBlockHash (blockHeight uint32) string {
	return bc.getBlockHash (blockHeight)
}

func (bc *BitcoinCore) GetCurrentBlockHash () string {
	return bc.getBestBlockHash ()
}

func (bc *BitcoinCore) parseTx (rawTx map [string] interface {}, includeBlockHeight bool) Tx {

	if rawTx ["txid"] == nil {
		return Tx {}
	}

	hashBytes := rawTx ["txid"].(string)
	version := uint32 (rawTx ["version"].(float64))
	lockTime := uint32 (rawTx ["locktime"].(float64))

	// getting this from the raw tx hex string because there isn't another easy way to get it from the Bitcoin Core JSON response
	bip141 := rawTx ["hex"].(string) [8:10] == "00"

	blockTime := int64 (0)
	if rawTx ["blocktime"] != nil {
		blockTime = int64 (rawTx ["blocktime"].(float64))
	}

	blockHash := ""
	if rawTx ["blockhash"] != nil {
		blockHash = rawTx ["blockhash"].(string)
	}

	// inputs
	vin := rawTx ["vin"].([] interface {})
	inputCount := len (vin)
	inputs := make ([] Input, inputCount)
	for i := 0; i < int (inputCount); i++ {
		inputJson := vin [i].(map [string] interface {})
		inputs [i] = bc.parseInput (inputJson, bip141)
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

	tx := Tx { id: hashBytes, blockTime: blockTime, blockHash: blockHash, version: version, coinbase: coinbase, bip141: bip141, inputs: inputs, outputs: outputs, lockTime: lockTime }

	if includeBlockHeight {
		block := bc.getBlock (rawTx ["blockhash"].(string), false)
		tx.blockHeight = uint32 (block ["height"].(float64))
	}

	return tx
}

func (bc *BitcoinCore) parseInput (inputJson map [string] interface {}, bip141 bool) Input {

	coinbase := inputJson ["coinbase"] != nil
	sequence := uint32 (inputJson ["sequence"].(float64))

	// input script
	var inputScriptBytes [] byte
	var err error
	if coinbase {
		inputScriptBytes, err = hex.DecodeString (inputJson ["coinbase"].(string))
	} else {
		scriptSig := inputJson ["scriptSig"].(map [string] interface {})
		inputScriptBytes, err = hex.DecodeString (scriptSig ["hex"].(string))
	}
	if err != nil { fmt.Println (err.Error ()); return Input {} }

	inputScript := NewScript (inputScriptBytes)

	// segregated witness
	segwit := Segwit {}
	if bip141 {
		if inputJson ["txinwitness"] != nil {
			segwit = bc.parseSegwit (inputJson ["txinwitness"].([] interface {}))
		} else {
			segwit = bc.parseSegwit (make ([] interface {}, 0))
		}
	}

	// if this is a coinbase input, we have everything we need
	if coinbase {
		return NewInput (coinbase, "", 0, inputScript, segwit, sequence)
	}

	// previous output
	previousOutputIndex := uint32 (inputJson ["vout"].(float64))
	previousOutputHash, err := hex.DecodeString (inputJson ["txid"].(string))
	if err != nil { fmt.Println (err.Error ()); return Input {} }

	return NewInput (coinbase, hex.EncodeToString (previousOutputHash), previousOutputIndex, inputScript, segwit, sequence)
}

func (bc *BitcoinCore) parseOutput (outputJson map [string] interface {}) Output {

	value := uint64 (outputJson ["value"].(float64) * 100000000)

	// output script
	outputScript := outputJson ["scriptPubKey"].(map [string] interface {})
	outputScriptBytes, err := hex.DecodeString (outputScript ["hex"].(string))
	if err != nil { fmt.Println (err.Error ()); return Output {} }

	script := NewScript (outputScriptBytes)

	// address
	address := ""
	if outputScript ["address"] != nil {
		address = outputScript ["address"].(string)
	}

	return NewOutput (value, script, address)
}

func (bc *BitcoinCore) parseSegwit (segwitFieldsHex [] interface {}) Segwit {

	fieldCount := len (segwitFieldsHex)

	fields := make ([] [] byte, fieldCount)
	for s := 0; s < fieldCount; s++ {
		fields [s], _ = hex.DecodeString (segwitFieldsHex [s].(string))
	}

	return NewSegwit (fields)
}

func (bc *BitcoinCore) getVersionStr () string {
	networkInfo := bc.getNetworkInfo ()
	if networkInfo == nil { return "" }
	if networkInfo ["subversion"] == nil { return "" }

	// the version must be extracted from the subversion field
	versionStr := networkInfo ["subversion"].(string)
	if strings.Contains (versionStr, ":") {
		parts := strings.Split (versionStr, ":")
		versionStr = parts [1]
	}

	if strings.Contains (versionStr, "/") {
		versionStr = strings.Replace (versionStr, "/", "", -1)
	}

	return versionStr
}

// API functions

func (bc *BitcoinCore) getBlock (blockHash string, withTxData bool) map [string] interface {} {

	verbosityLevel := 1
	if withTxData { verbosityLevel = 2 }
	jsonResult := bc.getJson ("getblock", [] interface {} { blockHash, verbosityLevel })

	var rawResponse map [string] interface {}
	err := json.Unmarshal (jsonResult, &rawResponse)
	if err != nil { fmt.Println (err.Error ()) }

	// check for error from node in json
	if rawResponse ["error"] != nil {
		fmt.Println (rawResponse ["error"].(map [string] interface {}) ["message"])
		return nil
	}

	return rawResponse ["result"].(map [string] interface {})
}

func (bc *BitcoinCore) getBestBlockHash () string {
	jsonResult := bc.getJson ("getbestblockhash", [] interface {} {})
	if len (jsonResult) == 0 { return "" }

	var rawResponse map [string] interface {}
	err := json.Unmarshal (jsonResult, &rawResponse)
	if err != nil { fmt.Println (err.Error ()) }

	// check for error from node in json
	if rawResponse ["error"] != nil {
		fmt.Println (rawResponse ["error"].(map [string] interface {}) ["message"])
		return ""
	}

	return rawResponse ["result"].(string)
}

func (bc *BitcoinCore) getBlockHash (blockHeight uint32) string {
	jsonResult := bc.getJson ("getblockhash", [] interface {} { blockHeight })

	var rawResponse map [string] interface {}
	err := json.Unmarshal (jsonResult, &rawResponse)
	if err != nil { fmt.Println (err.Error ()) }

	// check for error from node in json
	if rawResponse ["error"] != nil {
		fmt.Println (rawResponse ["error"].(map [string] interface {}) ["message"])
		return ""
	}

	return rawResponse ["result"].(string)
}

func (bc *BitcoinCore) getRawTransaction (txId string) map [string] interface {} {

	jsonResult := [] byte (nil)

	if txId != "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b" {
		jsonResult = bc.getJson ("getrawtransaction", [] interface {} { txId, true })
	} else {
		// the genesis transaction is a special case
		// Bitcoin Core won't return it with this API so we handle that case here
		// if other raw transaction JSON fields are used in the future, they might need to be added here
		rawJson := `{
						"result": {
							"txid": "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
							"version": 1,
							"size": 204,
							"vsize": 204,
							"weight": 816,
							"locktime": 0,
							"vin": [
								{
									"coinbase": "04ffff001d0104455468652054696d65732030332f4a616e2f32303039204368616e63656c6c6f72206f6e206272696e6b206f66207365636f6e64206261696c6f757420666f722062616e6b73",
									"sequence": 4294967295
								}
							],
							"vout": [
								{
									"value": 50.00000000,
									"n": 0,
									"scriptPubKey": {
										"hex": "4104678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac",
										"type": "pubkey"
									}
								}
							],
							"hex": "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4d04ffff001d0104455468652054696d65732030332f4a616e2f32303039204368616e63656c6c6f72206f6e206272696e6b206f66207365636f6e64206261696c6f757420666f722062616e6b73ffffffff0100f2052a01000000434104678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac00000000",
							"blockhash":"000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f",
							"blocktime": 1231006505
						},
						"error": null
					}`
		jsonResult = make ([] byte, len (rawJson))
		jsonResult = [] byte (rawJson)
	}

	if len (jsonResult) == 0 { return nil }

	var rawResponse map [string] interface {}
	err := json.Unmarshal (jsonResult, &rawResponse)
	if err != nil { fmt.Println (err.Error ()) }

	// check for error from node in json
	if rawResponse ["error"] != nil {
		fmt.Println (rawResponse ["error"].(map [string] interface {}) ["message"])
		return nil
	}

	return rawResponse ["result"].(map [string] interface {})
}

func (bc *BitcoinCore) getNetworkInfo () map [string] interface {} {
	jsonResult := bc.getJson ("getnetworkinfo", [] interface {} {})
	if len (jsonResult) == 0 { return map [string] interface {} {} }

	var rawResponse map [string] interface {}
	err := json.Unmarshal (jsonResult, &rawResponse)
	if err != nil { fmt.Println (err.Error ()) }

	// check for error from node in json
	if rawResponse ["error"] != nil {
		fmt.Println (rawResponse ["error"].(map [string] interface {}) ["message"])
		return map [string] interface {} {}
	}

	return rawResponse ["result"].(map [string] interface {})
}

func (bc *BitcoinCore) getJson (function string, params [] interface {}) [] byte {

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
	requestUrl := "http://" + app.Settings.GetNodeFullUrl () + "/"
	req, err := http.NewRequest (http.MethodPost, requestUrl, bytes.NewBuffer (requestJsonBytes))
	if err != nil { fmt.Println (err.Error ()) }

	req.SetBasicAuth (app.Settings.GetNodeUsername (), app.Settings.GetNodePassword ())

	// get the HTTP response
	client := &http.Client {}
	response, err := client.Do (req)
	if err != nil { fmt.Println (err.Error ()) }
	if response == nil {
		fmt.Println ("Node returned empty response.")
		if err != nil { fmt.Println (err.Error ()) }
		return [] byte {}
	}

	// return the JSON response
	json, err := io.ReadAll (response.Body)
	if err != nil { fmt.Println (err.Error ()) }

	return json
}

