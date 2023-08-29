
# JSON Objects

## BlockRequestOptions

        {
                NoTypes bool
                NoTxs bool
                NoUnknownSpendTypes bool
                ScriptUsageStats bool
                HumanReadable bool
        }

- **NoTypes**: Optional. If true, the response will not include spend types or output types. Default: false
- **NoTxs**: Optional. If true, the response will not include transaction data. Default: false
- **NoUnknownSpendTypes**: Optional. If true, the response will not include unknown spend types. Default: false
- **ScriptUsageStats**: Optional. If true, the response will include script usage stats. Default: false
- **HumanReadable**: Optional. If true, JSON will be returned in a human readable format with newlines and indentations. Default: false

***

## BlockRequest

        {
                hash string
                height uint32
                options BlockRequestOptions
        }

- **hash, height**: Optional. The block hash or height. If both are provided in the request, hash will be used and height will be ignored. Default: most recent block
- **options**: Optional. If not included in the request, default values will be used for all options.

***

## BlockTx

Used in the BlockResponse. This is a small summary of information about a transaction in a block.

        {
                Index uint16
                TxId string
                Bip141 bool
                InputCount uint16
                OutputCount uint16
        }

- **Index**: The index of the transaction within the block.
- **TxId**: The transaction id.
- **Bip141**: True if the transaction supports BIP 141. False otherwise.
- **InputCount**: The number of inputs in the transaction.
- **OutputCount**: The number of outputs in the transaction.

***

## SpendTypeList

If spend types are needed for a block, these objects will be returned in the response and should be sent back in subsequent output_types requests. (See example below.)

        {
                string: [ uint32 ]
        }

- **Key**: Transaction id.
- **Value**: Array of output indexes.

***

## TxPartType

Each field is a spend type or output type that points to the total number of objects of that type.

        {
                string: uint16
        }

- **Key**: Spend type name.
- **Value**: Number of inputs of this spend type found in the search results.

***

## BlockResponse

        {
                PreviousHash string
                NextHash string
                Hash string
                Height uint32
                Timestamp int64
                InputCount uint16
                OutputCount uint16

                Bip141Count uint16
                Txs [ BlockTx ]
                UnknownSpendTypes SpendTypeList
                KnownSpendTypes TxPartType
                OutputTypes TxPartType

                RedeemScriptMultisigCount uint16
                RedeemScriptCount uint16

                WitnessScriptMultisigCount uint16
                WitnessScriptCount uint16

                TapScriptOrdinalCount uint16
                TapScriptCount uint16
        }


- **PreviousHash**: Hash of the previous block, if one exists.
- **NextHash**: Hash of the next block, if one exists.
- **Hash**: Hash of the requested block.
- **Height**: Height of the requested block.
- **Timestamp**: Timestamp of the requested block.
- **InputCount**: Number of inputs in the requested block.
- **OutputCount**: Number of outputs in the requested block.
- **Bip141Count**: Number of transactions in the requested block that support BIP 141.
- **Txs**: Array of BlockTx objects, one per transaction in the requested block. If the NoTxs options is set to true, this array will not be included in the response.
- **UnknownSpendTypes**: Transaction ids of previous outputs and an array of output indexes for each. If the NoUnknownSpendTypes option is set to true, this object will not be included in the response.
- **KnownSpendTypes**: Name of each spend time and the number of times it occurs in the block. If the NoTypes option is set to true, this object will not be included in the response.
- **OutputTypes (unless NoTypes)**: Name of each output time and the number of times it occurs in the block. If the NoTypes option is set to true, this object will not be included in the response.
- **RedeemScriptMultisigCount**: Number of standard multisig redeem scripts in the block. If none exist, or if the ScriptUsageStats is set to false, this field will not be included in the response.
- **RedeemScriptCount**: Number of redeem scripts in the block. If none exist, or if the ScriptUsageStats is set to false, this field will not be included in the response.
- **WitnessScriptMultisigCount**: Number of standard multisig witness scripts in the block. If none exist, or if the ScriptUsageStats is set to false, this field will not be included in the response.
- **WitnessScriptCount**: Number of witness scripts in the block. If none exist, or if the ScriptUsageStats is set to false, this field will not be included in the response.
- **TapScriptOrdinalCount**: Number of ordinal tap scripts in the block. If none exist, or if the ScriptUsageStats is set to false, this field will not be included in the response.
- **TapScriptCount**: Number of tap scripts in the block. If none exist, or if the ScriptUsageStats is set to false, this field will not be included in the response.

***

# Examples

Request the most recent block using all default values.

        $ curl -X POST -d '{}' http://127.0.0.1:8080/rest/v1/block

Request the most recent block in human readable JSON without returning unknown spend type data.

        $ curl -X POST -d '{"options":{"HumanReadable":true,"NoUnknownSpendTypes":true}}' http://127.0.0.1:8080/rest/v1/block

