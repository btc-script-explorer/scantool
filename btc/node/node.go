package node

import (
//	"fmt"
	"errors"
	"sync"

	"github.com/btc-script-explorer/scantool/btc"
)

// data objects

type FieldData struct {
	Hex string
	Type string
}

type TxOutput struct {
	OutputIndex uint32
	OutputType string
	Value uint64
	Address string
	OutputScript map [string] interface {}
}

type BlockRequestOptions struct {
	HumanReadable bool
}

// BlockKey is either a block hash or block height formatted as a string
type BlockRequest struct {
	BlockKey string
	Options BlockRequestOptions
}

type TxRequest struct {
	TxId string
}

// node client

var proxy *NodeProxy = nil
var proxyError error = nil

func GetNodeProxy () (*NodeProxy, error) {
	if proxy == nil { return nil, errors.New ("The node proxy has not been initialized.") }
	return proxy, proxyError
}

func StartNodeProxy () (*NodeProxy, error) {
	var once sync.Once
	once.Do (initNodeProxy)

	return proxy, nil
}

func initNodeProxy () {

	proxy = &NodeProxy {}

	StartCache ()
	proxy.cache = GetCache ()
}

// node proxy

type NodeProxy struct {
	cache btcCache
}


func (np *NodeProxy) addScriptFields (scriptData map [string] interface {}, script btc.Script) {
	fieldData := make ([] FieldData, script.GetFieldCount ())
	for f, field := range script.GetFields () {
		fieldData [f] = FieldData { Hex: field.AsHex (), Type: field.AsType () }
	}
	scriptData ["Fields"] = fieldData

	if script.HasParseError () {
		scriptData ["ParseError"] = true
	}
}

// currently returns negative height on error
func (np *NodeProxy) GetCurrentBlockHeight () int32 {
	responseChannel := np.cache.GetCurrentBlockHeight ()
	return <- responseChannel
}

func (np *NodeProxy) GetPreviousOutput (requestChannel <-chan btc.PreviousOutputRequest) <-chan btc.Output {

/*
	previousOutputRequest := <- requestChannel

	previousOutput := np.cache.GetPreviousOutput (previousOutputRequest.PreviousTxId, previousOutputRequest.PreviousOutputIndex)

outputScript := previousOutput.GetOutputScript ()
scriptFields := outputScript.GetFields ()
fieldData := make ([] FieldData, len (scriptFields))
for f, field := range scriptFields {
	fieldData [f] = FieldData { Hex: field.AsHex (), Type: field.AsType () }
}
*/

	poc := make (chan btc.Output)
//	poc <- previousOutput
	return poc
}

func (np *NodeProxy) GetBlock (blockRequest BlockRequest) btc.Block {

	if len (blockRequest.BlockKey) == 0 { return btc.Block {} }

	responseChannel := np.cache.getBlock (blockRequest.BlockKey)
	return <- responseChannel
}

func (np *NodeProxy) GetTx (txRequest TxRequest) btc.Tx {

	if len (txRequest.TxId) == 0 { return btc.Tx {} }

	responseChannel := np.cache.getTx (txRequest.TxId)
	return <- responseChannel

/*
	tx, err := np.cache.getTx (txRequest.Id)
_=tx
_=err
	if tx.IsNil () { return nil }

	txData := make (map [string] interface {})

	txData ["BlockHeight"] = tx.GetBlockHeight ()
	txData ["BlockTime"] = tx.GetBlockTime ()
	txData ["BlockHash"] = tx.GetBlockHash ()
	txData ["Id"] = tx.GetTxId ()
	txData ["IsCoinbase"] = tx.IsCoinbase ()
	txData ["SupportsBip141"] = tx.SupportsBip141 ()
	txData ["LockTime"] = tx.GetLockTime ()

	// inputs
	inputs := make ([] map [string] interface {}, tx.GetInputCount ())
	for i, input := range tx.GetInputs () {

		inputData := make (map [string] interface {})

		inputData ["InputIndex"] = uint32 (i)
		inputData ["Coinbase"] = input.IsCoinbase ()
		inputData ["SpendType"] = input.GetSpendType ()
		inputData ["PreviousOutputTxId"] = input.GetPreviousOutputTxId ()
		inputData ["PreviousOutputIndex"] = input.GetPreviousOutputIndex ()
		inputData ["Sequence"] = input.GetSequence ()

		// input script
		inputScript := input.GetInputScript ()
		if !inputScript.IsNil () {
			inputScriptData := make (map [string] interface {})
			np.addScriptFields (inputScriptData, inputScript)
			inputData ["InputScript"] = inputScriptData
		}

		// redeem script
		redeemScript := input.GetRedeemScript ()
		if !redeemScript.IsNil () {
			redeemScriptData := make (map [string] interface {})
			np.addScriptFields (redeemScriptData, redeemScript)
			redeemScriptData ["Multisig"] = input.HasMultisigRedeemScript ()
			inputData ["RedeemScript"] = redeemScriptData
		}

		// segwit
		segwit := input.GetSegwit ()
		if !segwit.IsEmpty () {

			segwitData := make (map [string] interface {})

			// segwit fields
			fieldData := make ([] FieldData, segwit.GetFieldCount ())
			for f, field := range segwit.GetFields () {
				fieldData [f] = FieldData { Hex: field.AsHex (), Type: field.AsType () }
			}
			segwitData ["Fields"] = fieldData

			// witness script
			witnessScript := segwit.GetWitnessScript ()
			if !witnessScript.IsNil () {
				witnessScriptData := make (map [string] interface {})
				np.addScriptFields (witnessScriptData, witnessScript)
				witnessScriptData ["Multisig"] = input.HasMultisigWitnessScript ()
				segwitData ["WitnessScript"] = witnessScriptData
			}

			// tap script
			tapScript, _ := segwit.GetTapScript ()
			if !tapScript.IsNil () {
				tapScriptData := make (map [string] interface {})
				np.addScriptFields (tapScriptData, tapScript)
				tapScriptData ["Ordinal"] = tapScript.IsOrdinal ()
				segwitData ["TapScript"] = tapScriptData
			}

			inputData ["Segwit"] = segwitData
		}

		inputs [i] = inputData
	}
	txData ["Inputs"] = inputs
//	txData ["PreviousOutputRequests"] = api.getPreviousOutputRequestData (tx)

	// outputs
	outputs := make ([] TxOutput, tx.GetOutputCount ())
	for o, output := range tx.GetOutputs () {

		outputScript := output.GetOutputScript ()

		outputScriptData := make (map [string] interface {})
		np.addScriptFields (outputScriptData, outputScript)

		outputs [o] = TxOutput { OutputIndex: uint32 (o), OutputType: output.GetOutputType (), Value: output.GetValue (), Address: output.GetAddress (), OutputScript: outputScriptData }
	}
	txData ["Outputs"] = outputs

	tc := make (chan btc.Tx)
	tc <- btc.Tx {}
	return tc
*/
}

func (np *NodeProxy) GetBlockHash (blockHeight uint32) string {
	return "0000000000000000000137ced007fddf254c01c8771f2e8591db63b3cd531b2e"
}

func (np *NodeProxy) GetCurrentBlockHash () string {
	return "0000000000000000000137ced007fddf254c01c8771f2e8591db63b3cd531b2e"
}

func (np *NodeProxy) GetNodeVersion () string {
	return "Testing"
}

