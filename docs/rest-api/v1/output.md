# JSON Request Objects

## OutputOptions

Name | Type | Required | Default | Description
:---:|:---:|:---:|:---:|:---:
human_readable | bool | No | false | return human readable JSON

## OutputRequest

Name | Type | Required | Default | Description
:---:|:---:|:---:|:---:|:---:
tx_id | string | Yes | | transaction id
output_index | uint16 | Yes | | output index
options | OutputOptions | No | not included | options

# Examples

## A Taproot Output

OutputRequest

        {
                "tx_id": "7641c08f4bd299abfef26dcc6b477938f4a6c2eed2f224d1f5c1c86b4e09739d",
                "output_index": 1,
                "options": {
                        "human_readable": true
                }
        }

        $ curl -X POST -d '{"tx_id":"7641c08f4bd299abfef26dcc6b477938f4a6c2eed2f224d1f5c1c86b4e09739d","output_index":1,"options":{"human_readable":true}}' http://127.0.0.1:8080/rest/v1/output

Output response

        {
                "address": "bc1pmfr3p9j00pfxjh0zmgp99y8zftmd3s5pmedqhyptwy6lm87hf5sspknck9",
                "output_script": {
                        "fields": [
                                {
                                        "hex": "OP_1",
                                        "type": "OP_1"
                                },
                                {
                                        "hex": "da4710964f7852695de2da025290e24af6d8c281de5a0b902b7135fd9fd74d21",
                                        "type": "Witness Program (Public Key)"
                                }
                        ],
                        "hex": "5120da4710964f7852695de2da025290e24af6d8c281de5a0b902b7135fd9fd74d21",
                        "parse_error": false
                },
                "output_type": "Taproot",
                "value": 50000
        }

## A Legacy Multisig Output

OutputRequest

        {
                "tx_id": "b3e4d204d3a9b139789f9da9f6efd546b9d67b445d65231f4842133c4e30a41c",
                "output_index": 2,
                "options": {
                        "human_readable": true
                }
        }

        $ curl -X POST -d '{"tx_id":"b3e4d204d3a9b139789f9da9f6efd546b9d67b445d65231f4842133c4e30a41c","output_index":2,"options":{"human_readable":true}}' http://127.0.0.1:8080/rest/v1/output

Output response

        {
                "output_script": {
                        "fields": [
                                {
                                        "hex": "OP_1",
                                        "type": "OP_1"
                                },
                                {
                                        "hex": "02e992d2ea5c50a05dfb511bd2a28f861b16f885276b806a3e21a77b1df472c193",
                                        "type": "Public Key"
                                },
                                {
                                        "hex": "03b505bca806b5c4c31dc0ad205ad46b9ed09f03e0c5fd44e568fc3c031e048c86",
                                        "type": "Public Key"
                                },
                                {
                                        "hex": "02e5d464c868969dad8a193adf125b715601f8b6a7271e994cb6e6806abf2a0bf0",
                                        "type": "Public Key"
                                },
                                {
                                        "hex": "OP_3",
                                        "type": "OP_3"
                                },
                                {
                                        "hex": "OP_CHECKMULTISIG",
                                        "type": "OP_CHECKMULTISIG"
                                }
                        ],
                        "hex": "512102e992d2ea5c50a05dfb511bd2a28f861b16f885276b806a3e21a77b1df472c1932103b505bca806b5c4c31dc0ad205ad46b9ed09f03e0c5fd44e568fc3c031e048c862102e5d464c868969dad8a193adf125b715601f8b6a7271e994cb6e6806abf2a0bf053ae",
                        "parse_error": false
                },
                "output_type": "MultiSig",
                "value": 7800
        }

