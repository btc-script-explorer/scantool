# Bitcoin Script Analytics Project

[![AGPL licensed](https://img.shields.io/badge/license-AGPL-blue.svg)](https://github.com/btc-script-explorer/explorer/blob/master/LICENSE)

## Features

The application provides 2 main services.
The REST API connects to an ordinary bitcoin node, but returns data that is not provided by most nodes or online block explorers.
The web interface can display anything the REST API returns.

- REST API
  - Returns all fields in serialized scripts (redeem scripts, witness scripts and tap scripts).
  - Returns Spend Types of inputs in addition to Output Types.
  - Returns the data types for all script and segregated witness fields.
  - Returns any other fields that a bitcoin node can return.
  - Can be used as a back end for custom analysis tools as well as custom user interfaces.

- Web Site (Block/Tx Explorer)
  - Displays all data returned by the REST API.
  - Allows the user to view script fields and segregated witness fields as hex, text or data types.
  - Displays both blocks and transactions. (Address searches are not currently supported.)
  - Can be configured to serve web pages on any network interface and port.

## Motivation

#### Serialized Scripts

When people discuss Transaction Types in the bicoin blockchain, they are usually referring to Output Types.
However, outputs are passive elements and therefore they are relatively uninteresting from an analytical perspective.
An output has two main roles: to store value, and to reveal the methods (but not the data) that can be used to redeem that value.

Unlike outputs, inputs are active elements. They are the workhorses of the bitcoin blockchain, and they do the heavy lifting required to redeem funds.
In general, there are two ways funds can be redeemed.

The first is to provide a signature and public key and then call OP_CHECKSIG. In the early days of bitcoin, these were the legacy pay-to-public-key outputs that had no address
format and mostly used an uncompressed public key. Today we have Taproot outputs and its Key Path spend type which uses a Schnorr Signature and a slightly different type of public key.
But the basic concept behind all key-based transaction types is exactly the same. They verify a signature against a public key and a sequence of bytes.

The second way to redeem funds is to use a custom script. In the old "Wild West" days of the bitcoin blockchain when redemptions of non-standard outputs were not uncommon,
custom functionality was often written directly into the input script, whereas the newer standards require the input script to be empty in most cases. Over the years, support for custom scripts was moved to
serialized scripts. A serialized script is like a script with a script. After the first script executes, the serialized script is popped off the stack, parsed and executed.
The success of the transaction depends on the success of the serialized script.

There are 3 types of serialized scripts.
- Redeem Scripts (BIP 16, 2012)
- Witness Scripts (BIP 143, 2016)
- Tap Scripts (BIP 341, 2020)

Seeing the contents of these scripts is essential to understanding how transactions work, but most block explorers display them only as hex fields, the same way
they would display a signature or a public key. Being able to analyze redeem scripts also tells us what bitcoin transactions are being used for over time.
Currently, approximately 90% of redeem scripts and witness scripts are multisig transactions, and nearly 99% of tap scripts are ordinals.

#### Spend Types

We use the term Spend Type to describe the standard methods that inputs use to redeem funds.
Each Spend Type can redeem exactly one Output Type, but some Output Types are redeemable by multiple Spend Types.
Bitcoin Core recognizes 7 standard Output Types. The table below shows the 10 standard Spend Types that can be used to redeem the standard Output Types.
The table also shows the required contents of the input script and segregated witness for each Spend Type.
The Output Types are listed by the names assigned to them by Bitcoin Core. The Spend Types are listed by their commonly-used "P2" names.
Since these are all standard methods for redeeming funds, the data in both the inputs and the outputs must be exactly as shown in the table below, otherwise one or both will be considered non-standard.

![Spend Types](/assets/images/spend-type-table.jpg)

#### Script Fields

The script fields and segregated witness fields that redeem outputs represent many different data types.
Therefore, it is useful to have a quick and easy way to view these fields as different types.
For example, a field in a script could be an op code, a signature, a public key, a hash, a text message, part of a binary file or some piece of data that is not easily identifiable.
Providing a way to view these fields as hex or text, or have the system attempt to tell us what types of data they are, is useful for anyone interested in analyzing
script usage as well as anyone who simply wants to learn how the system works.

#### Custom Projects

The REST API can be used as a back end for research projects which might focus on analysis of specific spend types, output types, script types, opcodes or anything else.
Such a program would, in most cases, be simple to write, could be written in almost any language and would make it easy to put large of amounts of data into a database
where it could be analyzed more thoroughly. The REST API can also be used as a back end for custom user interfaces.

## Screen Shots

## Usage

#### Requirements

1. Access to a bitcoin node that has transaction indexing enabled.

#### Download

#### Quick Start (Bitcoin Core)

Obviously, the following ip addresses, port numbers, username and password should be replaced by the ones used in your specific setup. Do not use the ones shown here.
Currently, Bitcoin Core is the only supported node, but other will likely be supported in the future.

1. Bitcoin Core Config Settings

        txindex=1
        rpcbind=192.168.1.99:9999
        rpcallowip=192.168.1.77
        rpcuser=node_username
        rpcpassword=node_password

2. Explorer Config Settings

        bitcoin-core-addr=192.168.1.99
        bitcoin-core-port=9999
        bitcoin-core-username=node_username
        bitcoin-core-password=node_password
        addr=192.168.1.77
        port=8080

3. Run the Explorer

        $ ./explorer --config-file=./explorer.conf 
        
        ************************************************
        *      Node: 192.168.1.99:9999 (Bitcoin Core)  *
        *  Explorer: 192.168.1.77:8080                 *
        ************************************************

4. View the Web User Interface in a Browser

        http://192.168.1.77:8080

4. Send a REST Request from the Command Line

        $ curl -X GET http://192.168.1.77:8080/rest/v1/current_block_height
        {"Current_block_height":803131}

