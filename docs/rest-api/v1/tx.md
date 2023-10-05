# JSON Request Objects

## TxOptions

Name | Type | Required | Default | Description
:---:|:---:|:---:|:---:|:---:
include_input_detail | bool | No | false | (see below)
human_readable | bool | No | false | return human readable JSON

## TxRequest

Name | Type | Required | Default | Description
:---:|:---:|:---:|:---:|:---:
id | string | Yes | | transaction id
options | TxOptions | No | not included | options

# How to request a transaction

## The include_input_detail Option

The include_input_detail option will force the SCANTOOL to return detailed information about every input.
For the sake of efficiency, if caching is on and the input detail is already in the cache, a TxResponse might include detailed data even if it was not requested.

include_input_detail | Advantage | Disadvantage | Best Use Case
:---:|:---:|:---:|:---:
true | requires only one request | may take longer | automated script
false | responses may be faster | requires multiple requests | user interface

Input response field | include_input_detail=true | include_input_detail=false
:---:|:---:|:---:
coinbase | Yes | Yes
input_script | Yes | Yes
segwit | Yes, including serialized scripts | Yes, but without serialized scripts
sequence | Yes | Yes
redeem_script | Yes | No
spend_type | Yes | No
previous_output | if coinbase=false | No
previous_output_tx_id | if coinbase=false | if coinbase=false
previous_output_index | if coinbase=false | if coinbase=false


