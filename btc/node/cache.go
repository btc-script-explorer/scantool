package node

import (
	"fmt"
	"errors"
	"encoding/hex"
	"strconv"
	"sync"
	"time"

	"github.com/btc-script-explorer/scantool/app"
	"github.com/btc-script-explorer/scantool/btc"
)

type nodeClient interface {
	GetNodeType () string
	GetVersionString () string

	getBlock (blockHash string, withTxData bool) (map [string] interface {}, error)
	getBlockHash (blockHeight uint32) string
//	GetCurrentBlockHash () string
	getBestBlockHash () string
	getTx (txId string) (map [string] interface {}, error)
//	GetPreviousOutput (txId string, outputIndex uint32) btc.Output
}

func getNode () (nodeClient, error) {

	nodeType := app.Settings.GetNodeType ()

	switch nodeType {
		case "Bitcoin Core":
			bitcoinCore, err := NewBitcoinCore ()
			return bitcoinCore, err
	}

	return nil, errors.New (fmt.Sprintf ("Incorrect node credentials or unsupported node type %s", nodeType))
}

///////////////////////////////////////////////////////////////////////////////////////////////

type cachedBlock struct {
	timestampCreated int64
	timestampLastUsed int64
	block btc.Block
}

type cachedTx struct {
	timestampCreated int64
	timestampLastUsed int64
	tx btc.Tx
}

type cacheClientChannelPack struct {
	rawBlock chan<- map [string] interface {}
	rawTx chan<- map [string] interface {}
}

type cacheThreadChannelPack struct {
	rawBlock <-chan map [string] interface {}
	rawTx <-chan map [string] interface {}
}

// the caches
var blockMap sync.Map // height -> block
var blockIndex sync.Map // hash -> height

//var rawTxMap sync.Map // height:tx-index -> raw-tx
var txMap sync.Map // tx-id -> tx
//var txIndex sync.Map // height:tx-index -> tx-id

var rawInputMap sync.Map //tx-id:input-index

type btcCache struct {

	channel cacheClientChannelPack
	btcNode nodeClient
}

var cache *btcCache = nil

func StartCache () {

	var once sync.Once
	once.Do (initCache)

	// channels
	rawBlockChan := make (chan map [string] interface {})
	rawTxChan := make (chan map [string] interface {})
	cache.channel = cacheClientChannelPack { rawBlock: rawBlockChan, rawTx: rawTxChan }
	cacheThreadChannels := cacheThreadChannelPack { rawBlock: rawBlockChan, rawTx: rawTxChan }

	// start the cache thread
	go cache.run (cacheThreadChannels)
}

func initCache () {

	btcNode, e := getNode ()
	if e != nil { proxyError = e; return }

	cache = &btcCache {	btcNode: btcNode }
}

func GetCache () btcCache {
	if cache == nil {
		fmt.Println ("GetCache called before StartCache. Return empty cache.")
		return btcCache {}
	}

	return *cache
}

func makeBlock (rawBlock map [string] interface {}) btc.Block {
	previousHash := ""
	nextHash := ""
	if rawBlock ["previousblockhash"] != nil { previousHash = rawBlock ["previousblockhash"].(string) }
	if rawBlock ["nextblockhash"] != nil { nextHash = rawBlock ["nextblockhash"].(string) }

	rawTxs := rawBlock ["tx"].([] interface {})
	txCount := len (rawTxs)
	txIds := make ([] string, txCount)
	for t := 0; t < txCount; t++ {
		rawTx := rawTxs [t].(map [string] interface {})
		txIds [t] = rawTx ["txid"].(string)
	}

	return btc.NewBlock (rawBlock ["hash"].(string), previousHash, nextHash, uint32 (rawBlock ["height"].(float64)), int64 (rawBlock ["time"].(float64)), txIds)
}

