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

### How to request a transaction

The include_input_detail determines which fields will be included in the Input objects in the TxResponse.

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

Here are the effects of using each option.

include_input_detail | Advantage | Disadvantage
:---:|:---:|:---:
true | all data returned in one request | may take longer
false | initial response is fast | may request a separate request per input

