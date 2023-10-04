package node

import (
//	"fmt"
//	"errors"
	"sync"

	"github.com/btc-script-explorer/scantool/btc"
)

// data objects

// BlockKey is either a block hash or block height formatted as a string
type BlockRequest struct {
	BlockKey string
}

type TxRequest struct {
	TxId string
	IncludeInputDetail bool
}

type OutputRequest struct {
	TxId string
	OutputIndex uint16
}

// node proxy

type NodeProxy struct {
	cache btcCache
}

var proxy *NodeProxy = nil
var initProxyOnce sync.Once

func GetNodeProxy () (*NodeProxy, error) {
	initProxyOnce.Do (initNodeProxy)
	return proxy, nil
}

func initNodeProxy () {
	proxy = &NodeProxy { cache: GetCache () }
}

// currently returns negative height on error, should return two values
func (np *NodeProxy) GetCurrentBlockHeight () int32 {
	responseChannel := np.cache.getCurrentBlockHeight ()
	return <- responseChannel
}

func (np *NodeProxy) GetBlock (blockRequest BlockRequest) btc.Block {

	blockKey := blockRequest.BlockKey
	if len (blockKey) == 0 { blockKey = np.GetCurrentBlockHash () }

	return np.cache.getBlock (blockKey)
}

func (np *NodeProxy) GetTx (txRequest TxRequest) btc.Tx {
	if len (txRequest.TxId) != 64 { return btc.Tx {} }
	return np.cache.getTx (txRequest.TxId, txRequest.IncludeInputDetail)
}

func (np *NodeProxy) GetOutput (outputRequest OutputRequest) btc.Output {
	if len (outputRequest.TxId) != 64 { return btc.Output {} }
	return np.cache.getOutput (outputRequest.TxId, outputRequest.OutputIndex)
}

func (np *NodeProxy) GetCurrentBlockHash () string {
	return <- np.cache.getCurrentBlockHash ()
}

func (np *NodeProxy) GetNodeVersion () string {
	return np.cache.GetNodeVersionStr ()
}