func makeTx (rawTx map [string] interface {}) btc.Tx {

	// create the array of outputs
	vout := rawTx ["vout"].([] interface {})
	outputCount := len (vout)
	outputs := make ([] btc.Output, outputCount)
	for o := 0; o < int (outputCount); o++ {
		rawOutput := vout [o].(map [string] interface {})

		value := uint64 (rawOutput ["value"].(float64) * 100000000)

		// output script
		outputScript := rawOutput ["scriptPubKey"].(map [string] interface {})
		outputScriptBytes, err := hex.DecodeString (outputScript ["hex"].(string))
		if err != nil { fmt.Println (err.Error ()) }

		script := btc.NewScript (outputScriptBytes)

		// address
		address := ""
		if outputScript ["address"] != nil { address = outputScript ["address"].(string) }

		outputs [o] = btc.NewOutput (value, script, address)
	}

	// We will return empty inputs
	inputs := make ([] btc.Input, len (rawTx ["vin"].([] interface {})))

	return btc.NewTx (	rawTx ["txid"].(string),
						uint32 (rawTx ["version"].(float64)),
						inputs,
						outputs,
						uint32 (rawTx ["locktime"].(float64)),
// check input [0] to find out if it is the coinbase
xxx == 0,
// also, put the previous output data in the inputs for transactions
// when getting inputs, we will only go to the node if the rest of the input is empty data
						rawTx ["hex"].(string) [8:10] == "00",
						rawTx ["blockhash"].(string),
						rawTx ["blocktime"].(int64))
}

// this is a pass-through function
// the current block height is never cached
func (c *btcCache) GetCurrentBlockHeight () <-chan int32 {

	responseChannel := make (chan int32)

	go func (responseChannel chan<- int32) {
		currentBlockHeight := int32 (-1)

		blockHash := c.btcNode.getBestBlockHash ()
		if len (blockHash) > 0 {
			response, err := c.btcNode.getBlock (blockHash, false)

			if err != nil {
				fmt.Println ("NODE ERROR: " + err.Error ())
			}

			if response != nil {
				currentBlockHeight = int32 (response ["height"].(float64))
			}
		}

		responseChannel <- currentBlockHeight
	} (responseChannel)

	return responseChannel
}

func (c *btcCache) getBlock (blockKey string) <-chan btc.Block {

	responseChannel := make (chan btc.Block, 1)
	block := btc.Block {}
	blockHash := ""
	blockHeight := int32 (-1)

	// is it already cached?
	if c.isBlockHash (blockKey) {
fmt.Println (fmt.Sprintf ("block key %s is a hash", blockKey))
		blockHash = blockKey
		block = c.getBlockByHash (blockHash)
	} else {
fmt.Println (fmt.Sprintf ("block key %s is a height", blockKey))
		blockHeight = c.toBlockHeight (blockKey)
		if blockHeight >= 0 {
			block = c.getBlockByHeight (uint32 (blockHeight))
		}
	}

	if !block.IsNil () {
fmt.Println (fmt.Sprintf ("Found in cache: block %d", block.GetHeight ()))
		responseChannel <- block
	} else {
fmt.Println (fmt.Sprintf ("block %s not found in cache", blockKey))

		// it wasn't there
		// get the block from the node, return it to the appropriate channel and cache it
		go func (r chan<- btc.Block) {

			if len (blockHash) == 0 {
				if blockHeight >= 0 {
					blockHash = c.btcNode.getBlockHash (uint32 (blockHeight))
fmt.Println (fmt.Sprintf ("block height %d => block hash %s", blockHeight, blockHash))
				}
				if len (blockHash) == 0 {
fmt.Println (fmt.Sprintf ("block %s does not seem to exist", blockKey))
					r <- block
					return
				}
			}

fmt.Println (fmt.Sprintf ("requesting block %s from the node", blockHash))
			rawBlock, err := c.btcNode.getBlock (blockHash, true)
			if err != nil {
				fmt.Println (fmt.Sprintf ("NODE ERROR: %s", err.Error ()))
				r <- block
				return
			}

			if rawBlock == nil {
fmt.Println (fmt.Sprintf ("block %s does not seem to exist", blockHash))
				r <- block
				return
			}

			// return the block to the caller and cache it for later retrieval
			r <- makeBlock (rawBlock)
fmt.Println (fmt.Sprintf ("received block %s from the node, returning to caller", blockHash))
			c.channel.rawBlock <- rawBlock

//fmt.Println (fmt.Sprintf ("Got block %d from the node.", b.GetHeight ()))
		} (responseChannel)
	}

	return responseChannel


/*
	verbosityLevel := 1
	if withTxData { verbosityLevel = 2 }
	jsonResult := bc.getJson ("getblock", [] interface {} { blockHash, verbosityLevel })

	var rawResponse map [string] interface {}
	err := json.Unmarshal (jsonResult, &rawResponse)
	if err != nil { return nil, errors.New ("JSON ERROR: " + err.Error ()) }

	if rawResponse ["error"] != nil { return nil, errors.New ("BITCOIN CORE ERROR: " + rawResponse ["error"].(map [string] interface {}) ["message"].(string)) }
	if rawResponse ["result"] == nil { return nil, errors.New ("BITCOIN CORE ERROR: No response from node.") }

	return rawResponse ["result"].(map [string] interface {}), nil
*/
}

