# JSON Response Objects

## Field

Name | Type
---|---
hex | string
type | string

## Script

Name | Type
---|---
hex | string
fields | [] Field
parse_error | bool

## Segwit

Name | Type
---|---
fields | [] Field
witness_script | Script
tap_script | Script

## Input

Name | Type
---|---
coinbase | bool
input_script | Script
redeem_script | Script
sequence | uint32
spend_type | string
previous_output_tx_id | string
previous_output_index | uint16
previous_output | Output
segwit | Segwit

## Output

Name | Type
---|---
address | string
output_script | Script
output_type | string
value | uint64

## Tx

Name | Type
---|---
id | string
version | uint32
inputs | [] Input
outputs | [] Output
locktime | uint32
coinbase | bool
bip141 | bool
blockhash | string
blocktime | int64

## Block

Name | Type
---|---
hash | string
previous_hash | string
next_hash | string
height | uint32
timestamp | int64
tx_ids | [] string

