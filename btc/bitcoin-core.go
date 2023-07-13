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

	block := bc.getBlock (rawTx ["blockhash"].(string))
	blockHeight := uint32 (block ["height"].(float64))

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

	return Tx {	id: [32] byte (hashBytes),
				blockHeight: blockHeight,
				blockTime: blockTime,
				blockHash: [32] byte (blockHash),
				version: version,
				coinbase: coinbase,
				bip141: bip141,
				inputs: inputs,
				outputs: outputs,
				lockTime: lockTime }
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

	redeemScript := inputScript.GetSerializedScript ()

//////////////////////////////////////////////////////////////////
	// if this is a coinbase input, we have everything we need
	if coinbase {
		return NewInput (coinbase, [32] byte {}, 0, "COINBASE", inputScript, redeemScript, segwit, sequence)
	}

	possibleSpendType := ""

	// determine a possible/likely spend type
	if !segwit.IsNil () && !segwit.IsEmpty () {
		// there are segregated witness fields

		if !inputScript.IsEmpty () {

			// it has an non-empty input script
			// it is one of the p2sh-wrapped spend types
			if !redeemScript.IsNil () {
				if redeemScript.IsP2shP2wpkhRedeemScript () {
					possibleSpendType = "P2SH-P2WPKH"
				} else if redeemScript.IsP2shP2wshRedeemScript () {
					possibleSpendType = "P2SH-P2WSH"
				} else {
					// this should be impossible
					fmt.Println ("Segwit and Input Script exist, but redeem script is not a p2sh-wrapped script.")
				}
			} else {

				// there is a segregated witness and a non-empty input script, but no redeem script
				// this should be impossible
				fmt.Println ("Segwit and Input Script exist, but no redeem script.")

			}

		} else {

			// it has an empty input script
			// it is one of the witness types
			possibleWitnessScript := segwit.GetWitnessScript ()
			possibleTapScript, _ := segwit.GetTapScript ()

			if !possibleWitnessScript.IsNil () && possibleTapScript.IsNil () {
				// a Schnorr Signature can parse as a valid script, but if it is a valid Taproot Key Path spend then we assume that is the correct identification
				if segwit.IsValidTaprootKeyPath () {
					fmt.Println ("Segwit field is possible P2WSH witness script or Taproot Key Path Schnorr Signature. Assuming Taproot Key Path.")
					possibleSpendType = "Taproot Key Path"
				} else {
					possibleSpendType = "P2WSH"
				}
			} else if possibleWitnessScript.IsNil () && !possibleTapScript.IsNil () {
				possibleSpendType = "Taproot Script Path"
			} else if possibleWitnessScript.IsNil () && possibleTapScript.IsNil () {
				// there are no serialized scripts, it is one of the key-based witness spend types
				if segwit.IsValidTaprootKeyPath () {
					possibleSpendType = "Taproot Key Path"
				} else if segwit.IsValidP2wpkh () {
					possibleSpendType = "P2WPKH"
				} else {
					// this should be impossible
					fmt.Println ("Segregated Witness has no parsable serialized script but is not a key-based spend type either.")
				}
			} else {
				// both a witness script and a tap script are parsable
				// here we will assume that it is a Taproot Script Path spend
				fmt.Println ("Input has a parsable witness script and also a parsable tap script. Assuming Taproot Script Path.")
				possibleSpendType = "Taproot Script Path"
			}
		}

	} else {

		// no segregated witness fields
		// if we have a parsable redeem script, it could be p2sh
		// otherwise it is p2pk, multisig, p2pkh or nonstandard
		if !redeemScript.IsNil () {
			possibleSpendType = "P2SH"
		}
	}
//////////////////////////////////////////////////////////////////

	// previous output
	previousOutputIndex := uint32 (inputJson ["vout"].(float64))
	previousOutputHash, err := hex.DecodeString (inputJson ["txid"].(string))
	if err != nil { fmt.Println (err.Error ()); return Input {} }

	return NewInput (coinbase, [32] byte (previousOutputHash), previousOutputIndex, possibleSpendType, inputScript, redeemScript, segwit, sequence)
/*
	return Input {
		coinbase: coinbase,
		previousOutputTxId: [32] byte (previousOutputHash),
		previousOutputIndex: previousOutputIndex,
		spendType: possibleSpendType,
		inputScript: inputScript,
		redeemScript: redeemScript,
		segwit: segwit,
		sequence: sequence }
*/
}

func (bc *BitcoinCore) parseOutput (outputJson map [string] interface {}) Output {

	value := uint64 (outputJson ["value"].(float64) * 100000000)

	// output script
	outputScript := outputJson ["scriptPubKey"].(map [string] interface {})
	outputScriptBytes, err := hex.DecodeString (outputScript ["hex"].(string))
	if err != nil { fmt.Println (err.Error ()); return Output {} }

//	script := Script {}
//	if len (outputScriptBytes) > 0 {
		script := NewScript (outputScriptBytes)
//	}

	// address
	address := ""
	if outputScript ["address"] != nil {
		address = outputScript ["address"].(string)
	}

	return NewOutput (value, script, address)
}

func (bc *BitcoinCore) parseSegwit (segwitFieldsHex [] interface {}) Segwit {

	fieldCount := len (segwitFieldsHex)

	fields := make ([] string, fieldCount)
	for s := 0; s < fieldCount; s++ {
		fields [s] = segwitFieldsHex [s].(string)
	}

	return NewSegwit (fields)
}

func (bc *BitcoinCore) getBlock (blockHash string) map [string] interface {} {
	jsonResult := bc.getJson ("getblock", [] interface {} { blockHash, 1 })

	var rawResponse map [string] interface {}
	err := json.Unmarshal (jsonResult, &rawResponse)
	if err != nil { fmt.Println (err.Error ()) }

	// TODO: check for error from node in json

	return rawResponse ["result"].(map [string] interface {})
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

