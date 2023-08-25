### Send a Request for a Transaction data for a Block

By using the *NoTypes* option, we are asking not to receive spend types or output types, only block and transaction data.

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

### Send a Request for Spend Type and Output Type data for a Block

By using the *NoTxs* option, we are asking not to receive transaction data, only spend types and output types.
Getting the legacy spend types, requires the extra step of sending a separate request as shown below.

        $ curl -X POST -d '{"height":769895,"options":{"NoTxs":true,"HumanReadable":true}}' http://192.168.1.77:8080/rest/v1/block
        {
                "Hash": "0000000000000000000137ced007fddf254c01c8771f2e8591db63b3cd531b2e",
                "Height": 769895,
                "InputCount": 7,
                "KnownSpendTypes": {
                        "P2WPKH": 5
                },
                "NextHash": "000000000000000000075fce527f1d1c500c7d13406325a80092ad79c65dc53c",
                "OutputCount": 10,
                "OutputTypes": {
                        "OP_RETURN": 2,
                        "P2PKH": 2,
                        "P2SH": 4,
                        "P2WPKH": 2
                },
                "PreviousHash": "000000000000000000035acb4404247eafdf7d4ead9e2412ea75d71e55fedb11",
                "Timestamp": 1672591723,
                "UnknownSpendTypes": {
                        "98c7ada0b949831d8c2d1e11894d74dee7f098f117cbea626e6ba528e216462b": [
                                2
                        ]
                }
        }

***

### Send a Request for the Legacy Spend Type for a Single Previous Output

        $ curl -X POST -d '{"98c7ada0b949831d8c2d1e11894d74dee7f098f117cbea626e6ba528e216462b":[2]}' http://192.168.1.77:8080/rest/v1/previous_output_types
        {"98c7ada0b949831d8c2d1e11894d74dee7f098f117cbea626e6ba528e216462b:2":"P2PKH"}

***

### Send a Request for the Legacy Spend Types for Multiple Previous Outputs from Multiple Transactions

Multiple previous output transactions can be requested and each one can include multiple output indexes.

        $ curl -X POST -d '{"e4c6fd516460d1601c9374cead9b22ef8be5cdd35393af663752e1f4dfd59776":[1,3],"6b527545583c0afb4ed7d159923e225df5eae724dc2b906f3f58b49dc083c6f3":[3,6,10,19]}' http://192.168.1.77:8080/rest/v1/previous_output_types
        {"6b527545583c0afb4ed7d159923e225df5eae724dc2b906f3f58b49dc083c6f3:10":"P2PKH","6b527545583c0afb4ed7d159923e225df5eae724dc2b906f3f58b49dc083c6f3:19":"P2PKH","6b527545583c0afb4ed7d159923e225df5eae724dc2b906f3f58b49dc083c6f3:3":"P2PKH","6b527545583c0afb4ed7d159923e225df5eae724dc2b906f3f58b49dc083c6f3:6":"P2PKH","e4c6fd516460d1601c9374cead9b22ef8be5cdd35393af663752e1f4dfd59776:1":"P2PKH","e4c6fd516460d1601c9374cead9b22ef8be5cdd35393af663752e1f4dfd59776:3":"P2PKH"}

***

### Send a Request for a Transaction

        $ curl -X POST -d '{"id":"abdcfbd8b77c4b14372c58cd7bfc5a09ad5c04759c6699b0eaa19e9226746571","options":{"HumanReadable":true}}' http://192.168.1.77:8080/rest/v1/tx
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

***

### Send a Request for a Single Previous Output with the Input Included in the Response

        $ curl -X POST -d '{"InputTxId":"801906494bfa5710e3a47131640859222abf52391de5800844a79fd148d5a658","InputIndex":0,"PrevOutTxId":"f742b911259dd11278e9e3d34f2538c7d77837daef15fc00a047e3f13253aa0b","PrevOutIndex":24}' http://127.0.0.1:8888/rest/v1/prevout
        {"InputTxId":"801906494bfa5710e3a47131640859222abf52391de5800844a79fd148d5a658","InputIndex":0,"PrevOut":{"Value":769440,"OutputType":"P2PKH","Address":"12iFKzb55TnNURcqSpp3swtZKUTyV2nXxV","OutputScript":{"Fields":[{"Hex":"OP_DUP","Type":"OP_DUP"},{"Hex":"OP_HASH160","Type":"OP_HASH160"},{"Hex":"12c523e2edf0e0de04094f4df37ed2b4f5b26e37","Type":"Public Key Hash"},{"Hex":"OP_EQUALVERIFY","Type":"OP_EQUALVERIFY"},{"Hex":"OP_CHECKSIG","Type":"OP_CHECKSIG"}]}}}

***

