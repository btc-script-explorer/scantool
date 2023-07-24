package btc

import (
	"fmt"
	"sync"
	"os"

	"btctx/app"
)

type NodeClient interface {
	GetType () string
	GetVersionString () string
	GetBlock (blockHash string) Block
	GetBlockHash (blockHeight int) string
	GetCurrentBlockHash () string
	GetTx (txId string) Tx
	GetPreviousOutput (txId string, outputIndex uint32) Output
}

// singleton, only one node connection currently supported
var node NodeClient = nil
var once sync.Once

func GetNodeClient () NodeClient {
	once.Do (getNode)
	return node
}

func getNode () {
	settings := app.GetSettings ()
	nodeType := settings.Node.GetNodeType ()

	switch nodeType {
		case "BitcoinCore":
			bitcoinCore := NewBitcoinCore ()
			node = &bitcoinCore
			break
		default:
			fmt.Println ("Unsupported node type: " + nodeType)
			os.Exit (1)
	}
}

