# Screen Shots

### P2SH-P2WSH 2-of-3 multisig input, viewed as hex

Here we see the four distinct pieces of input data required for a P2SH-P2WSH redemption. The redeem script, in this case,
is actually a standard P2WSH output script. The witness script is where the real work is done.

![P2SH-P2WSH 2-of-3 multisig (as hex)](/assets/images/screen-shots/p2sh-p2wsh-2-3-multisig-as-hex.png)

***

### P2SH-P2WSH 2-of-3 multisig input, viewed as data types

This is the same input as above but displaying data types instead of hex values.

![P2SH-P2WSH 2-of-3 multisig (as types)](/assets/images/screen-shots/p2sh-p2wsh-2-3-multisig-as-types.png)

***

### Genesis Block Coinbase Script, viewed as text

![Genesis Coinbase (as text)](/assets/images/screen-shots/genesis-coinbase-as-text.png)

***

### Ordinal, viewed as text

Ordinals could be described as a standard within a standard. This is a standard Taproot Script Path redemption
with a tap script that also conforms to another standard. Nearly 99% of tap scripts during 2023 have been ordinals.
They seem to have peaked in late spring and early summer but the numbers appear to have dropped off after that.
The text view is the best way to see the structure of a standard ordinal.
The OP_CHECKSIG handles the redemption of funds. After that, an OP_0 followed by OP_IF guarantees that the OP_IF
block will never execute. The rest of the fields are meta-data and data fields.

![Ordinal (as text)](/assets/images/screen-shots/ordinal-as-text.png)

***

### Ordinal, viewed as data types

This is the same ordinal as above viewed as data types.

![Ordinal (as text)](/assets/images/screen-shots/ordinal-as-types.png)

***

### Ordinal representing a file, viewed as text

Ordinals can represent text files or binary files. Almost any file type can be used. This one is a small CSS file.

![Ordinal CSS (as text)](/assets/images/screen-shots/ordinal-css-as-text.png)

***

### OP_RETURN output message, viewed as text

From transaction ac49f0d3117a02545d86efff49be45fe94cd99f901456088a3dcb0e816cb6927

![Ordinal (as text)](/assets/images/screen-shots/op-return-message-as-text.png)

***

### Transaction Results

Transaction results include some data about the overall transaction and a list of inputs and outputs.
The values and addresses for the inputs come from their previous outputs.
The inputs and outputs can be opened by clicking on them.

![Transaction Results](/assets/images/screen-shots/tx-results.png)

***

### Block Results

Block results include some data about the overall block, optionally-loaded spend type and output type charts and a list of transactions.

![Block Results](/assets/images/screen-shots/block-results.png)

***

