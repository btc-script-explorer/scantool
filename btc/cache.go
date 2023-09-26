package btc

import (
	"fmt"
	"strconv"
	"time"
)

type cachedBlock struct {
	timestampCreated int64
	timestampUsed int64
	block Block
}

type cachedTx struct {
	timestampCreated int64
	timestampUsed int64
	tx Tx
}

type cacheChannelPack struct {
	rawBlockCacheIn <-chan map [string] interface {}
//	rawBlockCacheOut chan<- map [string] interface {}
	blockOut chan<- Block
	blockHashRequest <-chan string
	blockHeightRequest <-chan uint32

	txRequest <-chan string
	txRequestRaw <-chan string
//	txCacheIn chan<- map [string] interface {}
//	txCacheOut <-chan map [string] interface {}
	txOut chan<- Tx
	txOutRaw chan<- map [string] interface {}

//	inputRequest <-chan string
}

// thread functions

func cache (channelPack cacheChannelPack) {

	// the block index stores the block height as a string
	// this ensures that it will work even for the genesis block
	blockMap := make (map [uint32] cachedBlock) // height
	blockIndex := make (map [string] string) // hash -> height

	rawTxMap := make (map [string] map [string] interface {}) // block:tx
	txMap := make (map [string] cachedTx) // block:tx
	txIndex := make (map [string] string) // id -> block:tx

//	inputMap := make (map [string] cacheData) // block:tx:input

	for {
		select {

			// objects coming in to be cached

			case rawBlock := <- channelPack.rawBlockCacheIn:
				if rawBlock != nil {
					if rawBlock ["height"] == nil { return }

					// if the block already exists, there is nothing left to do
					blockHeight := uint32 (rawBlock ["height"].(float64))
//fmt.Println (fmt.Sprintf ("Received block %d to cache.", blockHeight))
					block := blockMap [blockHeight].block
					if !block.IsNil () { return }
//fmt.Println (fmt.Sprintf ("Caching block %d.", blockHeight))

					now := time.Now ().Unix ()

					// index and cache the block
					blockIndex [rawBlock ["hash"].(string)] = strconv.Itoa (int (blockHeight))
					blockMap [blockHeight] = cachedBlock { timestampCreated: now, timestampUsed: now, block: makeBlock (rawBlock) }

					// index and cache each raw transaction
					txs := rawBlock ["tx"].([] interface {})
					for i, rawTx := range txs {

						txMap := rawTx.(map [string] interface {})
						if txMap ["txid"] == nil { continue }

						txKey := strconv.Itoa (int (blockHeight)) + ":" + strconv.Itoa (i)
						rawTxMap [txKey] = txMap
					}
				}

			// objects being requested from the cache

			case blockHeight := <- channelPack.blockHeightRequest:

				channelPack.blockOut <- blockMap [blockHeight].block
//b := blockMap [blockHeight].block
//fmt.Println (fmt.Sprintf ("requested block %d, returned block %s (%d)", blockHeight, b.GetHash (), b.GetHeight ()))

			case blockHash := <- channelPack.blockHashRequest:

				retrievedBlock := Block {}

				blockHeightStr := blockIndex [blockHash]
				if len (blockHeightStr) > 0 {
					blockHeight, err := strconv.Atoi (blockHeightStr)
					retrievedBlock = blockMap [uint32 (blockHeight)].block

					if err != nil { fmt.Println (err.Error ()) }
				}

				channelPack.blockOut <- retrievedBlock
//fmt.Println (fmt.Sprintf ("requested block %s, returned block %s (%d)", blockHash, retrievedBlock.GetHash (), retrievedBlock.GetHeight ()))

			case txParam := <- channelPack.txRequest:

				retrievedTx := Tx {}

				// determine the key
				txKey := txParam
				if len (txParam) == 64 { txKey = txIndex [txParam] }

				// is it in the cache
				retrievedTx = txMap [txKey].tx
				if retrievedTx.IsNil () {
					if rawTxMap [txKey] != nil {
						tx := makeTx (rawTxMap [txKey])
						now := time.Now ().Unix ()
						txMap [txKey] = cachedTx { timestampCreated: now, timestampUsed: now, tx: tx }
						txIndex [tx.GetTxId ()] = txKey
						retrievedTx = tx
					}
				}

				channelPack.txOut <- retrievedTx
//fmt.Println (fmt.Sprintf ("requested tx %s, returned tx %s (%s)", txKey, retrievedTx.GetTxId (), txIndex [retrievedTx.GetTxId ()]))

			case txParam := <- channelPack.txRequestRaw:

				channelPack.txOutRaw <- rawTxMap [txParam]
//fmt.Println (fmt.Sprintf ("requested tx %s, returned tx %s (%s)", txKey, retrievedTx.GetTxId (), txIndex [retrievedTx.GetTxId ()]))
		}
	}
}

