# JSON Objects

## Field

        {
                Hex string
                Type string
        }

- **Hex**: The hex representation of the field.
- **Type**: The data type of the field.

***

## Script

        {
                Fields [ Field ]
                Ordinal bool
                Multisig bool
        }

- **Fields**: Array of fields.
- **Ordinal**: Only present for tap scripts. If true, the script is an ordinal.
- **Multisig**: Only present for redeem scripts and witness scripts. If true, the script is a standard multisig script.

***

## Input

        {
                InputIndex uint16
                Coinbase bool
                SpendType string
                Sequence uint32

                InputScript Script
                RedeemScript Script

                PreviousOutputTxId string
                PreviousOutputIndex uint16

                Segwit SegregatedWitness
        }

- **InputIndex**: Index of the input in the transaction.
- **Coinbase**: If true, this is a coinbase input.
- **InputScript**: Input script.
- **RedeemScript**: Only present if there is a redeem script.
- **SpendType**: Spend type of this input.
- **Sequence**: Sequence number for this input.
- **PreviousOutputTxId**: Transaction id for the previous output.
- **PreviousOutputIndex**: Output index for the previous output.
- **Segwit**: Only present for transactions that support BIP 141.

***

## Output

        {
                OutputIndex uint16
                OutputType string
                Value uint64
                Address string
                OutputScript Script
        }

- **OutputIndex**: Not present for previous outputs. Index of the output in the transaction.
- **OutputType**: Output type of this output.
- **Value**: Value of this output.
- **Address**: Address for this output.
- **OutputScript**: Output script.

***

## SegregatedWitness

        {
                Fields [ Field ]
                WitnessScript Script
                TapScript Script
        }

- **Fields**: Array of fields in the segregated witness.
- **WitnessScript**: Only present if there is a witness script.
- **TapScript**: Only present if there is a tap script.

***

## TxRequestOptions

        {
                HumanReadable bool
        }

- **HumanReadable**: Optional. If true, JSON will be returned in a human readable format with newlines and indentations. Default: false

***

## TxRequest

        {
                id string
                options TxRequestOptions
        }

- **id**: Required. The id of the transaction being requested.
- **options**: Optional. If not included in the request, default values will be used for all options.

***

## TxResponse

        {
                BlockHash string
                BlockHeight uint32
                BlockTime int64

                Id string
                SupportsBip141 bool
                IsCoinbase bool
                Inputs [ Input ]
                Outputs [ Output ]
                LockTime uint32

                PreviousOutputRequests [ PreviousOutputRequest ]
        }

- **BlockHash**: Hash of block that the transaction is in.
- **BlockHeight**: Height of block that the transaction is in.
- **BlockTime**: Time of block that the transaction is in.
- **Id**: Transaction id.
- **SupportsBip141**: True if the transaction supports BIP 141.
- **IsCoinbase**: True if the transaction is the coinbase transaction.
- **Inputs**: Array of inputs.
- **Outputs**: Array of outputs.
- **LockTime**: Lock time of the transaction.
- **PreviousOutputRequests**: Previous output request objects.

***

# Examples

## Previous Output Data

In order to get previous output data for all inputs in a transaction, a multi-step process is required.
In order to save time, and prevent too many requests being sent to your node all at once, the previous outputs are obtained separately.

