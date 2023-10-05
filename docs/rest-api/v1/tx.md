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

## include_input_detail = true

The include_input_detail option will force the SCANTOOL to return detailed information about every input.

Considerations:

- Detailed data will be returned for all inputs with only one request required.
- It will probably be slower per request, but faster overall.
- Better for automated scripts where response times are not as important.

Input fields returned in the response:

- coinbase
- input_script
- segwit (including serialized scripts)
- sequence
- redeem_script
- spend_type
- previous_output (if coinbase=false)
- previous_output_tx_id (if coinbase=false)
- previous_output_index (if coinbase=false)

## include_input_detail = false

The include_input_detail option will not force the SCANTOOL to return detailed information about every input.
Howeve, for the sake of efficiency, if caching is on and the input detail is already in the cache, a TxResponse might include detailed input data even if it was not requested.

Considerations:

- Only basic information about inputs will be returned in the initial response.
- Subsquent requests will be required in order to get detailed information about each input.
- It will probably be faster per request, but slower overall.
- Better for user interfaces using the REST API as a backend where response times are important.

Input fields returned in the response:

- coinbase
- input_script
- segwit (not including serialized scripts)
- sequence
- previous_output_tx_id (if coinbase=false)
- previous_output_index (if coinbase=false)

# Examples

## Using include_input_detail = false

TxRequest

        {
                "id": "bbe9e2fced55a2ac4fabb6c74cef8d9dda6cc1121a9782ee7e34fe97e32958cb",
                "options": {
                        "human_readable": true
                }
        }

        $ curl -X POST -d '{"id":"bbe9e2fced55a2ac4fabb6c74cef8d9dda6cc1121a9782ee7e34fe97e32958cb","options":{"human_readable":true}}' http://127.0.0.1:8888/rest/v1/tx

TxResponse

This response includes no previous output detail, no spend type and no serialized script information.
It appears rather ordinary on the surface.

        {
                "bip141": false,
                "blockhash": "0000000000000000001e7b7b2bf28d3e3813e0a37354f95288fe05fa18a0bc8a",
                "blocktime": 1540380059,
                "coinbase": false,
                "id": "bbe9e2fced55a2ac4fabb6c74cef8d9dda6cc1121a9782ee7e34fe97e32958cb",
                "inputs": [
                        {
                                "coinbase": false,
                                "input_script": {
                                        "fields": [
                                                {
                                                        "hex": "OP_0",
                                                        "type": "OP_0"
                                                },
                                                {
                                                        "hex": "00210384945b5a9ab287e5f09ea64f0c2e98df20e88265994c69e8c89d0363f33e3aec2102609e71e1a5bc5644145347d708f06614c3074224889e161414526ae81638161021036b86247cd493bdc28b4fabd6d8a5f7df4c0c5d478d99b461c47427c68addf14b21021885e29507238e435235e8797207e87fcbb510d2b799ec44ecac239bbb4dc2162102899c05efa42d7d6b438aa09673756921924cad7e34175a3788d57021908de73c55ae",
                                                        "type": "Data (173 Bytes)"
                                                }
                                        ],
                                        "hex": "004cad00210384945b5a9ab287e5f09ea64f0c2e98df20e88265994c69e8c89d0363f33e3aec2102609e71e1a5bc5644145347d708f06614c3074224889e161414526ae81638161021036b86247cd493bdc28b4fabd6d8a5f7df4c0c5d478d99b461c47427c68addf14b21021885e29507238e435235e8797207e87fcbb510d2b799ec44ecac239bbb4dc2162102899c05efa42d7d6b438aa09673756921924cad7e34175a3788d57021908de73c55ae",
                                        "parse_error": false
                                },
                                "previous_output_index": 0,
                                "previous_output_tx_id": "5edc0067f851af34e113a02cf44b8eaa422557545246c06306e9a3955519b7a0",
                                "sequence": 4294967295
                        }
                ],
                "locktime": 0,
                "outputs": [
                        {
                                "address": "1NY2Bb9hjWB7QCrQkaddpvCWgQYfrsxVhb",
                                "output_script": {
                                        "fields": [
                                                {
                                                        "hex": "OP_DUP",
                                                        "type": "OP_DUP"
                                                },
                                                {
                                                        "hex": "OP_HASH160",
                                                        "type": "OP_HASH160"
                                                },
                                                {
                                                        "hex": "ec3885ff8dba56dc8a8d1d78fa1e02288bd35b03",
                                                        "type": "Public Key Hash"
                                                },
                                                {
                                                        "hex": "OP_EQUALVERIFY",
                                                        "type": "OP_EQUALVERIFY"
                                                },
                                                {
                                                        "hex": "OP_CHECKSIG",
                                                        "type": "OP_CHECKSIG"
                                                }
                                        ],
                                        "hex": "76a914ec3885ff8dba56dc8a8d1d78fa1e02288bd35b0388ac",
                                        "parse_error": false
                                },
                                "output_type": "P2PKH",
                                "value": 10000
                        },
                        {
                                "address": "32cEdZHgVthrB9Fbeahfy5zDjFZg9UdvNw",
                                "output_script": {
                                        "fields": [
                                                {
                                                        "hex": "OP_HASH160",
                                                        "type": "OP_HASH160"
                                                },
                                                {
                                                        "hex": "0a10a742523c94086e3d92d76bb23feb28f02b8f",
                                                        "type": "Script Hash"
                                                },
                                                {
                                                        "hex": "OP_EQUAL",
                                                        "type": "OP_EQUAL"
                                                }
                                        ],
                                        "hex": "a9140a10a742523c94086e3d92d76bb23feb28f02b8f87",
                                        "parse_error": false
                                },
                                "output_type": "P2SH",
                                "value": 19919
                        }
                ],
                "version": 2
        }

