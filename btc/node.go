package btc

import (
	"fmt"
	"errors"
	"sync"

	"github.com/btc-script-explorer/scantool/app"
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

type BlockRequest struct {
	Hash string
	Height uint32
	Options BlockRequestOptions
}

type TxRequest struct {
	Id string
	Key string
}

// node client

type nodeClient interface {
	GetNodeType () string
	GetVersionString () string

	getBlock (blockHash string, withTxData bool) (map [string] interface {}, error)
	GetBlockHash (blockHeight uint32) string
	GetCurrentBlockHash () string
	GetTx (txId string) Tx
	GetPreviousOutput (txId string, outputIndex uint32) Output
}

func getNode () (nodeClient, error) {
	nodeType := app.Settings.GetNodeType ()

	switch nodeType {
		case "Bitcoin Core":
			bitcoinCore, err := NewBitcoinCore ()
			return bitcoinCore, err
	}

	return nil, errors.New (fmt.Sprintf ("Incorrect node credential, or Unsupported node type %s", nodeType))
}

// node proxy

func makeBlock (rawBlock map [string] interface {}) Block {
	previousHash := ""
	nextHash := ""
	if rawBlock ["previousblockhash"] != nil { previousHash = rawBlock ["previousblockhash"].(string) }
	if rawBlock ["nextblockhash"] != nil { nextHash = rawBlock ["nextblockhash"].(string) }

	return NewBlock (rawBlock ["hash"].(string), previousHash, nextHash, uint32 (rawBlock ["height"].(float64)), int64 (rawBlock ["time"].(float64)), uint16 (len (rawBlock ["tx"].([] interface {}))))
}

func makeTx (rawTx map [string] interface {}) Tx {
return Tx {}
}

type NodeProxy struct {
	n nodeClient

	rawBlockChannel chan map [string] interface {}
	blockChannel chan Block
	blockHashChannel chan string
	blockHeightChannel chan uint32

	rawTxChannel chan map [string] interface {}
	txChannel chan Tx
	txRequestChannel chan string
	txRequestRawChannel chan string

	channelPack cacheChannelPack
}

var proxy *NodeProxy = nil
var proxyError error = nil

func StartNodeProxy () (*NodeProxy, error) {
	var once sync.Once
	once.Do (initNodeProxy)

	go cache (proxy.channelPack)

	return proxy, nil
}

func GetNodeProxy () (*NodeProxy, error) {
	if proxy == nil { return nil, errors.New ("The node proxy has not been started yet.") }
	return proxy, proxyError
}

func initNodeProxy () {
	n, e := getNode ()
	if e != nil { proxyError = e; return }

	proxy = &NodeProxy { n: n }

	// create the channels and the channel pack to communicate with the caching thread

	// block query channels
	proxy.blockHashChannel = make (chan string)
	proxy.blockHeightChannel = make (chan uint32)

	// block data channels
	proxy.rawBlockChannel = make (chan map [string] interface {})
	proxy.blockChannel = make (chan Block)

	// tx query channel
	proxy.txRequestChannel = make (chan string)
	proxy.txRequestRawChannel = make (chan string)

	// tx data channels
	proxy.rawTxChannel = make (chan map [string] interface {})
	proxy.txChannel = make (chan Tx)

	proxy.channelPack = cacheChannelPack {	rawBlockCacheIn: proxy.rawBlockChannel,
											blockOut: proxy.blockChannel,
											blockHashRequest: proxy.blockHashChannel,
											blockHeightRequest: proxy.blockHeightChannel,

//											txCacheIn: proxy.rawTxChannel,
											txOut: proxy.txChannel,
											txOutRaw: proxy.rawTxChannel,
											txRequest: proxy.txRequestChannel,
											txRequestRaw: proxy.txRequestRawChannel }

//											inputRequest: make (chan string) }
}

func (np *NodeProxy) addScriptFields (scriptData map [string] interface {}, script Script) {
	fieldData := make ([] FieldData, script.GetFieldCount ())
	for f, field := range script.GetFields () {
		fieldData [f] = FieldData { Hex: field.AsHex (), Type: field.AsType () }
	}
	scriptData ["Fields"] = fieldData

	if script.HasParseError () {
		scriptData ["ParseError"] = true
	}
}

// return negative height on error
func (np *NodeProxy) GetCurrentBlockHeight () <-chan int32 {

	currentBlockHeight := int32 (-1)

	blockHash := np.n.GetCurrentBlockHash ()
	if len (blockHash) > 0 {
		response, err := np.n.getBlock (blockHash, false)

		if err != nil {
			fmt.Println ("NODE ERROR: " + err.Error ())
		}

		if response != nil {
			currentBlockHeight = int32 (response ["height"].(float64))
		}
	}

	c := make (chan int32)
	c <- currentBlockHeight
	return c
}

func (np *NodeProxy) GetPreviousOutput (requestChannel <-chan PreviousOutputRequest) <-chan Output {

	previousOutputRequest := <- requestChannel

	previousOutput := np.n.GetPreviousOutput (previousOutputRequest.PreviousTxId, previousOutputRequest.PreviousOutputIndex)

outputScript := previousOutput.GetOutputScript ()
scriptFields := outputScript.GetFields ()
fieldData := make ([] FieldData, len (scriptFields))
for f, field := range scriptFields {
	fieldData [f] = FieldData { Hex: field.AsHex (), Type: field.AsType () }
}

	poc := make (chan Output)
	poc <- previousOutput
	return poc
}

func (np *NodeProxy) GetTx (txRequest TxRequest) <-chan Tx {

	tx := np.n.GetTx (txRequest.Id)
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

	tc := make (chan Tx)
	tc <- Tx {}
	return tc
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

func (np *NodeProxy) GetBlock (blockRequest BlockRequest) <-chan Block {

	responseChannel := make (chan Block, 1)

	// first, we ask the cache for it
	np.blockHashChannel <- blockRequest.Hash
	block := <- np.blockChannel
	if !block.IsNil () {
		responseChannel <- block
//fmt.Println (fmt.Sprintf ("Found block %d in the cache!", block.GetHeight ()))
		return responseChannel
	}

	// get the block from the node and cache it
	go func (r chan<- Block) {

//np.txRequestChannel <- "769895:2"
//res := <- np.blockChannel
//fmt.Println ("existing: ", res)

		rawBlock, err := np.n.getBlock (blockRequest.Hash, true)
		if err != nil { fmt.Println (fmt.Sprintf ("NODE ERROR: %s", err.Error ())) }
		if rawBlock == nil { r <- Block {}; return }

		// return the block to the caller and cache it for later retrieval
		r <- makeBlock (rawBlock)
		np.rawBlockChannel <- rawBlock
//b := makeBlock (rawBlock)
//fmt.Println (fmt.Sprintf ("Got block %d from the node.", b.GetHeight ()))
	} (responseChannel)

	return responseChannel
}

