# JSON Request Objects

## InputOptions

Field | Type | Required | Default
---|---|---|---
include_input_detail | bool | No | false
human_readable | bool | No | false

## InputRequest

Field | Type | Required | Default
---|---|---|---
tx_id | string | Yes |
input_index | uint16 | Yes |
options | InputOptions | No | not included

***

# Examples

        {
                "tx_id": "adb4e7a5115b1073f5850ba88a8ff5bb4e7b6cf667fbc0e111e5ab245f01a14c"
                "input_index": 0
                {
                        "human_readable": true
                }
        }

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





        $ curl -X POST -d '{"InputTxId":"abdcfbd8b77c4b14372c58cd7bfc5a09ad5c04759c6699b0eaa19e9226746571","InputIndex":1,"PrevOutTxId":"ba6ce05c8e646b13b41ae44d23281ddcdbafeb64396b7d87855c233685a1400a","PrevOutIndex":0}' http://127.0.0.1:8080/rest/v1/previous_output
        {"InputTxId":"abdcfbd8b77c4b14372c58cd7bfc5a09ad5c04759c6699b0eaa19e9226746571","InputIndex":1,"PrevOut":{"Value":23248802,"OutputType":"Taproot","Address":"bc1pp2767q84l8ytnftxudxvyfs4y9z34r2dqr8ltj59pg6ysvf607qqcwwgdw","OutputScript":{"Fields":[{"Hex":"OP_1","Type":"OP_1"},{"Hex":"0abdaf00f5f9c8b9a566e34cc2261521451a8d4d00cff5ca850a3448313a7f80","Type":"Witness Program (Public Key)"}]}}}