func (c *btcCache) getTx (txId string) <-chan btc.Tx {


	responseChannel := make (chan btc.Tx, 1)

	// is it already cached?
	retrievedTx, exists := txMap.Load (txId)
	if exists {
		responseChannel <- retrievedTx.(btc.Tx)
	} else {
		// it wasn't there
		// get the tx from the node, return it to the appropriate channel and cache it
		go func (r chan<- btc.Tx) {

//np.txRequestChannel <- "769895:2"
//res := <- np.blockChannel
//fmt.Println ("existing: ", res)

fmt.Println (fmt.Sprintf ("requesting tx %s from the node", txId))
			rawTx, err := c.btcNode.getTx (txId)
			if err != nil {
				fmt.Println (fmt.Sprintf ("NODE ERROR: %s", err.Error ()))
				r <- btc.Tx {}
				return
			}

			if rawTx == nil {
fmt.Println (fmt.Sprintf ("tx %s does not seem to exist", txId))
				r <- btc.Tx {}
				return
			}

			// return the tx to the caller and cache it for later retrieval
			r <- makeTx (rawTx)
fmt.Println (fmt.Sprintf ("received tx %s from the node, returning to caller", txId))
			c.channel.rawTx <- rawTx

//b := makeBlock (rawBlock)
//fmt.Println (fmt.Sprintf ("Got block %d from the node.", b.GetHeight ()))
		} (responseChannel)
	}

//	tx := btc.Tx {}


return responseChannel

/*
			case txParam := <- channel.txRequest:

				// determine the key
				txKey := txParam
				if len (txParam) == 64 {
					k, found := c.txIndex.Load (txParam)
					if found { txKey = k.(string) }
				}

				// is it in the tx cache?
				tx, exists := c.txMap.Load (txKey)
				if exists {
					channel.tx <- tx.(btc.Tx)
					break
				}

				// is it in the raw tx cache?
				rawTx, exists := c.rawTxMap.Load (txKey)
				if exists {

					// process it
					tx := makeTx (rawTx.(map [string] interface {}))
					now := time.Now ().Unix ()

					// add it to the tx cache
					c.txMap.Store (txKey, cachedTx { timestampCreated: now, timestampLastUsed: now, tx: tx })
					c.txIndex.Store (tx.GetTxId (), txKey)

					// remove it from the raw tx cache
					c.rawTxMap.Delete (txKey)

					channel.tx <- tx
					break
				}

				channel.tx <- btc.Tx {}
*/

//fmt.Println (fmt.Sprintf ("requested tx %s, returned tx %s (%s)", txKey, retrievedTx.GetTxId (), txIndex [retrievedTx.GetTxId ()]))

//			case txParam := <- c.channelPack.txRequestRaw:

//				c.channelPack.txOutRaw <- rawTxMap [txParam]
//fmt.Println (fmt.Sprintf ("requested tx %s, returned tx %s (%s)", txKey, retrievedTx.GetTxId (), txIndex [retrievedTx.GetTxId ()]))

}