Request a block in human readable JSON by block hash.

        $ curl -X POST -d '{"hash":"00000000000000000005956ad0afdcba175f9be14e9fee92282c1a8a66b9a594","options":{"HumanReadable":true}}' http://127.0.0.1:8080/rest/v1/block

Request a block without returning any types. By using this option, we are asking not to receive spend types or output types, only block and transaction data.

        $ curl -X POST -d '{"height":769895,"options":{"NoTypes":true,"HumanReadable":true}}' http://192.168.1.77:8080/rest/v1/block
        {
                "Bip141Count": 3,
                "Hash": "0000000000000000000137ced007fddf254c01c8771f2e8591db63b3cd531b2e",
                "Height": 769895,
                "InputCount": 7,
                "NextHash": "000000000000000000075fce527f1d1c500c7d13406325a80092ad79c65dc53c",
                "OutputCount": 10,
                "PreviousHash": "000000000000000000035acb4404247eafdf7d4ead9e2412ea75d71e55fedb11",
                "Timestamp": 1672591723,
                "Txs": [
                        {
                                "Index": 0,
                                "TxId": "826ec483b6daad9a941680b870f6b0546087f49aace234f821ad3eabeb7cf738",
                                "Bip141": true,
                                "InputCount": 1,
                                "OutputCount": 3
                        },
                        {
                                "Index": 1,
                                "TxId": "55e7d863b62569d0b180a6d36511a49cb46b11470c077c962d67294d4ccfbddf",
                                "Bip141": true,
                                "InputCount": 2,
                                "OutputCount": 1
                        },
                        {
                                "Index": 2,
                                "TxId": "244eb1b6205a0b76f32169f80ea872019f92d6288586c2e5e62fdd4a817fd8d0",
                                "Bip141": false,
                                "InputCount": 1,
                                "OutputCount": 4
                        },
                        {
                                "Index": 3,
                                "TxId": "929c2d4f780a8ed935c6b0d71f48458bb6bdefb8c35e4b43ee4bbc51d1194f40",
                                "Bip141": true,
                                "InputCount": 3,
                                "OutputCount": 2
                        }
                ]
        }

***


Request the most recent block in human readable JSON with only script usage stats but no type data or transaction data.

        $ curl -X POST -d '{"options":{"NoTxs":true,"NoTypes":true,"ScriptUsageStats":true,"HumanReadable":true}}' http://127.0.0.1:8080/rest/v1/block
        {
                "Hash": "00000000000000000003db6f7b94a5bebcf0db944e60ec5235c23dfaa220840c",
                "Height": 805200,
                "InputCount": 7266,
                "OutputCount": 9758,
                "PreviousHash": "0000000000000000000040776f7a64838a3ec78c4326b85a7c3e7a1ea8e80f9e",
                "RedeemScriptCount": 65,
                "RedeemScriptMultisigCount": 62,
                "TapScriptCount": 50,
                "TapScriptOrdinalCount": 45,
                "Timestamp": 1693252818,
                "WitnessScriptCount": 756,
                "WitnessScriptMultisigCount": 701
        }

In order to analyze spend types in blocks, a multi-step process is required in order to gather the data.
This is because the spend types of legacy non-segwit inputs can not be known without looking at every previous output.
In order to save time, and throttle the requests to your node, the previous outputs are obtained separately.

