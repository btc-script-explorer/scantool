# Bitcoin Script Analytics Project

[![AGPL licensed](https://img.shields.io/badge/license-AGPL-blue.svg)](https://github.com/btc-script-explorer/explorer/blob/master/LICENSE)

## Features

The application provides 2 main services.

- Web Site (Block/Tx Explorer)
  - Displays serialized scripts (redeem scripts, witness scripts and tap scripts) which are displayed only in hex by most block explorers.
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

### Serialized Scripts

When people discuss Transaction Types in the bicoin blockchain, they are usually referring to Output Types.
However, outputs are relatively uninteresting from an analytical perspective. The only thing an output does is store a value and reveal a method (but not the actual data) required to redeem that value.

Inputs are the workhorses of the bitcoin blockchain. They do the heavy lifting required to transfer funds.
Many modern transactions are redeemed using serialized scripts. Visualizing these scripts is the best way to understand how they work, but most block explorers display them only as hex fields, the same way
they would display a signature or a public key. When we examine serialized scripts, we can see that there are a few more standard input types than there are output types. The serialized scripts themselves
might be standard redemption methods, or they might be custom scripts.

With serialized scripts, we are now seeing what could be called a standard within a standard. Ordinals would be an example of this. An ordinal is a standard serialized scripts that is included in a standard
Taproot Script Path redemption. More than 90% of all witness scripts and redeem scripts are Multisig transactions, which is another example of a standard within a standard.

In this discussion, we use the term Spend Type to describe the various ways that inputs redeem funds. Each Spend Type can redeem exactly one Output Type, but some Output Types are redeemable by multiple
Spend Types. Taproot is a good example. There are two Spend Types that redeem Taproot outputs, the Script Path and the Key Path.

Bitcoin Core recognizes 7 standard Output Types which can be redeemed using one of 10 standard Spend Types. These Output Types and their corresponding Spend Types, as well as the required contents of
the input script and segregated witness, are shown in the table below. The Output Types are listed by the names assigned to them by Bitcoin Core. The Spend Types are listed by their commonly-used "P2" names.
Since these are all standard methods for redeeming funds, the data in both the inputs and the outputs must be exactly as shown in the table below, otherwise one or both will be considered non-standard.

![Spend Types](/assets/images/spend-type-table.jpg)
