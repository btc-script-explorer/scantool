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

        $ curl -X POST -d '{"InputTxId":"801906494bfa5710e3a47131640859222abf52391de5800844a79fd148d5a658","InputIndex":0,"PrevOutTxId":"f742b911259dd11278e9e3d34f2538c7d77837daef15fc00a047e3f13253aa0b","PrevOutIndex":24}' http://127.0.0.1:8888/rest/v1/previous_output
        {"InputTxId":"801906494bfa5710e3a47131640859222abf52391de5800844a79fd148d5a658","InputIndex":0,"PrevOut":{"Value":769440,"OutputType":"P2PKH","Address":"12iFKzb55TnNURcqSpp3swtZKUTyV2nXxV","OutputScript":{"Fields":[{"Hex":"OP_DUP","Type":"OP_DUP"},{"Hex":"OP_HASH160","Type":"OP_HASH160"},{"Hex":"12c523e2edf0e0de04094f4df37ed2b4f5b26e37","Type":"Public Key Hash"},{"Hex":"OP_EQUALVERIFY","Type":"OP_EQUALVERIFY"},{"Hex":"OP_CHECKSIG","Type":"OP_CHECKSIG"}]}}}

***