**Step 1**: Get the transaction.


        $ curl -X POST -d '{"id":"abdcfbd8b77c4b14372c58cd7bfc5a09ad5c04759c6699b0eaa19e9226746571","options":{"HumanReadable":true}}' http://127.0.0.1:8888/rest/v1/tx
        {
                "BlockHash": "00000000000000000000f3d662ff2ad11674c542405a04052ad722be24f59821",
                "BlockHeight": 804677,
                "BlockTime": 1692893834,
                "Id": "abdcfbd8b77c4b14372c58cd7bfc5a09ad5c04759c6699b0eaa19e9226746571",
                "Inputs": [
                        {
                                "Coinbase": false,
                                "InputIndex": 0,
                                "InputScript": {
                                        "Fields": []
                                },
                                "PreviousOutputIndex": 0,
                                "PreviousOutputTxId": "76e940241753ffe7b97edbce626df8e94ad3789130a48188046d5aa1f3888668",
                                "Segwit": {
                                        "Fields": [
                                                {
                                                        "Hex": "cb4cb1e920cb77e1931260a27abfc793d0410ee1cd05f56234c5b5d15e7e0445d46ca09e901e76514af4121321be82b102e2c4daff1f948eca6cec07f4a02fd901",
                                                        "Type": "Schnorr Signature"
                                                }
                                        ]
                                },
                                "Sequence": 4294967295,
                                "SpendType": "Taproot Key Path"
                        },
                        {
                                "Coinbase": false,
                                "InputIndex": 1,
                                "InputScript": {
                                        "Fields": []
                                },
                                "PreviousOutputIndex": 0,
                                "PreviousOutputTxId": "ba6ce05c8e646b13b41ae44d23281ddcdbafeb64396b7d87855c233685a1400a",
                                "Segwit": {
                                        "Fields": [
                                                {
                                                        "Hex": "1b9db243c31d0b0473cb486f54d2e7782749d658217788885d5c054797b8a3052ab0b45e61500a9c3e9ac0fb9d3a69b97cdc769a6b51d2743c632c90879b180e01",
                                                        "Type": "Schnorr Signature"
                                                }
                                        ]
                                },
                                "Sequence": 4294967295,
                                "SpendType": "Taproot Key Path"
                        }
                ],
                "IsCoinbase": false,
                "LockTime": 0,
                "Outputs": [
                        {
                                "OutputIndex": 0,
                                "OutputType": "Taproot",
                                "Value": 23074264,
                                "Address": "bc1pl68v3e3pwnr5zdvmmygf8y7wzazrjr68g40cpmnt6an277gjmjlqp9qx7k",
                                "OutputScript": {
                                        "Fields": [
                                                {
                                                        "Hex": "OP_1",
                                                        "Type": "OP_1"
                                                },
                                                {
                                                        "Hex": "fe8ec8e62174c741359bd9109393ce1744390f47455f80ee6bd766af7912dcbe",
                                                        "Type": "Witness Program (Public Key)"
                                                }
                                        ]
                                }
                        },
                        {
                                "OutputIndex": 1,
                                "OutputType": "P2WSH",
                                "Value": 762596,
                                "Address": "bc1qw28a8jur9v5qexx9jmd77hv6m7d4el738c4eaj4fsc2eknutgdmswmyv5n",
                                "OutputScript": {
                                        "Fields": [
                                                {
                                                        "Hex": "OP_0",
                                                        "Type": "OP_0"
                                                },
                                                {
                                                        "Hex": "728fd3cb832b280c98c596dbef5d9adf9b5cffd13e2b9ecaa986159b4f8b4377",
                                                        "Type": "Witness Program (Script Hash)"
                                                }
                                        ]
                                }
                        }
                ],
                "PreviousOutputRequests": [
                        {
                                "InputTxId": "abdcfbd8b77c4b14372c58cd7bfc5a09ad5c04759c6699b0eaa19e9226746571",
                                "InputIndex": 0,
                                "PrevOutTxId": "76e940241753ffe7b97edbce626df8e94ad3789130a48188046d5aa1f3888668",
                                "PrevOutIndex": 0
                        },
                        {
                                "InputTxId": "abdcfbd8b77c4b14372c58cd7bfc5a09ad5c04759c6699b0eaa19e9226746571",
                                "InputIndex": 1,
                                "PrevOutTxId": "ba6ce05c8e646b13b41ae44d23281ddcdbafeb64396b7d87855c233685a1400a",
                                "PrevOutIndex": 0
                        }
                ],
                "SupportsBip141": true
        }

In this example, there is previous output data to obtain for 2 inputs. These objects are returned in the PreviousOutputRequest object of the original response.

**Step 2**: Send the requests for the previous outputs. See the previous_output API for more information.

        $ curl -X POST -d '{"InputTxId":"abdcfbd8b77c4b14372c58cd7bfc5a09ad5c04759c6699b0eaa19e9226746571","InputIndex":0,"PrevOutTxId":"76e940241753ffe7b97edbce626df8e94ad3789130a48188046d5aa1f3888668","PrevOutIndex":0}' http://127.0.0.1:8888/rest/v1/previous_output
        {"InputTxId":"abdcfbd8b77c4b14372c58cd7bfc5a09ad5c04759c6699b0eaa19e9226746571","InputIndex":0,"PrevOut":{"Value":589917,"OutputType":"Taproot","Address":"bc1pvd0gwtvld8tyjc08h3x2zkl2urntzawdr9nwtcalhyhj5vh8sl5qmmudj4","OutputScript":{"Fields":[{"Hex":"OP_1","Type":"OP_1"},{"Hex":"635e872d9f69d64961e7bc4ca15beae0e6b175cd1966e5e3bfb92f2a32e787e8","Type":"Witness Program (Public Key)"}]}}}

        $ curl -X POST -d '{"InputTxId":"abdcfbd8b77c4b14372c58cd7bfc5a09ad5c04759c6699b0eaa19e9226746571","InputIndex":1,"PrevOutTxId":"ba6ce05c8e646b13b41ae44d23281ddcdbafeb64396b7d87855c233685a1400a","PrevOutIndex":0}' http://127.0.0.1:8888/rest/v1/previous_output
        {"InputTxId":"abdcfbd8b77c4b14372c58cd7bfc5a09ad5c04759c6699b0eaa19e9226746571","InputIndex":1,"PrevOut":{"Value":23248802,"OutputType":"Taproot","Address":"bc1pp2767q84l8ytnftxudxvyfs4y9z34r2dqr8ltj59pg6ysvf607qqcwwgdw","OutputScript":{"Fields":[{"Hex":"OP_1","Type":"OP_1"},{"Hex":"0abdaf00f5f9c8b9a566e34cc2261521451a8d4d00cff5ca850a3448313a7f80","Type":"Witness Program (Public Key)"}]}}}

After these responses have been received, we have all the data for all the previous outputs in the transaction.

