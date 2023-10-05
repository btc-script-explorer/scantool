# Script Analytics Tool

[![AGPL licensed](https://img.shields.io/badge/license-AGPL-blue.svg)](https://github.com/btc-script-explorer/scantool/blob/master/LICENSE)

## What is it?

The **SC**ript **AN**alytics **TOOL** (SCANTOOL) is a tool for analyzing bitcoin scripts, including serialized scripts. It provides the following features.

- a REST API
- a web-based block and transaction explorer

It can be used as a research tool and/or a learning utility that provides data that other APIs and user interfaces do not.
It is intended to be used in a private network.

## How It Is Different Than Other REST APIs And Block Explorers

#### Spend Types

When people discuss transaction types in the bicoin blockchain, they are usually referring to standard output types.
But output types are only half of the process of transacting on the blockchain.
Inputs do most of the work of transfering funds, and they have standard types too.

There are 7 standard redeemable output types, and 10 standard input types, which we refer to here as spend types.
Each spend type can redeem exactly one output type, but some output types are redeemable by multiple spend types.

An example of the difference between the two would be Taproot. Taproot is actually an output type. There is a specific format that identifies a Taproot output.
But looking at the output tells us nothing about how it will be redeemed. Taproot outputs can be redeemed using the Key Path spend type or the Script Path spend type.
Taproot is a single output types that can be redeemed with one of two spend types.

Most bitcoin node APIs and online block explorers do not identify input types.
The SCANTOOL identifies both output types and spend types.

The table below shows which spend types can be used to redeem which output types.
It also shows the required contents of the input script and segregated witness for each spend type.
The output types are listed by the names assigned to them by Bitcoin Core. The spend types are listed by their commonly-used "P2" (pay-to) names.
Since these are all standard methods for redeeming funds, all input data must be exactly as shown in the table below with almost no exceptions, otherwise the redemption method will be considered non-standard.

![Spend Types](/docs/images/spend_types.png)

#### Serialized Scripts

A serialized script is a script included as a field in an input script or segregated witness block.
These scripts allow bitcoin transactions to be customizable.

There are 3 types of serialized scripts:
- Redeem Script (BIP 16) is the last field of an input script that redeems a P2SH output.
- Witness Script (BIP 143) is the last field in the segregated witness for a P2SH-P2WSH or P2WSH input.
- Tap Script (BIP 341) is the field immediately before the control block in the segregated witness for a Taproot Script Path input.

The need for serialized scripts evolved out of the need for multisig transactions.
When the pay-to-scripthash output type was introduced, multisig wallets started using serialized scripts instead of the legacy multisig transaction type.
Today, more than 93% of redeem scripts and witness scripts are used for multisig transactions. Nearly 99% of tap scripts are used for ordinals.

The 10 standard spend types can be divided into 5 classes of 2 transaction types. Each class provides one key-based and one script-based method for redeeming outputs.
The script-based spend types, with the exception of the legacy pay-to-key multisig transactions, all contain serialized scripts.

![Transaction Generations](/docs/images/spend_type_classes.png)

When a script-based transaction is confirmed, the serialized script is parsed and executed, and must succeed in order for the transaction to succeed.
Therefore, viewing the contents of serialized scripts is essential to understanding how script-based transactions work,
but most bitcoin node APIs and online block explorers display them only as hex fields, the same way they would display a signature or a public key.

The SCANTOOL provides fully parsed serialized scripts.

#### Script Field Data Types

The bitcoin blockchain contains a variety of different types of data, many of which have little or nothing to do with monetary transactions.
For example, a segregated witness field could be a signature, a public key or a hash.
It could also be a text message, a hex representation of a section of a binary file or some piece of data that is not easily identifiable.
A script field could be any of those things as well, and it could also be an opcode.
Having a way to view these fields by their data type can be useful for anyone interested in analyzing script usage as well as anyone who simply wants to learn how bitcoin transactions work.

(See the [Screen Shots](/docs/screen-shots.md) section for examples.)

#### Custom Projects

The REST API can be used as a back end for research projects which might focus on analysis of specific spend types, output types, script types, opcodes or anything else of interest.
A client application could be written in almost any language in a relatively short period of time and could be used to put large amounts of data into a database where it can be analyzed more thoroughly.
(See the [Blockchain Analysis](/docs/rest-api/v1/blockchain_analysis.md) section for examples.) The REST API can also be used as a back end for custom user interfaces.

## Usage

#### Requirements

1. Access to a bitcoin node that has transaction indexing enabled.

#### Download

Go to the [Releases](https://github.com/btc-script-explorer/scantool/releases/latest) page and download the appropriate file for your system.

#### Quick Start (with Bitcoin Core)

**Each build comes with its own QUICK START GUIDE.**

For this example, we will configure and run the scantool connecting to Bitcoin Core.
Our example will assume the following:
- the Bitcoin node's RPC service is available at **192.168.1.99:9999**
- the Bitcoin node allows RPC connections from **192.168.1.77**.
- the scantool REST API and web interface will be available at **192.168.1.77:8080**

(Obviously, the ip addresses, port numbers, username and password shown here must be replaced by the ones used in your specific setup. **Do not use the example values shown here.**)

1. Make sure the following Bitcoin Core settings are set. The txindex setting must be set to 1.
   **Use the values for your specific system.**

        txindex=1
        rpcbind=192.168.1.99:9999
        rpcallowip=192.168.1.77
        rpcuser=node_username
        rpcpassword=node_password

2. Create a file called scantool.conf in the same directory as the scantool executable, and put the following settings in it. (Other file locations can also be used.)
   **Use the values for your specific system.**

        bitcoin-core-addr=192.168.1.99
        bitcoin-core-port=9999
        bitcoin-core-username=node_username
        bitcoin-core-password=node_password
        addr=192.168.1.77
        port=8080

3. Run the scantool.

        $ ./scantool --config-file=./scantool.conf
        
        *****************************************************************************
        *                                                                           *
        *                             SCANTOOL 1.0.0                                *
        *                                                                           *
        *  Node: Bitcoin Core 25.0.0                                                *
        *        192.168.1.99:9999                                                  *
        *                                                                           *
        *   Web: http://192.168.1.77:8080/web/                                      *
        *                                                                           *
        *  REST: curl -X GET http://192.168.1.77:8080/rest/v1/current_block_height  *
        *        curl -X POST -d '{}' http://192.168.1.77:8080/rest/v1/block        *
        *                                                                           *
        *  Caching: On                                                              *
        *                                                                           *
        *****************************************************************************



4. View the web interface in a browser.

        http://192.168.1.77:8080/web/

5. Send a REST request from the command line.

        $ curl -X GET http://192.168.1.77:8080/rest/v1/current_block_height
        {"current_block_height":803131}

## Documentation

### Settings

All settings on the command line should begin with "--". In the config file, the "--" should not be present.

Setting | Required | Default | Description
---|---|---|---
bitcoin-core-addr | Yes | | The IP address from a rpcbind setting in Bitcoin Core.
bitcoin-core-port | Yes | | The port number from the same rpcbind setting in Bitcoin Core.
bitcoin-core-username | Yes | | The rpcuser setting in Bitcoin Core.
bitcoin-core-password | Yes | | The rpcpassword setting in Bitcoin Core.

Setting | Required | Default | Description
---|---|---|---
addr | if no-web=false | | The IP address the web interface should be available on.
port | if no-web=false | | The port number the web interface should be available on.

Setting | Required | Default | Description
---|---|---|---
no-web | No | false | Turns off the web interface.
caching | No | false | Turns caching on for better performance.
config-file | No | | Location of the config file. Only applicable on the command line.

\* Cache size is not currently managed, so the cache will only grow.

### Web Interface

The web interface allows search by:
- block hash
- block height
- transaction id

For more information, see the [screen shots](/docs/screen-shots.md).

### REST API

- [JSON Responses](/docs/rest-api/v1/json_response_objects.md)
- JSON Requests
  - [Block](/docs/rest-api/v1/block.md)
  - [Transaction](/docs/rest-api/v1/tx.md)
  - [Input](/docs/rest-api/v1/input.md)
  - [Output](/docs/rest-api/v1/output.md)
  - [Current Block Height](/docs/rest-api/v1/current_block_height.md)
  - [Blockchain Analysis/Research](/docs/rest-api/v1/blockchain_analysis.md)

