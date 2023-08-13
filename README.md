# Bitcoin Script Analytics Project

[![AGPL licensed](https://img.shields.io/badge/license-AGPL-blue.svg)](https://github.com/btc-script-explorer/explorer/blob/master/LICENSE)

## Features

The application is primarily a research tool that can also be used as a learning tool with a focus on analysis of scripts. It provides 2 main services.

- Web Site (Block/Tx Explorer)
  - Displays parsed serialized scripts (redeem scripts, witness scripts and tap scripts) which are displayed only in hex by most block explorers.
  - Displays Spend Types of inputs in addition to Output Types.
  - Allows the user to view script fields and segregated witness fields as hex, text or data types.
  - Displays both blocks and transactions. (Address searches are not currently supported.)
  - Can be configured to serve web pages on any network interface and port.

- REST API
  - Returns everything the web site displays, plus other data.
  - Could be used (by a custom-written program) to store specific data in a database. Such data could be used to monitor and analyze various different things in the blockchain over time. Some of things it could be used to analyze are:
    - Spend Types
    - Output Types
    - Ordinals
    - Multisig usage in serialized scripts
    - Usage of specific opcodes

## Motivation

When people discuss "transaction types" in the bicoin blockchain, they are usually referring only to output types.
However, outputs are relatively uninteresting from an analytical perspective. The only thing an output does is store a value and reveal a method (but not the actual data) required to redeem that value.
The redemption method for an output is what people usually refer to when they talk about transaction types. Some examples of output types are Taproot, p2wsh or even non-standard.

Inputs are the workhorses of the bitcoin blockchain. They do the heavy lifting required to transfer funds from one address to another. Inputs are also much more complex than outputs.
In some cases, an input script might contain a field that can be parsed as an entirely separate script.
In other cases, one of the segregated witness fields might be a script.
These serialized scripts must be parsed and executed as part of the verification of the transaction.

Understanding these serialized scripts is crucial in understanding how many inputs redeem funds from their previous outputs. However, since the serialized scripts are really just stack items
provided as a field in the input script or segregated witness, most block explorers display them only as hex, the same as they would for a signature or a public key. This makes it very difficult to
visualize how the funds are actually redeemed.

In this discussion, we use the term Spend Type to describe the various ways that inputs redeem funds. Each Spend Type can redeem exactly one Output Type, but an Output Type might be redeemable by multiple
Spend Types. Taproot is a good example. There are two spend paths for Taproot outputs, the Script Path and the Key Path. However, the Output Types are exactly the same and indistinguishable from one another.
An Output Type is known when the output is created, long before it is spent. A Spend Type might not be known until the funds are actually redeemed.

Bitcoin Core recognizes 7 standard Output Types which can be redeemed using one of 10 different standard Spend Types. These Output Types and their possible Spend Types, as well as the required contents of
the input script and segregated witness, are shown in the table below. The Output Types are listed by the names given to them by Bitcoin Core. The Spend Types are listed by their commonly-used "p2" format.
Since these are all standard methods for redeeming funds, the data in both the inputs and the outputs must be exactly as shown in the table below, otherwise one or both will be considered non-standard.

![Spend Types](/assets/images/spend-type-table.jpg)
