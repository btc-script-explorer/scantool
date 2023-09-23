package btc

import (
	"fmt"
	"sync"
	"os"

	"github.com/btc-script-explorer/scantool/app"
)

type NodeClient interface {
	GetNodeType () string
	GetVersionString () string

	GetBlock (blockHash string) Block
	GetBlockHash (blockHeight uint32) string
	GetCurrentBlockHash () string
	GetTx (txId string) Tx
	GetPreviousOutput (txId string, outputIndex uint32) Output
}

// only one node connection currently supported
var node NodeClient = nil
var once sync.Once

func GetNodeClient () NodeClient {
	once.Do (getNode)
	return node
}

func getNode () {
	nodeType := app.Settings.GetNodeType ()

	switch nodeType {
		case "Bitcoin Core":
			bitcoinCore := NewBitcoinCore ()
			node = &bitcoinCore
			break
		default:
			fmt.Println ("Unsupported node type: " + nodeType)
			os.Exit (1)
	}
}

