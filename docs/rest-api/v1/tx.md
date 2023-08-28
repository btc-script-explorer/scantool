
# Block

## BlockRequestOptionsObject

        {
                NoTypes bool
                NoTxs bool
                NoUnknownSpendTypes bool
                ScriptUsageStats bool
                HumanReadable bool
        }

##### NoTypes

If true, the response will not include spend types or output types.
**Default**: false

##### NoTxs

If true, the response will not include transaction data.
**Default**: false

##### NoUnknownSpendTypes

If true, the response will not include unknown spend types.
**Default**: false

##### ScriptUsageStats

If true, the response will include script usage stats.
**Default**: false

##### HumanReadable

If true, JSON will be returned in a human readable format with newlines and indentations.
**Default**: false

***

## BlockRequestObject

        hash string
        height uint32
        options BlockRequestOptionsObject

#### hash, height

The block hash or height. If both are provided in the request, height will be ignored.
**Default**: the most recent block

#### options

The options object is optional. If it is not included in the request, default values for all options will be used.

***

## BlockTxObject

Used in the BlockResponseObject. This is a small summary of information about a transaction in a block.

        {
                Index uint16
                TxId string
                Bip141 bool
                InputCount uint16
                OutputCount uint16
        }

#### Index

The index of the transaction within the block.

#### TxId

The transaction id.

#### Bip141

True if the transaction supports BIP 141.

#### InputCount

The number of inputs in the transaction.

#### OutputCount

The number of outputs in the transaction.

***

## SpendTypeListObject

Each field is a transaction id that points to an array of output indexes.
If spend types are needed for a block, these objects should be sent back in subsequent previous_output_types requests.

        {
                string: [ uint32 ]
        }


## TxPartType

Each field is a spend type or output type that points to the total number of objects of that type.

        {
                string: uint16
        }

***

## BlockResponseObject

        {
                PreviousHash string
                NextHash string
                Hash string
                Height uint32
                Timestamp int64
                InputCount uint16
                OutputCount uint16

                Bip141Count uint16
                Txs [ BlockTxObject ]
                UnknownSpendTypes SpendTypeListObject
                KnownSpendTypes TxPartType
                OutputTypes TxPartType

                RedeemScriptMultisigCount uint16
                RedeemScriptCount uint16

                WitnessScriptMultisigCount uint16
                WitnessScriptCount uint16

                TapScriptOrdinalCount uint16
                TapScriptCount uint16
        }


#### PreviousHash

Hash of the block before the requested block, if one exists.

#### NextHash

Hash of the block after the requested block, if one exists.

#### Hash

Hash of the requested block.

#### Height

Height of the requested block.

#### Timestamp

Timestamp of the requested block.

#### InputCount

Number of inputs in the requested block.

#### OutputCount

Number of outputs in the requested block.

#### Bip141Count

Number of transactions in the requested block that support BIP 141.

#### Txs

An array of BlockTxObject objects, one per transaction in the requested block.
If the NoTxs options is set to true, this array will not be included in the response.

#### UnknownSpendTypes

The transaction ids of previous outputs and an array of output indexes for each.
If the NoUnknownSpendTypes option is set to true, this object will not be included in the response.

#### KnownSpendTypes

The name of each spend time and the number of times it occurs in the block.
If the NoTypes option is set to true, this object will not be included in the response.

#### OutputTypes (unless NoTypes)

The name of each output time and the number of times it occurs in the block.
If the NoTypes option is set to true, this object will not be included in the response.

#### RedeemScriptMultisigCount

The number of standard multisig redeem scripts in the block.
If none exist, or if the ScriptUsageStats is set to false, this field will not be included in the response.

#### RedeemScriptCount

The total number of redeem scripts in the block.
If none exist, or if the ScriptUsageStats is set to false, this field will not be included in the response.

#### WitnessScriptMultisigCount

The number of standard multisig witness scripts in the block.
If none exist, or if the ScriptUsageStats is set to false, this field will not be included in the response.

#### WitnessScriptCount

The total number of witness scripts in the block.
If none exist, or if the ScriptUsageStats is set to false, this field will not be included in the response.

#### TapScriptOrdinalCount

The number of ordinal tap scripts in the block.
If none exist, or if the ScriptUsageStats is set to false, this field will not be included in the response.

#### TapScriptCount

