# Bitcoin Script Analytics Project

[![AGPL licensed](https://img.shields.io/badge/license-AGPL-blue.svg)](https://github.com/btc-script-explorer/explorer/blob/master/LICENSE)

## What It Is

- a research and learning utility that provides data that other APIs and user interfaces do not
- intended to be used in a private network

## What It Is Not

- a wallet application

## How It Is Different Than Other REST APIs And Block Explorers

#### Spend Types

When people discuss transaction types in the bicoin blockchain, they are usually referring to standard output types.
Bitcoin Core recognizes 7 standard output types.

But inputs do the work of transfering funds and there is very little software, either REST APIs or block explorers, that does a good job of identifying their types.
There are 10 standard input types which we will call spend types.
Each spend type can redeem exactly one output type, but some output types are redeemable by multiple spend types.

The table below shows which spend types can be used to redeem which output types.
It also shows the required contents of the input script and segregated witness for each spend type.
The output types are listed by the names assigned to them by Bitcoin Core. The spend types are listed by their commonly-used "P2" names.
Since these are all standard methods for redeeming funds, all input data must be exactly as shown in the table below, otherwise the redemption method will be considered non-standard.

![Spend Types](/assets/images/spend-type-table.png)

#### Serialized Scripts

There have been 5 generations of standard bitcoin spend types, each of which provides one key-based and one script-based method for redeeming outputs.
In each of the script-based types, there is a serialized script provided in the input data. That script is parsed and executed and must succeed in order for the transaction to succeed.
The legacy multisig scripts were the ancestors of modern serialized scripts. The "generations" shown here did not necessarily evolve in the order they appear in the table below.

![Transaction Generations](/assets/images/tx-generations.png)

There are 3 types of serialized scripts:
- Redeem Script (BIP 16, 2012) is the last field of any input script that redeems a P2SH output.
- Witness Script (BIP 143, 2016) is the last segregated witness field in a P2SH-P2WSH or P2WSH input.
- Tap Script (BIP 341, 2020) is the segregated witness field before the control block in a Taproot Script Path input.

Viewing the contents of serialized scripts is essential to understanding how transactions work, but most block explorers display them only as hex fields, the same way
they would display a signature or a public key. Analysis of serialized scripts also allows us to see what serialized scripts are being used for.
Currently, approximately 90% of redeem scripts and witness scripts are standard multisig transactions, and nearly 99% of tap scripts are ordinals.

#### Script Field Data Types

Script fields and segregated witness fields can represent many different types of data.
Therefore, it is useful to have a quick and easy way to view these fields as different types, or have the system identify which types they appear to be.
For example, a field in a script could be an op code, a signature, a public key, a hash, a text message, part of a binary file or some piece of data that is not easily identifiable.
Having a way to change viewing modes for these fields would be useful for anyone interested in analyzing script usage as well as anyone who simply wants to learn how the system works.
(See the [Screen Shots](/examples/screen-shots.md) section for examples.)

#### Custom Projects

The REST API can be used as a back end for research projects which might focus on analysis of specific spend types, output types, script types, opcodes or anything else of interest.
A client application could be written in almost any language in a relatively short period of time and could be used to put large amounts of data into a database where it can be analyzed more thoroughly.
The REST API can also be used as a back end for custom user interfaces.

## Usage

#### Requirements

1. Access to a bitcoin node that has transaction indexing enabled.

#### Download

#### Quick Start (Bitcoin Core)

1. The following Bitcoin Core settings are required.
Obviously, the ip addresses, port numbers, username and password shown here must be replaced by the ones used in your specific setup. Do not use the example values shown here.

        txindex=1
        rpcbind=192.168.1.99:9999
        rpcallowip=192.168.1.77
        rpcuser=node_username
        rpcpassword=node_password

2. Create a file called explorer.conf in the same directory as the explorer executable. (Other locations can also be used.)
As explained above, the ip addresses, port numbers, username and password shown here must be replaced by the ones used in your specific setup. Do not use the example values shown here.

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

4. View the Web Interface in a Browser

        http://192.168.1.77:8080

5. Send a REST Request from the Command Line

        $ curl -X GET http://192.168.1.77:8080/rest/v1/current_block_height
        {"Current_block_height":803131}

## Documentation

#### [Settings](/docs/app-settings.md)

#### [Website Screen Shots](/examples/screen-shots.md)

#### [Block](/docs/rest-api/v1/block.md)
#### [Transaction](/docs/rest-api/v1/tx.md)
#### [Output Types](/docs/rest-api/v1/output_types.md)
#### [Previous Output](/docs/rest-api/v1/previous_output.md)
#### [Current Block Height](/docs/rest-api/v1/current_block_height.md)
#### [Serialize Script Usage](/docs/rest-api/v1/serialized_script_usage.md)

