package node

import (
	"fmt"
	"errors"
	"encoding/hex"
	"strconv"
	"sync"
	"time"
//	"runtime"

	"github.com/shopspring/decimal"

	"github.com/btc-script-explorer/scantool/app"
	"github.com/btc-script-explorer/scantool/btc"
)

type nodeClient interface {
	GetVersionString () string

	getNodeType () string
	getVersionStr () string

	getBlock (blockHash string, withTxData bool) (map [string] interface {}, error)
	getTx (txId string) (map [string] interface {}, error)
	getBlockHash (blockHeight uint32) string
	getBestBlockHash () string
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
	block chan<- btc.Block
	tx chan<- btc.Tx
}

type cacheThreadChannelPack struct {
	block <-chan btc.Block
	tx <-chan btc.Tx
}

// caches
//var blockMap sync.Map // height -> block
var blockIndex sync.Map // hash -> height
//var txMap sync.Map // tx-id -> tx

//var countDataMutex sync.Mutex
//var blockCount uint32
//var txCount uint32

var blockCacheMutex sync.Mutex
var blockMap map [uint32] cachedBlock
//var blockIndex map [string] uint32

var txCacheMutex sync.Mutex
var txMap map [string] cachedTx

type btcCache struct {
	channel cacheClientChannelPack
	btcNode nodeClient
	caching bool
}

var cache *btcCache = nil
var initCacheOnce sync.Once

func initCache () {

	if cache != nil { return }

	cachingOn := app.Settings.IsCachingOn ()

	btcNode, _ := getNode ()
	cache = &btcCache {	btcNode: btcNode, caching: cachingOn }

	if cache.caching {

		blockMap = make (map [uint32] cachedBlock)
		txMap = make (map [string] cachedTx)

		// channels
		blockChan := make (chan btc.Block)
		txChan := make (chan btc.Tx)
		cache.channel = cacheClientChannelPack { block: blockChan, tx: txChan }

		// start the cache thread
		cacheThreadChannels := cacheThreadChannelPack { block: blockChan, tx: txChan }
		go run (cacheThreadChannels)
	}
}

