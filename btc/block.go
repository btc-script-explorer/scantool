package btc

import (
)

type Block struct {
	hash string
	previousHash string
	nextHash string
	height uint32
	timestamp int64
	txs [] Tx
}

func NewBlock (hash string, previous string, next string, height uint32, timestamp int64, txs [] Tx) Block {
	return Block { hash: hash, previousHash: previous, nextHash: next, height: height, timestamp: timestamp, txs: txs }
}

func (b *Block) IsNil () bool {
	return len (b.hash) == 0
}

func (b *Block) GetHash () string {
	return b.hash
}

func (b *Block) GetPreviousHash () string {
	return b.previousHash
}

func (b *Block) GetNextHash () string {
	return b.nextHash
}

func (b *Block) GetHeight () uint32 {
	return b.height
}

func (b *Block) GetTxs () [] Tx {
	return b.txs
}

func (b *Block) GetTx (index int) Tx {
	return b.txs [index]
}

func (b *Block) GetTimestamp () int64 {
	return b.timestamp
}

// returns inputs, outputs
func (b *Block) GetInputOutputCounts () (uint16, uint16) {

	inputCount := uint16 (0)
	outputCount := uint16 (0)
	for _, tx := range b.txs {
		inputCount += tx.GetInputCount ()
		outputCount += tx.GetOutputCount ()
	}
	return inputCount, outputCount
}

func (b *Block) GetPendingPreviousOutputs () map [string] [] uint32 {

	unknownPreviousOutputTypes := make (map [string] [] uint32)
	for _, tx := range b.txs {
		for _, input := range tx.GetInputs () {
			if input.IsCoinbase () { continue }

			spendType := input.GetSpendType ()
			if len (spendType) == 0 {
				unknownPreviousOutputTypes [input.previousOutputTxId] = append (unknownPreviousOutputTypes [input.previousOutputTxId], input.GetPreviousOutputIndex ())
			}
		}
	}

	return unknownPreviousOutputTypes
}