## Using include_input_detail = true

TxRequest

We are requesting the same transaction as above, but with input detail this time.

        {
                "id": "bbe9e2fced55a2ac4fabb6c74cef8d9dda6cc1121a9782ee7e34fe97e32958cb",
                "options": {
                        "include_input_detail": true
                        "human_readable": true
                }
        }

        $ curl -X POST -d '{"id":"bbe9e2fced55a2ac4fabb6c74cef8d9dda6cc1121a9782ee7e34fe97e32958cb","options":{"include_input_detail":true,"human_readable":true}}' http://127.0.0.1:8888/rest/v1/tx

TxResponse

This response includes the input's previous output detail, spend type and all serialized script information.
From this we can see that the input is not ordinary at all. The redeem script is one of approximately five existing 0-of-5 multisig serialized scripts in the blockchain.
No signatures are required to be verfied in order for the script to succeed. The script will always succeed because it does essentially nothing.
However, the transaction itself is still relatively secure because, in order to steal the funds, someone would have to guess the exact script that produced the script hash.

        {
                "bip141": false,
                "blockhash": "0000000000000000001e7b7b2bf28d3e3813e0a37354f95288fe05fa18a0bc8a",
                "blocktime": 1540380059,
                "coinbase": false,
                "id": "bbe9e2fced55a2ac4fabb6c74cef8d9dda6cc1121a9782ee7e34fe97e32958cb",
                "inputs": [
                        {
                                "coinbase": false,
                                "input_script": {
                                        "fields": [
                                                {
                                                        "hex": "OP_0",
                                                        "type": "OP_0"
                                                },
                                                {
                                                        "hex": "00210384945b5a9ab287e5f09ea64f0c2e98df20e88265994c69e8c89d0363f33e3aec2102609e71e1a5bc5644145347d708f06614c3074224889e161414526ae81638161021036b86247cd493bdc28b4fabd6d8a5f7df4c0c5d478d99b461c47427c68addf14b21021885e29507238e435235e8797207e87fcbb510d2b799ec44ecac239bbb4dc2162102899c05efa42d7d6b438aa09673756921924cad7e34175a3788d57021908de73c55ae",
                                                        "type": "SERIALIZED REDEEM SCRIPT"
                                                }
                                        ],
                                        "hex": "004cad00210384945b5a9ab287e5f09ea64f0c2e98df20e88265994c69e8c89d0363f33e3aec2102609e71e1a5bc5644145347d708f06614c3074224889e161414526ae81638161021036b86247cd493bdc28b4fabd6d8a5f7df4c0c5d478d99b461c47427c68addf14b21021885e29507238e435235e8797207e87fcbb510d2b799ec44ecac239bbb4dc2162102899c05efa42d7d6b438aa09673756921924cad7e34175a3788d57021908de73c55ae",
                                        "parse_error": false
                                },
                                "previous_output": {
                                        "address": "3KEuii76sHXXtJNDKHvXNocEm3shTE4LYu",
                                        "output_script": {
                                                "fields": [
                                                        {
                                                                "hex": "OP_HASH160",
                                                                "type": "OP_HASH160"
                                                        },
                                                        {
                                                                "hex": "c082452ddc5bdc870feabcabff778f1fb2177396",
                                                                "type": "Script Hash"
                                                        },
                                                        {
                                                                "hex": "OP_EQUAL",
                                                                "type": "OP_EQUAL"
                                                        }
                                                ],
                                                "hex": "a914c082452ddc5bdc870feabcabff778f1fb217739687",
                                                "parse_error": false
                                        },
                                        "output_type": "P2SH",
                                        "value": 30699
                                },
                                "previous_output_index": 0,
                                "previous_output_tx_id": "5edc0067f851af34e113a02cf44b8eaa422557545246c06306e9a3955519b7a0",
                                "redeem_script": {
                                        "fields": [
                                                {
                                                        "hex": "OP_0",
                                                        "type": "OP_0"
                                                },
                                                {
                                                        "hex": "0384945b5a9ab287e5f09ea64f0c2e98df20e88265994c69e8c89d0363f33e3aec",
                                                        "type": "Public Key"
                                                },
                                                {
                                                        "hex": "02609e71e1a5bc5644145347d708f06614c3074224889e161414526ae816381610",
                                                        "type": "Public Key"
                                                },
                                                {
                                                        "hex": "036b86247cd493bdc28b4fabd6d8a5f7df4c0c5d478d99b461c47427c68addf14b",
                                                        "type": "Public Key"
                                                },
                                                {
                                                        "hex": "021885e29507238e435235e8797207e87fcbb510d2b799ec44ecac239bbb4dc216",
                                                        "type": "Public Key"
                                                },
                                                {
                                                        "hex": "02899c05efa42d7d6b438aa09673756921924cad7e34175a3788d57021908de73c",
                                                        "type": "Public Key"
                                                },
                                                {
                                                        "hex": "OP_5",
                                                        "type": "OP_5"
                                                },
                                                {
                                                        "hex": "OP_CHECKMULTISIG",
                                                        "type": "OP_CHECKMULTISIG"
                                                }
                                        ],
                                        "hex": "00210384945b5a9ab287e5f09ea64f0c2e98df20e88265994c69e8c89d0363f33e3aec2102609e71e1a5bc5644145347d708f06614c3074224889e161414526ae81638161021036b86247cd493bdc28b4fabd6d8a5f7df4c0c5d478d99b461c47427c68addf14b21021885e29507238e435235e8797207e87fcbb510d2b799ec44ecac239bbb4dc2162102899c05efa42d7d6b438aa09673756921924cad7e34175a3788d57021908de73c55ae",
                                        "parse_error": false
                                },
                                "sequence": 4294967295,
                                "spend_type": "P2SH"
                        }
                ],
                "locktime": 0,
                "outputs": [
                        {
                                "address": "1NY2Bb9hjWB7QCrQkaddpvCWgQYfrsxVhb",
                                "output_script": {
                                        "fields": [
                                                {
                                                        "hex": "OP_DUP",
                                                        "type": "OP_DUP"
                                                },
                                                {
                                                        "hex": "OP_HASH160",
                                                        "type": "OP_HASH160"
                                                },
                                                {
                                                        "hex": "ec3885ff8dba56dc8a8d1d78fa1e02288bd35b03",
                                                        "type": "Public Key Hash"
                                                },
                                                {
                                                        "hex": "OP_EQUALVERIFY",
                                                        "type": "OP_EQUALVERIFY"
                                                },
                                                {
                                                        "hex": "OP_CHECKSIG",
                                                        "type": "OP_CHECKSIG"
                                                }
                                        ],
                                        "hex": "76a914ec3885ff8dba56dc8a8d1d78fa1e02288bd35b0388ac",
                                        "parse_error": false
                                },
                                "output_type": "P2PKH",
                                "value": 10000
                        },
                        {
                                "address": "32cEdZHgVthrB9Fbeahfy5zDjFZg9UdvNw",
                                "output_script": {
                                        "fields": [
                                                {
                                                        "hex": "OP_HASH160",
                                                        "type": "OP_HASH160"
                                                },
                                                {
                                                        "hex": "0a10a742523c94086e3d92d76bb23feb28f02b8f",
                                                        "type": "Script Hash"
                                                },
                                                {
                                                        "hex": "OP_EQUAL",
                                                        "type": "OP_EQUAL"
                                                }
                                        ],
                                        "hex": "a9140a10a742523c94086e3d92d76bb23feb28f02b8f87",
                                        "parse_error": false
                                },
                                "output_type": "P2SH",
                                "value": 19919
                        }
                ],
                "version": 2
        }

