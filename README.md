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
While inputs do most of the work of transfering funds, most block explorers show only output types.

There are 7 standard redeemable output types. There are also 10 standard input types, which we refer to here as spend types.
Each spend type can redeem exactly one output type, but some output types are redeemable by multiple spend types.

Most REST APIs available through bitcoin nodes do not return any information about spend types and it is very hard to find a block explorer that identifies them.
The SCANTOOL does both.

The table below shows which spend types can be used to redeem which output types.
It also shows the required contents of the input script and segregated witness for each spend type.
The output types are listed by the names assigned to them by Bitcoin Core. The spend types are listed by their commonly-used "P2" (pay-to) names.
Since these are all standard methods for redeeming funds, all input data must be exactly as shown in the table below with almost no exceptions, otherwise the redemption method will be considered non-standard.

![Spend Types](/docs/images/spend-type-table.png)

#### Serialized Scripts

There are 3 types of serialized scripts:
- Redeem Script (BIP 16) is the last field of any input script that redeems a P2SH output.
- Witness Script (BIP 143) is the last segregated witness field in a P2SH-P2WSH or P2WSH input.
- Tap Script (BIP 341) is the segregated witness field before the control block in a Taproot Script Path input.

Serialized scripts appear in 4 of the 10 standard spend types.

The 10 standard spend types can be divided into 5 generations of bitcoin transaction types, each of which provides one key-based and one script-based method for redeeming outputs.
In each of the script-based spend types, a serialized script is provided with the input data. The legacy MultiSig spend type is the ancestor of modern script-based spend types,
but it does not actually contain a serialized script. The "generations" shown here did not necessarily evolve in the order they appear in the table below.

![Transaction Generations](/docs/images/tx-generations.png)

When a script-based transaction is confirmed, the serialized script is parsed and executed, and must succeed in order for the transaction to succeed.
Viewing the contents of serialized scripts is essential to understanding how transactions work, but most block explorers display them only as hex fields, the same way
they would display a signature or a public key.

The web-based explorer provided with the SCANTOOL displays fully parsed serialized scripts and provides information about them that few, if any, other tools do.

#### Script Field Data Types

A segregated witness field could be a signature, a public key, a hash, a text message or some piece of data that is not easily identifiable.
A script field could be any of those things as well, or it could also be an opcode.
Having a way to view these fields by their data type is useful for anyone interested in analyzing script usage as well as anyone who simply wants to learn how the system works.
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

### [Settings](/docs/app-settings.md)

### [Web Application](/docs/screen-shots.md)

### REST API

- [JSON Responses](/docs/rest-api/v1/json_response_objects.md)
- JSON Requests
  - [Block](/docs/rest-api/v1/block.md)
  - [Transaction](/docs/rest-api/v1/tx.md)
  - [Input](/docs/rest-api/v1/input.md)
  - [Output](/docs/rest-api/v1/output.md)
  - [Current Block Height](/docs/rest-api/v1/current_block_height.md)
  - [Blockchain Analysis/Research](/docs/rest-api/v1/blockchain_analysis.md)