func (c *btcCache) isBlockHash (blockKey string) bool {
	return len (blockKey) == 64
}

func (c *btcCache) toBlockHeight (blockKey string) int32 {
	height, err := strconv.Atoi (blockKey)
	if err != nil { return -1 }
	return int32 (height)
}


func (c *btcCache) getBlockByHash (blockHash string) btc.Block {

	retrievedBlock := btc.Block {}

	blockHeight, exists := blockIndex.Load (blockHash)
	if exists {
		b, found := blockMap.Load (blockHeight)
		if found { retrievedBlock = b.(cachedBlock).block }
	}

	return retrievedBlock

//fmt.Println (fmt.Sprintf ("requested block %s, returned block %s (%d)", blockHash, retrievedBlock.GetHash (), retrievedBlock.GetHeight ()))

}

func (c *btcCache) getBlockByHeight (blockHeight uint32) btc.Block {

	retrievedBlock := btc.Block {}
fmt.Println (fmt.Sprintf ("loading block %d", blockHeight))
	b, found := blockMap.Load (blockHeight)
fmt.Println (fmt.Sprintf ("found %p", b))
	if found { retrievedBlock = b.(cachedBlock).block }
	return retrievedBlock

//b := blockMap [blockHeight].block
//fmt.Println (fmt.Sprintf ("requested block %d, returned block %s (%d)", blockHeight, b.GetHash (), b.GetHeight ()))

}

func (c *btcCache) run (channel cacheThreadChannelPack) {

// only caching, processing? and cleanup should be done in this thread

	for {
		select {

			// to be cached

			case rawBlock := <- channel.rawBlock:

				if rawBlock != nil {

					if rawBlock ["height"] == nil { break }
					blockHeight := uint32 (rawBlock ["height"].(float64))
fmt.Println (fmt.Sprintf ("cache thread received block %d", blockHeight))

					// if the block already exists, there is nothing left to do
					_, found := blockMap.Load (blockHeight)
					if found {
fmt.Println (fmt.Sprintf ("ignoring request because block %d is already in the cache", blockHeight))
						break
					}

fmt.Println (fmt.Sprintf ("caching block %d", blockHeight))
					now := time.Now ().Unix ()

					// index and cache the block
					block := makeBlock (rawBlock)
fmt.Println (fmt.Sprintf ("created block has hash %s", block.GetHash ()))

fmt.Println (fmt.Sprintf ("storing block %d", blockHeight))
					blockMap.Store (blockHeight, cachedBlock { timestampCreated: now, timestampLastUsed: now, block: block })
b, _ := blockMap.Load (blockHeight)
fmt.Println (fmt.Sprintf ("data stored: %p", b))
					blockIndex.Store (block.GetHash (), blockHeight)

					// index and cache each raw transaction
					txs := rawBlock ["tx"].([] interface {})
					for _, rawTx := range txs {

						txObj := rawTx.(map [string] interface {})
						if txObj ["txid"] == nil { continue }

txObj ["blockhash"] = block.GetHash ()
txObj ["blocktime"] = block.GetTimestamp ()
/*
txObj ["txkey"] = strconv.Itoa (int (block.GetHeight ())) + ":" + strconv.Itoa (i)

for k, v := range txObj {
	fmt.Println (k, " = ", v)
	fmt.Println ()
}
*/

tx := makeTx (txObj)
fmt.Println (fmt.Sprintf ("caching tx %s", tx.GetTxId ()))
						txMap.Store (tx.GetTxId (), tx)
					}
				}


			case rawTx := <- channel.rawTx:

				if rawTx != nil {
				}


		}
	}
}