The total number of tap scripts in the block.
If none exist, or if the ScriptUsageStats is set to false, this field will not be included in the response.





## Examples

Request the most recent block using all default values.

        $ curl -X POST -d '{}' http://127.0.0.1:8080/rest/v1/block

Request the most recent block in human readable JSON without returning unknown spend type data.

        $ curl -X POST -d '{"options":{"HumanReadable":true,"NoUnknownSpendTypes":true}}' http://127.0.0.1:8080/rest/v1/block

Request a block with a block hash instead of a block height.

        $ curl -X POST -d '{"hash":"00000000000000000005956ad0afdcba175f9be14e9fee92282c1a8a66b9a594","options":{"HumanReadable":true}}' http://127.0.0.1:8080/rest/v1/block

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

If you want to analyze spend types in blocks, it is a multi-step process to gather the data.
This is because the spend types of legacy non-segwit inputs can not be known without getting the previous outputs.
In order to save time, and avoid bombarding the node with requests, the previous outputs are obtained separately.
The first step is to get the block.

If this were a real blockchain research project, the HumanReadable option would not be used, but we use it here to make the response easier to read.

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

In this sample block, only one of the input spend types was known. The rest must be retrived separately.
The previous_output_types API call will return only spend types identified by their previous outpoint.
(If you need to be identified by their transaction id and input index, the prevout API call must be used.)

The spend types can be retrieved in batches or individually. There are 8 separate transactions to get previous output data from in this case.
So we will send 2 groups of 4.

		$ curl -X POST -d '{"031918f9778991941e5a03cf6042a1c4ad5c1987d96cad512cdbefed1e35a900":[0],"05db92ec996950dd9d4344b9006cbd0d96c5d228ac3c32754034a73998a55bd7":[1],"14897f4eb049a47296bad20f53f7da63bd500b8ee3d86e80fcf298ac81324c66":[189],"292c0b21b6fd8ec5d332f76fb6bf17e3f84336b3225ff4a780ee5ef4f76c26d4":[152]}' http://127.0.0.1:8080/rest/v1/previous_output_types
        {"031918f9778991941e5a03cf6042a1c4ad5c1987d96cad512cdbefed1e35a900:0":"P2PKH","05db92ec996950dd9d4344b9006cbd0d96c5d228ac3c32754034a73998a55bd7:1":"P2PKH","14897f4eb049a47296bad20f53f7da63bd500b8ee3d86e80fcf298ac81324c66:189":"P2PKH","292c0b21b6fd8ec5d332f76fb6bf17e3f84336b3225ff4a780ee5ef4f76c26d4:152":"P2PKH"}

		$ curl -X POST -d '{"72b329e94dab613c6dd82c670a7deb8c00c406af0a776b3ed70e92b4be9760ba":[88,143],"7e9b26fef1afa0523a420d4747077b8d9d44defd118d9d6447a1e56ecdc0dd05":[122],"8b5d1c025d50f254352203facf2d893547ec2b5ee786dcaf7fb8e9f9ce8222a0":[1],"a28bbb62247f9c8b4a00baad467900df7f2e78dd763128c40c1d82e5c2c69fd7":[121]}' http://127.0.0.1:8080/rest/v1/previous_output_types
        {"72b329e94dab613c6dd82c670a7deb8c00c406af0a776b3ed70e92b4be9760ba:143":"P2PKH","72b329e94dab613c6dd82c670a7deb8c00c406af0a776b3ed70e92b4be9760ba:88":"P2PKH","7e9b26fef1afa0523a420d4747077b8d9d44defd118d9d6447a1e56ecdc0dd05:122":"P2PKH","8b5d1c025d50f254352203facf2d893547ec2b5ee786dcaf7fb8e9f9ce8222a0:1":"P2PKH","a28bbb62247f9c8b4a00baad467900df7f2e78dd763128c40c1d82e5c2c69fd7:121":"P2PKH"}

The result is that the block has 10 non-coinbase inputs, 1 P2WPKH and 9 P2PKH.

