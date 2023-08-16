# Bitcoin Script Analytics Project

[![AGPL licensed](https://img.shields.io/badge/license-AGPL-blue.svg)](https://github.com/btc-script-explorer/explorer/blob/master/LICENSE)

## Features

The application provides 2 main services.
- The REST API that connects to an ordinary bitcoin node, but returns data that is not provided by most nodes or online block explorers. See the Motivation section below for details.
- The web interface can display anything the REST API returns.

## Motivation

#### Spend Types

When people discuss Transaction Types in the bicoin blockchain, they are usually referring to standard Output Types.
Bitcoin Core recognizes 7 standard Output Types.
But output types tell less than half the story. Outputs themselves are passive elements. 
Their main roles are to store value and to reveal the methods (but not the data) that can be used to redeem that value.
Other than that, they do nothing. Outputs just sit there in the blockchain waiting to be redeemed by an input.

Inputs are the workhorses of the bitcoin system. They do all of the work invovled in redeeming outputs.
There are 10 standard input types which we will refer to here as Spend Types.
Each Spend Type can redeem exactly one Output Type, but some Output Types are redeemable by multiple Spend Types.
The table below shows which Spend Types can be used to redeem which Output Types.
It also shows the required contents of the input script and segregated witness for each Spend Type.

The Output Types are listed by the names assigned to them by Bitcoin Core. The Spend Types are listed by their commonly-used "P2" names.
Since these are all standard methods for redeeming funds, all input data must be exactly as shown in the table below, otherwise the redemption method will be considered non-standard.

![Spend Types](/assets/images/spend-type-table.png)

#### Serialized Scripts

There have been 5 generations of standard bitcoin spend types, each of which provides one key-based and one script-based method for redeeming outputs. The legacy multisig transactions were the
ancestors of modern script-based transaction types. In each of the script-based types, there is a serialized script provided in the input data. For these transaction types, the serialized script
is parsed and executed as part of the verification of the transaction. These "generations" did not necessarily evolve in the order they appear in the table below.

![Transaction Generations](/assets/images/tx-generations.png)

There are 3 types of serialized scripts:
- Redeem Script (BIP 16, 2012) is the last field of an input script that redeems a P2SH output.
- Witness Script (BIP 143, 2016) is the last segregated witness field in a P2SH-P2WSH or P2WSH input.
- Tap Script (BIP 341, 2020) is the segregated witness field before the control block in a Taproot Spend Path input.

Viewing the contents of serialized scripts is essential to understanding how transactions work, but most block explorers display them only as hex fields, the same way
they would display a signature or a public key. Analysis of serialized scripts also allows us to see what various spend types are being used for.
Currently, approximately 90% of redeem scripts and witness scripts are multisig transactions, and nearly 99% of tap scripts are ordinals.

#### Script Fields

Script fields and segregated witness fields can represent many different types of data.
Therefore, it is useful to have a quick and easy way to view these fields as different types, or have the system identify which types they appear to be.
For example, a field in a script could be an op code, a signature, a public key, a hash, a text message, part of a binary file or some piece of data that is not easily identifiable.
Having a way to change viewing modes for these fields would be useful for anyone interested in analyzing script usage as well as anyone who simply wants to learn how the system works.
(See the Screen Shots section below for examples.)

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

