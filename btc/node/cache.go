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
	getNodeType () string
	getVersionString () string

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
var blockMap sync.Map // height -> block
var blockIndex sync.Map // hash -> height
var txMap sync.Map // tx-id -> tx

type btcCache struct {

	channel cacheClientChannelPack
	btcNode nodeClient
}

var cache *btcCache = nil

func StartCache () {

	var once sync.Once
	once.Do (initCache)

	// channels
	blockChan := make (chan btc.Block)
	txChan := make (chan btc.Tx)
	cache.channel = cacheClientChannelPack { block: blockChan, tx: txChan }
	cacheThreadChannels := cacheThreadChannelPack { block: blockChan, tx: txChan }

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

	isBip141 := rawTx ["hex"].(string) [8:10] == "00"

	// outputs
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
	blockHeight := int32 (-1)

	// is it already cached?
	if c.isBlockHash (blockKey) {
		blockHash = blockKey
		blockHeight, exists := blockIndex.Load (blockHash)
		if exists {
			b, found := blockMap.Load (blockHeight)
			if found { block = b.(cachedBlock).block }
		}
	} else {
		blockHeight = c.toBlockHeight (blockKey)
		if blockHeight >= 0 {
			b, found := blockMap.Load (blockHeight)
			if found { block = b.(cachedBlock).block }
		}
	}

	if !block.IsNil () {
fmt.Println (fmt.Sprintf ("FOUND: block %d", block.GetHeight ()))
		return block
	}

fmt.Println (fmt.Sprintf ("NOT FOUND: block %s", blockKey))

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
fmt.Println (fmt.Sprintf ("block %s does not seem to exist", blockKey))
				r <- block
				return
			}
		}

		// try to get it from the node
fmt.Println (fmt.Sprintf ("REQUESTING: block %s", blockHash))
		rawBlock, err := c.btcNode.getBlock (blockHash, true)

		succeeded := err == nil
		blockExists := rawBlock != nil

		if !succeeded { fmt.Println (fmt.Sprintf ("NODE ERROR: %s", err.Error ())) }
//		if !blockExists { fmt.Println (fmt.Sprintf ("block %s does not seem to exist", blockHash)) }
		if succeeded && blockExists {

			// create the block and cache it
			block = makeBlock (rawBlock)
			c.channel.block <- block
		}

		// return it to the caller
		r <- block

		// cache the transactions
		txs := rawBlock ["tx"].([] interface {})
		for _, rawTx := range txs {
			txObj := rawTx.(map [string] interface {})
			if txObj ["txid"] == nil { continue }

			txObj ["blockhash"] = block.GetHash ()
			txObj ["blocktime"] = float64 (block.GetTimestamp ())

			c.channel.tx <- makeTx (txObj)
		}
	} (r)

	return <- r
}

func (c *btcCache) threadTxFromNode (txId string, withPreviousOutputs bool, r chan<- btc.Tx) {

fmt.Println (fmt.Sprintf ("REQUESTING: tx %s", txId))
	rawTx, err := c.btcNode.getTx (txId)

	succeeded := err == nil
	txExists := rawTx != nil

	if !succeeded { fmt.Println (fmt.Sprintf ("NODE ERROR: %s", err.Error ())) }
//			if !txExists { fmt.Println (fmt.Sprintf ("tx %s does not seem to exist", txId)) }

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
		}

		// create the tx and cache it
		tx = makeTx (rawTx)
		c.channel.tx <- tx
	}

	// return it to the caller
	r <- tx
}

func (c *btcCache) getTx (txId string, withPreviousOutputs bool) btc.Tx {

	// is it already cached?
	t, found := txMap.Load (txId)
	if found {
		tx := t.(cachedTx).tx
		if !tx.IsNil () {
fmt.Println (fmt.Sprintf ("FOUND: tx %s", tx.GetTxId ()))




			return tx
		}
	}
fmt.Println (fmt.Sprintf ("NOT FOUND: tx %s", txId))

	// it wasn't there, get it from the node
	r := make (chan btc.Tx)
	go c.threadTxFromNode (txId, withPreviousOutputs, r)
	return <- r
}

func (c *btcCache) getOutput (txId string, outputIndex uint16) btc.Output {

	output := btc.Output {}

	// is it already cached?
	tx := btc.Tx {}
	t, found := txMap.Load (txId)
	if found {
		tx = t.(cachedTx).tx
		if tx.IsNil () {
fmt.Println (fmt.Sprintf ("FOUND: NIL tx %s", txId))
			r := make (chan btc.Tx)
			go c.threadTxFromNode (txId, false, r)
			tx = <- r
		}
	} else {
		r := make (chan btc.Tx)
		go c.threadTxFromNode (txId, false, r)
		tx = <- r
	}

	if tx.IsNil () || outputIndex >= tx.GetOutputCount () {
		return output
	}

	// the transaction exists

	return tx.GetOutput (outputIndex)
}

func (c *btcCache) isBlockHash (blockKey string) bool {
	return len (blockKey) == 64
}

func (c *btcCache) toBlockHeight (blockKey string) int32 {
	height, err := strconv.Atoi (blockKey)
	if err != nil { return -1 }
	return int32 (height)
}


func (c *btcCache) run (channel cacheThreadChannelPack) {

	for true {

		select {

			case block := <- channel.block:

				if block.IsNil () { break }

				blockHeight := block.GetHeight ()

				// does it already exist?
				_, found := blockMap.Load (blockHeight)
				if found {
fmt.Println (fmt.Sprintf ("IGNORING: block %d (already cached)", blockHeight))
					break
				}

				// index and cache the block
				go func () {
					now := time.Now ().Unix ()
fmt.Println (fmt.Sprintf ("CACHING: block %d", blockHeight))
					blockMap.Store (blockHeight, cachedBlock { timestampCreated: now, timestampLastUsed: now, block: block })
					blockIndex.Store (block.GetHash (), blockHeight)
				} ()


			case tx := <- channel.tx:

				if tx.IsNil () { break }

				txId := tx.GetTxId ()

				// does it already exist?
				_, found := txMap.Load (txId)
				if found {
fmt.Println (fmt.Sprintf ("IGNORING: tx %s (already cached)", txId))
					break
				}

				// cache the tx
				go func () {
					now := time.Now ().Unix ()
//fmt.Println (fmt.Sprintf ("CACHING: tx %s", txId))
					txMap.Store (txId, cachedTx { timestampCreated: now, timestampLastUsed: now, tx: tx })
				} ()
		}
	}
}

