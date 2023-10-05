# JSON Request Objects

## InputOptions

Name | Type | Required | Default | Description
:---:|:---:|:---:|:---:|:---:
human_readable | bool | No | false | return human readable JSON

## InputRequest

Name | Type | Required | Default | Description
:---:|:---:|:---:|:---:|:---:
tx_id | string | Yes | | transaction id
input_index | uint16 | Yes | | input index
options | InputOptions | No | not included | options

***

## Example 1

A coinbase input. Since these have no previous output, the response does not include as many fields.

InputRequest

        {
                "tx_id": "adb4e7a5115b1073f5850ba88a8ff5bb4e7b6cf667fbc0e111e5ab245f01a14c",
                "input_index": 0,
                "options": {
                        "human_readable": true
                }
        }

        $ curl -X POST -d '{"tx_id":"adb4e7a5115b1073f5850ba88a8ff5bb4e7b6cf667fbc0e111e5ab245f01a14c","input_index":0,"options":{"human_readable":true}}' http://127.0.0.1:8080/rest/v1/input

Input response

        {
                "coinbase": true,
                "input_script": {
                        "fields": [
                                {
                                        "hex": "01bf0b",
                                        "type": "Data (3 Bytes)"
                                },
                                {
                                        "hex": "fabe6d6d57a1bb7bb7085843083e39169ab323f84d990be8e4ad3d847d12b0cf3186778701000000000000000d650800a802a60000000000000000e7b57407042f736c7573682f",
                                        "type": "Data (71 Bytes)"
                                }
                        ],
                        "hex": "0301bf0bfabe6d6d57a1bb7bb7085843083e39169ab323f84d990be8e4ad3d847d12b0cf3186778701000000000000000d650800a802a60000000000000000e7b57407042f736c7573682f",
                        "parse_error": true
                },
                "segwit": {
                        "fields": [
                                {
                                        "hex": "0000000000000000000000000000000000000000000000000000000000000000"
                                }
                        ]
                },
                "sequence": 0
        }

***

## Example 2

A P2SH-P2WSH input. This is the only spend type that includes two serialized scripts. The previous output is P2SH, so the input script contains a redeem script.
It is also a wrapped P2WSH transaction, so the segregated witness contains a witness script.

InputRequest

        {
                "tx_id": "042c4f45e5bd4a0e24262436fcdc48dff83d98ee16a841ada62c2f460572a414",
                "input_index": 4,
                "options": {
                        "human_readable": true
                }
        }

        $ curl -X POST -d '{"tx_id":"042c4f45e5bd4a0e24262436fcdc48dff83d98ee16a841ada62c2f460572a414","input_index":4,"options":{"human_readable":true}}' http://127.0.0.1:8888/rest/v1/input

Input response

        {
                "coinbase": false,
                "input_script": {
                        "fields": [
                                {
                                        "hex": "0020edfa6bc5ecef026a3ea2ba6caae46d0fd36322ce2d6c7fcb06b66f4b15705c1a",
                                        "type": "SERIALIZED REDEEM SCRIPT"
                                }
                        ],
                        "hex": "220020edfa6bc5ecef026a3ea2ba6caae46d0fd36322ce2d6c7fcb06b66f4b15705c1a",
                        "parse_error": false
                },
                "previous_output": {
                        "address": "3McjpTqks39EXHATreiet3fPt2P13Lrf1q",
                        "output_script": {
                                "fields": [
                                        {
                                                "hex": "OP_HASH160",
                                                "type": "OP_HASH160"
                                        },
                                        {
                                                "hex": "da93687f54efc36ddba2bb0e89d1791f9742530d",
                                                "type": "Script Hash"
                                        },
                                        {
                                                "hex": "OP_EQUAL",
                                                "type": "OP_EQUAL"
                                        }
                                ],
                                "hex": "a914da93687f54efc36ddba2bb0e89d1791f9742530d87",
                                "parse_error": false
                        },
                        "output_type": "P2SH",
                        "value": 69604952
                },
                "previous_output_index": 1,
                "previous_output_tx_id": "a27a5ffe03c9f22a3ea138df597c3c1e1a98e8d5e83d4cf7848ad5e96dc2d8fb",
                "redeem_script": {
                        "fields": [
                                {
                                        "hex": "OP_0",
                                        "type": "OP_0"
                                },
                                {
                                        "hex": "edfa6bc5ecef026a3ea2ba6caae46d0fd36322ce2d6c7fcb06b66f4b15705c1a",
                                        "type": "Witness Program (Script Hash)"
                                }
                        ],
                        "hex": "0020edfa6bc5ecef026a3ea2ba6caae46d0fd36322ce2d6c7fcb06b66f4b15705c1a",
                        "parse_error": false
                },
                "segwit": {
                        "fields": [
                                {
                                        "hex": "",
                                        "type": "Data (0 Bytes)"
                                },
                                {
                                        "hex": "304402200942fdab8d46d4663e251e8c8a2eb85613ab34ec2e039d88430e01f6553bf0290220656bee7fa241861e063c4c7927626921165e360bf17731220ebc54a9d4e36d2701",
                                        "type": "Signature"
                                },
                                {
                                        "hex": "3044022075516cedac86c5d2f3c136520f3ee6e28ec7e7585ae0b85c2c96284e58855c210220620e18b5f08bc4f1472f6e969bc389ce6a67c953107ace9256353707b03517a001",
                                        "type": "Signature"
                                },
                                {
                                        "hex": "5221022ec1e2182e1fbae2c95c45d108e139bf4547e61fd5c8460f6f38682158ee587f210381a36e10cc711aedfc2378df0eb9f65e472122bb025b1f2ee7457498bc19fe2a52ae",
                                        "type": "SERIALIZED WITNESS SCRIPT"
                                }
                        ],
                        "witness_script": {
                                "fields": [
                                        {
                                                "hex": "OP_2",
                                                "type": "OP_2"
                                        },
                                        {
                                                "hex": "022ec1e2182e1fbae2c95c45d108e139bf4547e61fd5c8460f6f38682158ee587f",
                                                "type": "Public Key"
                                        },
                                        {
                                                "hex": "0381a36e10cc711aedfc2378df0eb9f65e472122bb025b1f2ee7457498bc19fe2a",
                                                "type": "Public Key"
                                        },
                                        {
                                                "hex": "OP_2",
                                                "type": "OP_2"
                                        },
                                        {
                                                "hex": "OP_CHECKMULTISIG",
                                                "type": "OP_CHECKMULTISIG"
                                        }
                                ],
                                "hex": "5221022ec1e2182e1fbae2c95c45d108e139bf4547e61fd5c8460f6f38682158ee587f210381a36e10cc711aedfc2378df0eb9f65e472122bb025b1f2ee7457498bc19fe2a52ae",
                                "parse_error": false
                        }
                },
                "sequence": 4294967295,
                "spend_type": "P2SH-P2WSH"
        }

