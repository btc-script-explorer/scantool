		request:
		{
			"f742b911259dd11278e9e3d34f2538c7d77837daef15fc00a047e3f13253aa0b": [0, 24],
			"f76874221ce8d7961f91b9ad5e827fb558d5ce95bc60b2722112d3384069c61b": [17, 21],
			"f7cb6da78444ff5029c5bd1f9ae6c64a1d2d65c604f2989c90cf9dc323692ed8": [0, 2]
		}
		curl -X POST -d "{\"f742b911259dd11278e9e3d34f2538c7d77837daef15fc00a047e3f13253aa0b\":[0,24],\"f76874221ce8d7961f91b9ad5e827fb558d5ce95bc60b2722112d3384069c61b\":[17,21],\"f7cb6da78444ff5029c5bd1f9ae6c64a1d2d65c604f2989c90cf9dc323692ed8\":[0,2]}" http://127.0.0.1:8888/rest/v1/output_types

		response:
		{
			"f742b911259dd11278e9e3d34f2538c7d77837daef15fc00a047e3f13253aa0b:0": "P2PKH",
			"f742b911259dd11278e9e3d34f2538c7d77837daef15fc00a047e3f13253aa0b:24": "P2PKH",
			"f76874221ce8d7961f91b9ad5e827fb558d5ce95bc60b2722112d3384069c61b:17": "P2PKH",
			"f76874221ce8d7961f91b9ad5e827fb558d5ce95bc60b2722112d3384069c61b:21": "P2PKH",
			"f7cb6da78444ff5029c5bd1f9ae6c64a1d2d65c604f2989c90cf9dc323692ed8:0": "P2PKH",
			"f7cb6da78444ff5029c5bd1f9ae6c64a1d2d65c604f2989c90cf9dc323692ed8:2": "P2PKH"
		}

The output_types API call will return only spend types identified by their previous outpoint.
(If you need them to be identified by their transaction id and input index, the previous_output API call must be used.)






		request:
		{
			"InputTxId":"801906494bfa5710e3a47131640859222abf52391de5800844a79fd148d5a658",
			"InputIndex":0,
			"PrevOutTxId":"f742b911259dd11278e9e3d34f2538c7d77837daef15fc00a047e3f13253aa0b",
			"PrevOutIndex":24
		}
		curl -X POST -d '{"InputTxId":"801906494bfa5710e3a47131640859222abf52391de5800844a79fd148d5a658","InputIndex":0,"PrevOutTxId":"f742b911259dd11278e9e3d34f2538c7d77837daef15fc00a047e3f13253aa0b","PrevOutIndex":24}' http://127.0.0.1:8888/rest/v1/previous_output

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


