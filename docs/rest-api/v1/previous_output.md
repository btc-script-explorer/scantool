# JSON Objects

## PreviousOutputRequest

        {
                InputTxId string
                InputIndex uint16
                PrevOutTxId string
                PrevOutIndex uint16
        }

- **InputTxId**: Required. Transaction id of the input that the previous output belongs to.
- **InputIndex**: Required. Index of the input that the previous output belongs to.
- **PrevOutTxId**: Required. Transaction id of the previous output.
- **PrevOutIndex**: Required. Index of the previous output.

***

## PreviousOutputResponse

        {
                InputTxId string
                InputIndex uint16
                PrevOut Output
        }

- **InputTxId**: Transaction id of the input that the previous output belongs to.
- **InputIndex**: Index of the input that the previous output belongs to.
- **PrevOut**: Previous output for this input.

***

# Example

        $ curl -X POST -d '{"InputTxId":"abdcfbd8b77c4b14372c58cd7bfc5a09ad5c04759c6699b0eaa19e9226746571","InputIndex":1,"PrevOutTxId":"ba6ce05c8e646b13b41ae44d23281ddcdbafeb64396b7d87855c233685a1400a","PrevOutIndex":0}' http://127.0.0.1:8888/rest/v1/previous_output
        {"InputTxId":"abdcfbd8b77c4b14372c58cd7bfc5a09ad5c04759c6699b0eaa19e9226746571","InputIndex":1,"PrevOut":{"Value":23248802,"OutputType":"Taproot","Address":"bc1pp2767q84l8ytnftxudxvyfs4y9z34r2dqr8ltj59pg6ysvf607qqcwwgdw","OutputScript":{"Fields":[{"Hex":"OP_1","Type":"OP_1"},{"Hex":"0abdaf00f5f9c8b9a566e34cc2261521451a8d4d00cff5ca850a3448313a7f80","Type":"Witness Program (Public Key)"}]}}}