func GetCache () btcCache {
	initCacheOnce.Do (initCache)
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

	isBip141 := rawTx ["hex"].(string) [8:10] == "00"

	// outputs
	vout := rawTx ["vout"].([] interface {})
	outputCount := len (vout)
	outputs := make ([] btc.Output, outputCount)
	for o := 0; o < int (outputCount); o++ {
		rawOutput := vout [o].(map [string] interface {})

//		value := uint64 (rawOutput ["value"].(float64) * 100000000)
		dValue := decimal.NewFromFloat (rawOutput ["value"].(float64))
		value := uint64 (dValue.Mul (decimal.NewFromInt (100000000)).IntPart ())

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

	// inputs
	vin := rawTx ["vin"].([] interface {})
	inputCount := len (vin)
	inputs := make ([] btc.Input, inputCount)
	for i := 0; i < int (inputCount); i++ {
		rawInput := vin [i].(map [string] interface {})

		var inputScriptBytes [] byte
		isCoinbase := i == 0 && rawInput ["coinbase"] != nil
		previousOutputTxId := ""
		previousOutputIndex := uint16 (0)
		if isCoinbase {
			inputScriptBytes, _ = hex.DecodeString (rawInput ["coinbase"].(string))
		} else {
			scriptSig := rawInput ["scriptSig"].(map [string] interface {})
			inputScriptBytes, _ = hex.DecodeString (scriptSig ["hex"].(string))

			previousOutputTxId = rawInput ["txid"].(string)
			previousOutputIndex = uint16 (rawInput ["vout"].(float64))
		}

		segwit := btc.Segwit {}
		if isBip141 {
			segwitFields := make ([] [] byte, 0)
			if rawInput ["txinwitness"] != nil {
				rawSegwitFields := rawInput ["txinwitness"].([] interface {})
				segwitFieldCount := len (rawSegwitFields)
				for s := 0; s < segwitFieldCount; s++ {
					segwitField, _ := hex.DecodeString (rawSegwitFields [s].(string))
					segwitFields = append (segwitFields, segwitField)
				}
			}

			segwit = btc.NewSegwit (segwitFields)
		}

		previousOutput := btc.Output {}
		if rawInput ["previous_output"] != nil {
			rawPreviousOutput := rawInput ["previous_output"].(map [string] interface {})
			outputScriptBytes, _ := hex.DecodeString (rawPreviousOutput ["output_script"].(string))
			previousOutput = btc.NewOutput (rawPreviousOutput ["value"]. (uint64), btc.NewScript (outputScriptBytes), rawPreviousOutput ["address"].(string))
		}

		inputs [i] = btc.NewInput (	isCoinbase,
									previousOutputTxId,
									previousOutputIndex,
									btc.NewScript (inputScriptBytes),
									segwit,
									uint32 (rawInput ["sequence"].(float64)),
									previousOutput)
	}

	return btc.NewTx (	rawTx ["txid"].(string),
						uint32 (rawTx ["version"].(float64)),
						inputs,
						outputs,
						uint32 (rawTx ["locktime"].(float64)),
						inputs [0].IsCoinbase (),
						isBip141,
						rawTx ["blockhash"].(string),
						int64 (rawTx ["blocktime"].(float64)))
}

// this is a pass-through function
// the current block hash is never cached
func (c *btcCache) getCurrentBlockHash () <-chan string {

	responseChannel := make (chan string, 1)

	go func (responseChannel chan<- string) {
		responseChannel <- c.btcNode.getBestBlockHash ()
	} (responseChannel)

	return responseChannel
}

// this is a pass-through function
// the current block height is never cached
func (c *btcCache) getCurrentBlockHeight () <-chan int32 {

	responseChannel := make (chan int32, 1)

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

func (c *btcCache) getBlock (blockKey string) btc.Block {

	block := btc.Block {}

	blockHash := ""
	blockHeight := uint32 (0xffffffff)

	// is it already cached?
	if c.isBlockHash (blockKey) {
		blockHash = blockKey

		indexedBlockHeight, exists := blockIndex.Load (blockHash)
		if exists {
			blockHeight = indexedBlockHeight.(uint32)
		}
	} else {
		blockHeight = c.toBlockHeight (blockKey)
	}

	if c.caching {
		blockCacheMutex.Lock ()
		block = blockMap [blockHeight].block
		blockCacheMutex.Unlock ()

		if !block.IsNil () {
//fmt.Println (fmt.Sprintf ("FOUND: block %d", block.GetHeight ()))
			return block
		}
	}

//fmt.Println (fmt.Sprintf ("NOT FOUND: block %s", blockKey))

	// it wasn't there
	// get the block from the node, return it to the appropriate channel and cache it
	r := make (chan btc.Block)
	go func (r chan<- btc.Block) {

		// make sure we have the block hash
		if len (blockHash) == 0 {
			if blockHeight >= 0 {
				blockHash = c.btcNode.getBlockHash (uint32 (blockHeight))
			}
			if len (blockHash) == 0 {
				r <- block
				return
			}
		}

		// try to get it from the node
//fmt.Println (fmt.Sprintf ("REQUESTING: block %s", blockHash))
		rawBlock, err := c.btcNode.getBlock (blockHash, true)

		succeeded := err == nil
		if !succeeded {
			fmt.Println (fmt.Sprintf ("NODE ERROR: %s", err.Error ()))
			r <- block
			return
		}

		blockExists := rawBlock != nil
		if blockExists {
			block = makeBlock (rawBlock)
		}

		// return it to the caller
		r <- block

		// cache it
		if blockExists {
			if c.caching {
				c.channel.block <- block

				// cache the transactions
				txs := rawBlock ["tx"].([] interface {})
				for _, rawTx := range txs {
					txObj := rawTx.(map [string] interface {})
					if txObj ["txid"] == nil { continue }

					txObj ["blockhash"] = block.GetHash ()
					txObj ["blocktime"] = float64 (block.GetTimestamp ())

					c.channel.tx <- makeTx (txObj)
				}
			}
		}
	} (r)

	return <- r
}

func (c *btcCache) threadTxFromNode (txId string, withPreviousOutputs bool, r chan<- btc.Tx) {

//fmt.Println (fmt.Sprintf ("REQUESTING: tx %s", txId))
//before := time.Now ().Unix ()
	rawTx, err := c.btcNode.getTx (txId)
//after := time.Now ().Unix ()
//fmt.Println (fmt.Sprintf ("read tx from node took %d seconds", after - before))
//fmt.Println (fmt.Sprintf ("RECEIVED: tx %s", txId))

	succeeded := err == nil
	txExists := rawTx != nil

	if !succeeded { fmt.Println (fmt.Sprintf ("NODE ERROR: %s", err.Error ())) }

	tx := btc.Tx {}

	if succeeded && txExists {

		if withPreviousOutputs {

			// get every previous output
			vin := rawTx ["vin"].([] interface {})
			inputCount := len (vin)
			for i := 0; i < int (inputCount); i++ {
				rawInput := vin [i].(map [string] interface {})
				if rawInput ["coinbase"] != nil {
					continue
				}

				previousOutputTxId := rawInput ["txid"].(string)
				previousOutputIndex := uint16 (rawInput ["vout"].(float64))

				prevTx := c.getTx (previousOutputTxId, false)
				previousOutput := prevTx.GetOutput (previousOutputIndex)
				previousOutputScript := previousOutput.GetOutputScript ()

				rawPreviousOutput := make (map [string] interface {})
				rawPreviousOutput ["value"] = previousOutput.GetValue ()
				rawPreviousOutput ["address"] = previousOutput.GetAddress ()
				rawPreviousOutput ["output_script"] = previousOutputScript.AsHex ()
				rawPreviousOutput ["output_type"] = previousOutput.GetOutputType ()
				rawInput ["previous_output"] =  rawPreviousOutput
				vin [i] = rawInput
			}
//fmt.Println (vin)
		}

		// create the tx and cache it
		tx = makeTx (rawTx)

		if c.caching { c.channel.tx <- tx }
	}

	// return it to the caller
	r <- tx
}

func (c *btcCache) getTx (txId string, withPreviousOutputs bool) btc.Tx {

	// is it already cached?
//	t, found := txMap.Load (txId)

	found := false

	t := cachedTx {}
	if c.caching {

		txCacheMutex.Lock ()
		t = txMap [txId]
		txCacheMutex.Unlock ()

		found = !t.tx.IsNil ()
	}

	if found {

//		txCacheObj := t.(cachedTx)
		txCacheObj := t
		tx := txCacheObj.tx

		if !tx.IsNil () {
//fmt.Println (fmt.Sprintf ("FOUND: tx %s", tx.GetTxId ()))

			includesPreviousOutputs := tx.IsCoinbase ()
			if !tx.IsCoinbase () {
				firstInput := tx.GetInput (0)
				firstPrevOut := firstInput.GetPreviousOutput ()
				includesPreviousOutputs = len (firstPrevOut.GetOutputType ()) != 0
			}

			if withPreviousOutputs && !includesPreviousOutputs {

				// get the previous outputs, re-evaluate the inputs and re-cache the transaction
				for i, input := range tx.GetInputs () {
					previousOutput := c.getOutput (input.GetPreviousOutputTxId (), input.GetPreviousOutputIndex ())
					tx.SetPreviousOutput (uint16 (i), previousOutput)
				}
			}

			if c.caching {
				now := time.Now ().Unix ()
				updatedTx := cachedTx { timestampCreated: txCacheObj.timestampCreated, timestampLastUsed: now, tx: tx }

//				txMap.Swap (txId, updatedTx)

				txCacheMutex.Lock ()
				txMap [txId] = updatedTx
				txCacheMutex.Unlock ()
			}

			return tx
		}
	}
//fmt.Println (fmt.Sprintf ("NOT FOUND: tx %s", txId))

	// it wasn't there, get it from the node
	r := make (chan btc.Tx, 1)
	go c.threadTxFromNode (txId, withPreviousOutputs, r)
	return <- r
}

func (c *btcCache) getOutput (txId string, outputIndex uint16) btc.Output {

	// is it already cached?
	tx := c.getTx (txId, false)
	if tx.IsNil () || outputIndex >= tx.GetOutputCount () {
		return btc.Output {}
	}

	return tx.GetOutput (outputIndex)
}

func (c *btcCache) GetNodeVersionStr () string {
	return c.btcNode.GetVersionString ()
}

func (c *btcCache) isBlockHash (blockKey string) bool {
	return len (blockKey) == 64
}

func (c *btcCache) toBlockHeight (blockKey string) uint32 {
	height, err := strconv.Atoi (blockKey)
	if err != nil { return 0xffffffff }
	return uint32 (height)
}

func run (channel cacheThreadChannelPack) {

	for true {

		select {

			case block := <- channel.block:

				if block.IsNil () { break }

				blockHeight := block.GetHeight ()

				// does it already exist?
//				_, found := blockMap.Load (blockHeight)

				blockCacheMutex.Lock ()
				blockInCache := blockMap [blockHeight].block
				blockCacheMutex.Unlock ()

				found := !blockInCache.IsNil ()
				if found {
//fmt.Println (fmt.Sprintf ("IGNORING: block %d (already cached)", blockHeight))
					break
				}

				// index and cache the block
				go func () {
					now := time.Now ().Unix ()
//fmt.Println (fmt.Sprintf ("CACHING: block %d", blockHeight))

//					blockMap.Store (blockHeight, cachedBlock { timestampCreated: now, timestampLastUsed: now, block: block })

					blockCacheMutex.Lock ()
					blockMap [blockHeight] = cachedBlock { timestampCreated: now, timestampLastUsed: now, block: block }
					blockCacheMutex.Unlock ()

					blockIndex.Store (block.GetHash (), blockHeight)
				} ()


			case tx := <- channel.tx:

				if tx.IsNil () { break }

				txId := tx.GetTxId ()

				// does it already exist?
//				_, found := txMap.Load (txId)

				txCacheMutex.Lock ()
				t := txMap [txId]
				txCacheMutex.Unlock ()

				found := !t.tx.IsNil ()

				if found {
//fmt.Println (fmt.Sprintf ("IGNORING: tx %s (already cached)", txId))
					break
				}

				// cache the tx
				go func () {
					now := time.Now ().Unix ()
//fmt.Println (fmt.Sprintf ("CACHING: tx %s", txId))

//					txMap.Store (txId, cachedTx { timestampCreated: now, timestampLastUsed: now, tx: tx })

					txCacheMutex.Lock ()
					txMap [txId] = cachedTx { timestampCreated: now, timestampLastUsed: now, tx: tx }
//					txCount := len (txMap)
//if txCount >= 100 {
//	for k, _ := range txMap {
//		delete (txMap, k)
//	}
//}
					txCacheMutex.Unlock ()

//var ms runtime.MemStats
//runtime.ReadMemStats (&ms)
//if txCount == 0 { txCount = 1 }
//fmt.Println (fmt.Sprintf ("***** %d threads, %d bytes allocated, %d transactions (%d bytes/tx) *****", runtime.NumGoroutine (), ms.HeapAlloc, txCount, ms.HeapAlloc / uint64 (txCount)))
				} ()
		}
	}
}

