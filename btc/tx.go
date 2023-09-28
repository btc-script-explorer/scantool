package btc

import (
)

type Tx struct {
	id string
	version uint32
	inputs [] Input
	outputs [] Output
	lockTime uint32

	coinbase bool
	bip141 bool

	blockHash string
	blockTime int64
}

func NewTx (id string, version uint32, inputs [] Input, outputs [] Output, lockTime uint32, coinbase bool, bip141 bool, blockHash string, blockTime int64) Tx {
	return Tx { id: id, version: version, inputs: inputs, outputs: outputs, lockTime: lockTime, coinbase: coinbase, bip141: bip141, blockHash: blockHash, blockTime: blockTime }
}

func (tx *Tx) IsNil () bool {
	return len (tx.id) == 0
}

func (tx *Tx) IsCoinbase () bool {
	return tx.coinbase
}

func (tx *Tx) GetTxId () string {
	return tx.id
}

func (tx *Tx) GetBlockHash () string {
	return tx.blockHash
}

func (tx *Tx) GetBlockTime () int64 {
	return tx.blockTime
}

func (tx *Tx) GetVersion () uint32 {
	return tx.version
}

func (tx *Tx) SupportsBip141 () bool {
	return tx.bip141
}

func (tx *Tx) GetInputCount () uint16 {
	return uint16 (len (tx.inputs))
}

func (tx *Tx) GetInputs () [] Input {
	return tx.inputs
}

func (tx *Tx) GetOutputCount () uint16 {
	return uint16 (len (tx.outputs))
}

func (tx *Tx) GetOutputs () [] Output {
	return tx.outputs
}

func (tx *Tx) GetLockTime () uint32 {
	return tx.lockTime
}

