# Screen Shots

### Input, Hex View, 2-of-3 Multisig

Here we see the four distinct pieces of input data required for a P2SH-P2WSH redemption. The redeem script, in this case,
is actually a standard P2WSH output script. The witness script is where the real work is done.

![P2SH-P2WSH 2-of-3 multisig (as hex)](/docs/images/screen-shots/p2sh-p2wsh-2-3-multisig-as-hex.png)

***

### Input, Type View, 2-of-3 Multisig

This is the same input as above but displaying data types instead of hex values.

![P2SH-P2WSH 2-of-3 multisig (as types)](/docs/images/screen-shots/p2sh-p2wsh-2-3-multisig-as-types.png)

***

### Input, Text View, Coinbase

![Genesis Coinbase (as text)](/docs/images/screen-shots/genesis-coinbase-as-text.png)

***

### Input, Text View, Ordinal

The text view is the best way to see the structure of a standard ordinal.
The OP_CHECKSIG handles the redemption of funds. After that, an OP_0 followed by OP_IF guarantees that the OP_IF
block will never execute. The rest of the fields are meta-data and data fields.

![Ordinal (as text)](/docs/images/screen-shots/ordinal-as-text.png)

***

### Input, Type View, Ordinal

This is the same ordinal as above viewed as data types.

![Ordinal (as text)](/docs/images/screen-shots/ordinal-as-types.png)

***

### Input, Text View, Ordinal

Ordinals can represent text files or binary files. Almost any file type can be used. This one is a small CSS file.

![Ordinal CSS (as text)](/docs/images/screen-shots/ordinal-css-as-text.png)

***

### Output, Text View, OP_RETURN

From transaction ac49f0d3117a02545d86efff49be45fe94cd99f901456088a3dcb0e816cb6927

![Ordinal (as text)](/docs/images/screen-shots/op-return-message-as-text.png)

***

### Transaction

Transaction results include some data about the overall transaction and a list of inputs and outputs.
The values and addresses for the inputs come from their previous outputs.
The inputs and outputs can be opened by clicking on them.

![Transaction Results](/docs/images/screen-shots/tx-results.png)

***

### Block

Block results include some data about the overall block and a list of transactions.

![Block Results](/docs/images/screen-shots/block-results.png)

***