-------------------------------------------------------




		request:
		{
			"id": "c3e384db67470346df163a2fa50024d674ef1b3e75aa97ec6534d806c82fee7e",
			"options":
			{
			}
		}
		curl -X POST -d '{"id":"61e26d407c17e8ee33a8b166c78f78c53cdcdc0078ae1f9405e6583cfb90eaf4","options":{"HumanReadable":true}}' http://127.0.0.1:8888/rest/v1/tx

		response:
		{
			"height": 789012,
			"hash": "00000000000000000005956ad0afdcba175f9be14e9fee92282c1a8a66b9a594",
			"previous-hash":
			"next-hash":
			"timestamp":
			"txs":
			[
				{
					"index": 0
					"id": "",
					"bip141": true,
					"input-count": 4444,
					"output-count": 5555
				}
			]
		}






		request:
		curl -X GET http://127.0.0.1:8888/rest/v1/current_block_height

		response:
		{
			"Current_block_height": 802114
		}






		request:
		{
			"f742b911259dd11278e9e3d34f2538c7d77837daef15fc00a047e3f13253aa0b": [0, 24],
			"f76874221ce8d7961f91b9ad5e827fb558d5ce95bc60b2722112d3384069c61b": [17, 21],
			"f7cb6da78444ff5029c5bd1f9ae6c64a1d2d65c604f2989c90cf9dc323692ed8": [0, 2]
		}
		curl -X POST -d "{\"f742b911259dd11278e9e3d34f2538c7d77837daef15fc00a047e3f13253aa0b\":[0,24],\"f76874221ce8d7961f91b9ad5e827fb558d5ce95bc60b2722112d3384069c61b\":[17,21],\"f7cb6da78444ff5029c5bd1f9ae6c64a1d2d65c604f2989c90cf9dc323692ed8\":[0,2]}" http://127.0.0.1:8888/rest/v1/previous_output_types

		response:
		{
			"f742b911259dd11278e9e3d34f2538c7d77837daef15fc00a047e3f13253aa0b:0": "P2PKH",
			"f742b911259dd11278e9e3d34f2538c7d77837daef15fc00a047e3f13253aa0b:24": "P2PKH",
			"f76874221ce8d7961f91b9ad5e827fb558d5ce95bc60b2722112d3384069c61b:17": "P2PKH",
			"f76874221ce8d7961f91b9ad5e827fb558d5ce95bc60b2722112d3384069c61b:21": "P2PKH",
			"f7cb6da78444ff5029c5bd1f9ae6c64a1d2d65c604f2989c90cf9dc323692ed8:0": "P2PKH",
			"f7cb6da78444ff5029c5bd1f9ae6c64a1d2d65c604f2989c90cf9dc323692ed8:2": "P2PKH"
		}







		request:
		{
			"InputTxId":"801906494bfa5710e3a47131640859222abf52391de5800844a79fd148d5a658",
			"InputIndex":0,
			"PrevOutTxId":"f742b911259dd11278e9e3d34f2538c7d77837daef15fc00a047e3f13253aa0b",
			"PrevOutIndex":24
		}
		curl -X POST -d '{"InputTxId":"801906494bfa5710e3a47131640859222abf52391de5800844a79fd148d5a658","InputIndex":0,"PrevOutTxId":"f742b911259dd11278e9e3d34f2538c7d77837daef15fc00a047e3f13253aa0b","PrevOutIndex":24}' http://127.0.0.1:8888/rest/v1/prevout

		response:
		{
			"InputTxId":"801906494bfa5710e3a47131640859222abf52391de5800844a79fd148d5a658",
			"InputIndex":0,
			"PrevOut":
			{
				"Value":769440,
				"OutputType":"P2PKH",
				"Address":"12iFKzb55TnNURcqSpp3swtZKUTyV2nXxV",
				"OutputScript":
				{
					"Fields":[
								{"Hex":"OP_DUP","Type":"OP_DUP"},
								{"Hex":"OP_HASH160","Type":"OP_HASH160"},
								{"Hex":"12c523e2edf0e0de04094f4df37ed2b4f5b26e37","Type":"Public Key Hash"},
								{"Hex":"OP_EQUALVERIFY","Type":"OP_EQUALVERIFY"},
								{"Hex":"OP_CHECKSIG","Type":"OP_CHECKSIG"}
					]
				}
			}
		}








		request:
		curl -X POST http://127.0.0.1:8888/rest/v1/serialized_script_usage
		curl -X POST -d '{"height":801234,"options":{"HumanReadable":true}}' http://127.0.0.1:8888/rest/v1/serialized_script_usage
		curl -X POST -d '{"height":786501,"options":{"HumanReadable":true}}' http://127.0.0.1:8888/rest/v1/serialized_script_usage

		response:


