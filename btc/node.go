package btc

import (
	"fmt"
	"sync"
	"os"

	"btctx/app"
)

/*
Electrum Servers

https://github.com/spesmilo/electrumx (Python)
https://electrumx-spesmilo.readthedocs.io/en/latest/
https://electrum.readthedocs.io/en/latest/

https://github.com/romanz/electrs (Rust)

*/

type NodeClient interface {
	GetType () string
	GetVersionString () string
	GetBlock (blockHash string, verbose bool) Block
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

