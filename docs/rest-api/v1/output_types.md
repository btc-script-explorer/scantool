# JSON Objects

## OutputTypesRequest

        {
                string [ uint16 ]
        }

- **Key**: Transaction id that contains the output in question.
- **Value**: Indexes of the outputs for which output types are to be obtained.

***

## OutputTypesResponse

        {
                string string
        }

- **Key**: Outpoint of the output for which the output type is to be obtained.
- **Value**: The output type of the output.

***

# Example

		$ curl -X POST -d '{"72b329e94dab613c6dd82c670a7deb8c00c406af0a776b3ed70e92b4be9760ba":[88,143],"7e9b26fef1afa0523a420d4747077b8d9d44defd118d9d6447a1e56ecdc0dd05":[122],"8b5d1c025d50f254352203facf2d893547ec2b5ee786dcaf7fb8e9f9ce8222a0":[1],"a28bbb62247f9c8b4a00baad467900df7f2e78dd763128c40c1d82e5c2c69fd7":[121]}' http://127.0.0.1:8080/rest/v1/output_types
        {"72b329e94dab613c6dd82c670a7deb8c00c406af0a776b3ed70e92b4be9760ba:143":"P2PKH","72b329e94dab613c6dd82c670a7deb8c00c406af0a776b3ed70e92b4be9760ba:88":"P2PKH","7e9b26fef1afa0523a420d4747077b8d9d44defd118d9d6447a1e56ecdc0dd05:122":"P2PKH","8b5d1c025d50f254352203facf2d893547ec2b5ee786dcaf7fb8e9f9ce8222a0:1":"P2PKH","a28bbb62247f9c8b4a00baad467900df7f2e78dd763128c40c1d82e5c2c69fd7:121":"P2PKH"}

