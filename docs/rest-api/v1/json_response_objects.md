# JSON Response Objects

## Field

Name | Type | When Included
---|---|---|---
hex | string | Always
type | string | When known

## Script

Name | Type | When Included
---|---|---|---
hex | string | Always
fields | [] Field | Always
parse_error | bool | Always

## Segwit

Name | Type | When Included
---|---|---|---
fields | [] Field | Always
witness_script | Script | if present
tap_script | Script | if present

## Input

Name | Type | When Included
---|---|---|---
coinbase | bool | Always
input_script | Script | Always
redeem_script | Script | coinbase = false and include_input_detail = true and previous output type = P2SH
sequence | uint32 | Always
spend_type | string | coinbase = false
previous_output_tx_id | string | coinbase = false
previous_output_index | uint16 | coinbase = false
previous_output | Output | coinbase = false or include_input_detail = true
segwit | Segwit | parent tx supports BIP141