**Step 1**: Get the block.

        $ curl -X POST -d '{"height":752522,"options":{"HumanReadable":true}}' http://127.0.0.1:8080/rest/v1/block
        {
                "Bip141Count": 2,
                "Hash": "000000000000000000087b79d6441b48bd5d287673dc4bf2ee0b0ecf4103cbf0",
                "Height": 752522,
                "InputCount": 11,
                "KnownSpendTypes": {
                        "P2WPKH": 1
                },
                "NextHash": "00000000000000000008fa3e9f787a2c5e240d582dcfd77d88dbc4a949b84988",
                "OutputCount": 11,
                "OutputTypes": {
                        "OP_RETURN": 4,
                        "P2PKH": 2,
                        "P2SH": 1,
                        "P2WPKH": 4
                },
                "PreviousHash": "00000000000000000000df0fae230bdf9066a21723bdfc7fa6f8702b15ad3fe9",
                "Timestamp": 1662264017,
                "Txs": [
                        {
                                "Index": 0,
                                "TxId": "88108ebfb5f34b9200c6328fe88a66f6b0247dbaec22f26b3fed12b743a7ee12",
                                "Bip141": true,
                                "InputCount": 1,
                                "OutputCount": 5
                        },
                        {
                                "Index": 1,
                                "TxId": "3f8208fe1fab400344b193588f1d8490d05d1b058aad5d91f8e5f78ed6fbb0c1",
                                "Bip141": false,
                                "InputCount": 2,
                                "OutputCount": 1
                        },
                        {
                                "Index": 2,
                                "TxId": "b46f04f7616c0613ede7b19c567a7547ee14ae53da1171b415405999bff54f77",
                                "Bip141": true,
                                "InputCount": 1,
                                "OutputCount": 2
                        },
                        {
                                "Index": 3,
                                "TxId": "42097f3acb7f79a51ed8380d32763d8e8e64470c8ff7ae302708f82b1706547d",
                                "Bip141": false,
                                "InputCount": 6,
                                "OutputCount": 1
                        },
                        {
                                "Index": 4,
                                "TxId": "e869ffcf61e961dd28e856620b5a499a229c8a564c14bed5a394cb39cb48c41c",
                                "Bip141": false,
                                "InputCount": 1,
                                "OutputCount": 2
                        }
                ],
                "UnknownSpendTypes": {
                        "031918f9778991941e5a03cf6042a1c4ad5c1987d96cad512cdbefed1e35a900": [
                                0
                        ],
                        "05db92ec996950dd9d4344b9006cbd0d96c5d228ac3c32754034a73998a55bd7": [
                                1
                        ],
                        "14897f4eb049a47296bad20f53f7da63bd500b8ee3d86e80fcf298ac81324c66": [
                                189
                        ],
                        "292c0b21b6fd8ec5d332f76fb6bf17e3f84336b3225ff4a780ee5ef4f76c26d4": [
                                152
                        ],
                        "72b329e94dab613c6dd82c670a7deb8c00c406af0a776b3ed70e92b4be9760ba": [
                                88,
                                143
                        ],
                        "7e9b26fef1afa0523a420d4747077b8d9d44defd118d9d6447a1e56ecdc0dd05": [
                                122
                        ],
                        "8b5d1c025d50f254352203facf2d893547ec2b5ee786dcaf7fb8e9f9ce8222a0": [
                                1
                        ],
                        "a28bbb62247f9c8b4a00baad467900df7f2e78dd763128c40c1d82e5c2c69fd7": [
                                121
                        ]
                }
        }

In this example, only one of the input spend types was known in the original response. The rest must be retrived separately.
There are 8 separate transactions to get previous output data from.
The spend types can be retrieved in batches or individually.

**Step 2**: Request 2 groups of 4 output types.

		$ curl -X POST -d '{"031918f9778991941e5a03cf6042a1c4ad5c1987d96cad512cdbefed1e35a900":[0],"05db92ec996950dd9d4344b9006cbd0d96c5d228ac3c32754034a73998a55bd7":[1],"14897f4eb049a47296bad20f53f7da63bd500b8ee3d86e80fcf298ac81324c66":[189],"292c0b21b6fd8ec5d332f76fb6bf17e3f84336b3225ff4a780ee5ef4f76c26d4":[152]}' http://127.0.0.1:8080/rest/v1/output_types
        {"031918f9778991941e5a03cf6042a1c4ad5c1987d96cad512cdbefed1e35a900:0":"P2PKH","05db92ec996950dd9d4344b9006cbd0d96c5d228ac3c32754034a73998a55bd7:1":"P2PKH","14897f4eb049a47296bad20f53f7da63bd500b8ee3d86e80fcf298ac81324c66:189":"P2PKH","292c0b21b6fd8ec5d332f76fb6bf17e3f84336b3225ff4a780ee5ef4f76c26d4:152":"P2PKH"}

		$ curl -X POST -d '{"72b329e94dab613c6dd82c670a7deb8c00c406af0a776b3ed70e92b4be9760ba":[88,143],"7e9b26fef1afa0523a420d4747077b8d9d44defd118d9d6447a1e56ecdc0dd05":[122],"8b5d1c025d50f254352203facf2d893547ec2b5ee786dcaf7fb8e9f9ce8222a0":[1],"a28bbb62247f9c8b4a00baad467900df7f2e78dd763128c40c1d82e5c2c69fd7":[121]}' http://127.0.0.1:8080/rest/v1/output_types
        {"72b329e94dab613c6dd82c670a7deb8c00c406af0a776b3ed70e92b4be9760ba:143":"P2PKH","72b329e94dab613c6dd82c670a7deb8c00c406af0a776b3ed70e92b4be9760ba:88":"P2PKH","7e9b26fef1afa0523a420d4747077b8d9d44defd118d9d6447a1e56ecdc0dd05:122":"P2PKH","8b5d1c025d50f254352203facf2d893547ec2b5ee786dcaf7fb8e9f9ce8222a0:1":"P2PKH","a28bbb62247f9c8b4a00baad467900df7f2e78dd763128c40c1d82e5c2c69fd7:121":"P2PKH"}

We now have a full analysis of spend types for the block. It has a total of 10 non-coinbase inputs: 1 P2WPKH, 9 P2PKH.

