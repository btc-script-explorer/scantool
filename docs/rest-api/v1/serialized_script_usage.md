# JSON Objects

## SerializedScriptUsageOptions

        {
                redeem bool
                witness bool
                tap bool
                HumanReadable bool
        }

- **redeem**: Optional. Get redeem script usage.
- **witness**: Optional. Get witness script usage.
- **tap**: Optional. Get tap script usage.
- **HumanReadable**: Optional. If true, JSON will be returned in a human readable format with newlines and indentations. Default: false

## SerializedScriptUsageRequest

        {
                height uint32
                options SerializedScriptUsageOptions
        }

- **height**: Required. Block height of the block to examine.
- **options**: Optional.

***

## MultisigResponse

        {
                tx string
                input uint16
                sig_count uint16
                key_count uint16
        }

- **tx**: Id of the transaction where the multisig transaction was discovered.
- **input**: Index of the input where the multisig transaction was discovered.
- **sig_count**: Number of signatures required to be verified.
- **key_count**: Number of public keys provided.

***

## OrdinalResponse

        {
                tx string
                input uint16
                mimetype string
                ord ?
        }

- **tx**: Id of the transaction where the ordinal was discovered.
- **input**: Index of the input where the ordinal was discovered.
- **mimetype**: Mimetype of the ordinal.
- **ord**: Can be a string or an object. If the ordinal contains a JSON object, this field will be that object.

***

# Example

## Multisig Transactions in Redeem Scripts

        $ curl -X POST -d '{"height":805348,"options":{"redeem":true,"HumanReadable":true}}' http://127.0.0.1:8888/rest/v1/serialized_script_usage
        {
                "redeem": [
                        {
                                "input": 0,
                                "key_count": 3,
                                "sig_count": 2,
                                "tx": "931b3138214dd743eb4adfeb1306ec4ca01f2ed5586bfa0ed278299cfd5824f3"
                        },
                        {
                                "input": 0,
                                "key_count": 3,
                                "sig_count": 2,
                                "tx": "988b25d6a09c3c400f30e0be79f42a17fb974c582ed4edae1a33d5b9448f9651"
                        }
                ]
        }

## Ordinals in Tap Scripts

        $ curl -X POST -d '{"height":805348,"options":{"tap":true,"HumanReadable":true}}' http://127.0.0.1:8888/rest/v1/serialized_script_usage
        {
                "ordinals": [
                        {
                                "input": 0,
                                "mimetype": "text/plain;charset=utf-8",
                                "ord": "805348.bitmap",
                                "tx": "9ec6b09d4dc977f5d03b0af5ed618d37a22bb2097147834566a66de7ac2df403"
                        },
                        {
                                "input": 0,
                                "mimetype": "text/plain;charset=utf-8",
                                "ord": {
                                        "amt": "20.80936695",
                                        "op": "transfer",
                                        "p": "brc-20",
                                        "tick": "ordi"
                                },
                                "tx": "7c795e245831cc9b2f5c470e291f38072d9e42c487c8312506f9f13e2b0d3311"
                        }
                ]
        }

