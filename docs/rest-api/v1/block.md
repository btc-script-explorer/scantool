# JSON Request Objects

## BlockOptions

Name | Type | Required | Default | Description
:---:|:---:|:---:|:---:|:---:
human_readable | bool | No | false | return human readable JSON

## BlockRequest

Name | Type | Required | Default | Description
:---:|:---:|:---:|:---:|:---:
hash | string | No | | block hash
height | uint32 | No | | block height
options | BlockOptions | No | not included | options

# Examples

## By Hash

BlockRequest

        {
                "hash": "0000000000000000000146140c9fc04604169e3227fa72eee1432c51f3ee95ca",
                "options": {
                        "human_readable": true
                }
        }

        $ curl -X POST -d '{"hash":"0000000000000000000146140c9fc04604169e3227fa72eee1432c51f3ee95ca","options":{"human_readable":true}}' http://127.0.0.1:8080/rest/v1/block

Block response

        {
                "hash": "0000000000000000000146140c9fc04604169e3227fa72eee1432c51f3ee95ca",
                "previous_hash": "0000000000000000000090484603d143880715d160d79e740f017e1f22c1a707",
                "next_hash": "00000000000000000000441a20faa76b2fd18249148ef67f8b26152e02992222",
                "height": 771982,
                "version": 536870912,
                "timestamp": 1673738660,
                "tx_ids": [
                        "69114ce23943ea95710b496db051e433ca9b0070b4eab85f53a9223ecc7ce5c9",
                        "abae967ad5780db9e3fe134e44bc7cace14527a804e857f15fac9032efd5082d",
                        "e2438edcb03f84007df593c4fea9eec9bd41894a896aa53c23d3bb5a2bcf5929",
                        "863cd611208dbdae88e1dc1a15fdc9db37b3c593e846cfb8d1613f14ac846ea8",
                        "b04742866aa3576c5035d7e1725f3584242669b1cbdcf1d94362ade08e88028c",
                        "7ddfa200dfacb422b0db0006f1bd0a9bc9d0dfc5ae51c8b670d373218df0ad01",
                        "06fbd17b746f068321d52571735ce67f5f7f52ba8b1ba78ce3dffbfbfc05b365",
                        "bf28798720f6c44189f44c2e86acaa5a044018b33ba52c6e4bb59adcfc86bdf4",
                        "7ec6fb2e59e0e593ab1d6aecae5a4e9172f175a2a7c500147babfaa3dc835951",
                        "afe3363fa4a6b65e1ccfc822649b91e4ed91094a84758a20e6f5d1b37121a9d4"
                ]
        }

## By Height

BlockRequest

        {
                "height": 772525,
                "options": {
                        "human_readable": true
                }
        }

        $ curl -X POST -d '{"height":772525,"options":{"human_readable":true}}' http://127.0.0.1:8080/rest/v1/block

Block response

        {
                "hash": "00000000000000000003368434623902477ad9de86f8ab9b7e8f5d42a0a819e5",
                "previous_hash": "00000000000000000000051870f680c15311c692f71bb473ed8b3b1b0a4a152d",
                "next_hash": "00000000000000000000bb745dfe66e258841673279849c4d396804154b62f30",
                "height": 772525,
                "version": 844341248,
                "timestamp": 1674046544,
                "tx_ids": [
                        "1ace110bf47d626831cd99ecc52db463d7348f187e3d47691009a6e361c2351c",
                        "7a0ca6da556397945d3e031375095ec1300c6eab9d3dab796bafda08887dfae1",
                        "9af12f75caf8b426b23624cda28df5f33a40f4926f307388f005de2f839e1b82",
                        "93ff7d57f1c596272d603ba82de128f6d7c454ac87ce067621bb1c79d05a3a01",
                        "227453b21205b52f6d543994c50f108b8895236315700662202ac0590861906c",
                        "920d478a1ce569ae36066560d65c37ac0a4a01cffc012b57f24cd14eec3a3f83",
                        "d5f48057c171434940237e842c41cba6269c334dd0a51015ce0dbaf8de484faf",
                        "805be6e09ae9d7c36250eceb814ac22790bb721c6aa2e78b34943cbbfa6d46ff",
                        "d3a9ce72b8612a7110ec52fb85398ba1081734facb38bc91579f89f8b056154f",
                        "53cf6fd51032495ac4c2abaeaf473be7df82a26d0b485ae06f0d3ea393773bbf",
                        "0a9c7552c9008ef17bd495dcab2753ea381c7217ed334cebd174fa1047e0e598",
                        "1ce00c920594470f96ba95e21b769c31f8ae03b3f1b51e91e1a34d34ee54aae2",
                        "274815126635790771eca56db754b7bf6b3e54de6ebc246dd347212c6ecf012e"
                ]
        }

## Empty Request

BlockRequest

When neither hash nor height is included in the request, the most recent block will be returned.

        {
                "options": {
                        "human_readable": true
                }
        }

        $ curl -X POST -d '{"options":{"human_readable":true}}' http://127.0.0.1:8080/rest/v1/block

Block response

        {
                "hash": "0000000000000000000137ced007fddf254c01c8771f2e8591db63b3cd531b2e",
                "previous_hash": "000000000000000000035acb4404247eafdf7d4ead9e2412ea75d71e55fedb11",
                "next_hash": "000000000000000000075fce527f1d1c500c7d13406325a80092ad79c65dc53c",
                "height": 769895,
                "version": 536895488,
                "timestamp": 1672591723,
                "tx_ids": [
                        "826ec483b6daad9a941680b870f6b0546087f49aace234f821ad3eabeb7cf738",
                        "55e7d863b62569d0b180a6d36511a49cb46b11470c077c962d67294d4ccfbddf",
                        "244eb1b6205a0b76f32169f80ea872019f92d6288586c2e5e62fdd4a817fd8d0",
                        "929c2d4f780a8ed935c6b0d71f48458bb6bdefb8c35e4b43ee4bbc51d1194f40"
                ]
        }

