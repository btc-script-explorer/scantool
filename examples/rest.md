### Send a Request for a Transaction data for a Block

We are not asking for spend types or output types, only block and transaction data.

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

Same block as above, but we are not asking for transaction data, only spend types and output types.
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

### Send a Request for legacy Spend Types for a set of Previous Output Transactions

        $ curl -X POST -d '{"98c7ada0b949831d8c2d1e11894d74dee7f098f117cbea626e6ba528e216462b":[2]}' http://192.168.1.77:8080/rest/v1/previous_output_types
        {"98c7ada0b949831d8c2d1e11894d74dee7f098f117cbea626e6ba528e216462b:2":"P2PKH"}

***

### Send a Request for legacy Spend Types for a set of Previous Output Transactions

Multiple previous output transactions can be requested and each one can include multiple output indexes.

        $ curl -X POST -d '{"e4c6fd516460d1601c9374cead9b22ef8be5cdd35393af663752e1f4dfd59776":[1,3],"6b527545583c0afb4ed7d159923e225df5eae724dc2b906f3f58b49dc083c6f3":[3,6,10,19]}' http://192.168.1.77:8080/rest/v1/previous_output_types
        {"6b527545583c0afb4ed7d159923e225df5eae724dc2b906f3f58b49dc083c6f3:10":"P2PKH","6b527545583c0afb4ed7d159923e225df5eae724dc2b906f3f58b49dc083c6f3:19":"P2PKH","6b527545583c0afb4ed7d159923e225df5eae724dc2b906f3f58b49dc083c6f3:3":"P2PKH","6b527545583c0afb4ed7d159923e225df5eae724dc2b906f3f58b49dc083c6f3:6":"P2PKH","e4c6fd516460d1601c9374cead9b22ef8be5cdd35393af663752e1f4dfd59776:1":"P2PKH","e4c6fd516460d1601c9374cead9b22ef8be5cdd35393af663752e1f4dfd59776:3":"P2PKH"}

***

