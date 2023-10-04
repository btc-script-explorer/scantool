package node

import (
	"fmt"
	"io"
	"errors"
	"bytes"
	"strings"
	"encoding/json"
	"net/http"

	"github.com/btc-script-explorer/scantool/app"
)

type BitcoinCore struct {
	version string
}

func NewBitcoinCore () (*BitcoinCore, error) {
	
	bc := BitcoinCore {}
	bc.version = bc.getVersionStr ()
	if len (bc.version) == 0 { return nil, errors.New ("Failed to connect to Bitcoin Node.") }
	return &bc, nil
}

func (bc *BitcoinCore) getNodeType () string {
	return "Bitcoin Core"
}

func (bc *BitcoinCore) GetVersionString () string {
	return bc.getNodeType () + " " + bc.version
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

func (bc *BitcoinCore) getBlock (blockHash string, withTxData bool) (map [string] interface {}, error) {

	verbosityLevel := 1
	if withTxData { verbosityLevel = 2 }
	jsonResult := bc.getJson ("getblock", [] interface {} { blockHash, verbosityLevel })

	var rawResponse map [string] interface {}
	err := json.Unmarshal (jsonResult, &rawResponse)
	if err != nil { return nil, errors.New ("JSON ERROR: " + err.Error ()) }

	if rawResponse ["error"] != nil { return nil, errors.New ("BITCOIN CORE ERROR: " + rawResponse ["error"].(map [string] interface {}) ["message"].(string)) }
	if rawResponse ["result"] == nil { return nil, errors.New ("BITCOIN CORE ERROR: No response from node.") }

	return rawResponse ["result"].(map [string] interface {}), nil
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

func (bc *BitcoinCore) getTx (txId string) (map [string] interface {}, error) {

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

	if len (jsonResult) == 0 {
		return nil, errors.New ("No result from node.")
	}

	var rawResponse map [string] interface {}
	jsonError := json.Unmarshal (jsonResult, &rawResponse)
	if jsonError != nil {
		return nil, jsonError
	}

	// check for error from node in json
	if rawResponse ["error"] != nil {
		return nil, errors.New ("NODE ERROR: " + rawResponse ["error"].(map [string] interface {}) ["message"].(string))
	}

	return rawResponse ["result"].(map [string] interface {}), nil
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

