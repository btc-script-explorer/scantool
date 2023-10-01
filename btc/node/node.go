package node

import (
//	"fmt"
	"errors"
	"sync"

	"github.com/btc-script-explorer/scantool/btc"
)

// data objects

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
	InputsVerbose bool
}

type OutputRequest struct {
	TxId string
	OutputIndex uint16
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

// currently returns negative height on error
func (np *NodeProxy) GetCurrentBlockHeight () int32 {
	responseChannel := np.cache.getCurrentBlockHeight ()
	return <- responseChannel
}

/*
func (np *NodeProxy) GetPreviousOutput (requestChannel <-chan btc.PreviousOutputRequest) <-chan btc.Output {


//	previousOutputRequest := <- requestChannel

//	previousOutput := np.cache.GetPreviousOutput (previousOutputRequest.PreviousTxId, previousOutputRequest.PreviousOutputIndex)

//outputScript := previousOutput.GetOutputScript ()
//scriptFields := outputScript.GetFields ()
//fieldData := make ([] FieldData, len (scriptFields))
//for f, field := range scriptFields {
//	fieldData [f] = FieldData { Hex: field.AsHex (), Type: field.AsType () }
//}

	poc := make (chan btc.Output)
//	poc <- previousOutput
	return poc
}
*/

func (np *NodeProxy) GetBlock (blockRequest BlockRequest) btc.Block {

	blockKey := blockRequest.BlockKey
	if len (blockKey) == 0 { blockKey = np.GetCurrentBlockHash () }

	return np.cache.getBlock (blockKey)
}

func (np *NodeProxy) GetTx (txRequest TxRequest) btc.Tx {

	if len (txRequest.TxId) != 64 { return btc.Tx {} }
	return np.cache.getTx (txRequest.TxId, txRequest.InputsVerbose)
}

func (np *NodeProxy) GetOutput (outputRequest OutputRequest) btc.Output {

	if len (outputRequest.TxId) != 64 { return btc.Output {} }
	return np.cache.getOutput (outputRequest.TxId, outputRequest.OutputIndex)
}

func (np *NodeProxy) GetBlockHash (blockHeight uint32) string {
	return "0000000000000000000137ced007fddf254c01c8771f2e8591db63b3cd531b2e"
}

func (np *NodeProxy) GetCurrentBlockHash () string {
	return <- np.cache.getCurrentBlockHash ()
}

func (np *NodeProxy) GetNodeVersion () string {
	return "Testing"
}

