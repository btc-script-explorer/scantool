# Script Analytics Tool

[![AGPL licensed](https://img.shields.io/badge/license-AGPL-blue.svg)](https://github.com/btc-script-explorer/scantool/blob/master/LICENSE)

The **Sc**ript **An**alytics **Tool** (scantool) is a REST API as well as a web-based block explorer. It can be used as a research tool and/or a learning utility that provides data that other APIs and user interfaces do not.
It is intended to be used in a private network, although it doesn't necessarily have to be. It is not intended to be a wallet application.

## How It Is Different Than Other REST APIs And Block Explorers

#### Spend Types

When people discuss transaction types in the bicoin blockchain, they are usually referring to standard output types.
However, inputs do the work of transfering funds, but most block explorers do not distinguish between input types, only output types.

There are 7 standard output types and 10 standard input types, which we will call spend types.
Each spend type can redeem exactly one output type, but some output types are redeemable by multiple spend types.

It is very hard to find a RPC API or block explorer that identifies spend types in addition to output types.
This application provides both.

The table below shows which spend types can be used to redeem which output types.
It also shows the required contents of the input script and segregated witness for each spend type.
The output types are listed by the names assigned to them by Bitcoin Core. The spend types are listed by their commonly-used "P2" (pay-to) names.
Since these are all standard methods for redeeming funds, all input data must be exactly as shown in the table below with almost no exceptions, otherwise the redemption method will be considered non-standard.

![Spend Types](/docs/images/spend-type-table.png)

#### Serialized Scripts

There have been 5 generations of standard bitcoin transaction types, each of which provides one key-based and one script-based method for redeeming outputs.
In each of the script-based types, there is a serialized script provided in the input data. The serialized script is parsed and executed, and it must succeed in order for the transaction to succeed.
The legacy multisig scripts were the ancestors of modern serialized scripts. The "generations" shown here did not necessarily evolve in the order they appear in the table below.

![Transaction Generations](/docs/images/tx-generations.png)

There are 3 types of serialized scripts:
- Redeem Script (BIP 16, 2012) is the last field of any input script that redeems a P2SH output.
- Witness Script (BIP 143, 2016) is the last segregated witness field in a P2SH-P2WSH or P2WSH input.
- Tap Script (BIP 341, 2020) is the segregated witness field before the control block in a Taproot Script Path input.

Viewing the contents of serialized scripts is essential to understanding how transactions work, but most block explorers display them only as hex fields, the same way
they would display a signature or a public key.

This application parses serialized scripts and provides data about them that few, if any, other analytics tools do.

#### Script Field Data Types

Script fields and segregated witness fields can represent many different types of data.
Therefore, it is useful to have a quick and easy way to view these fields as different types, or have the system identify which types they appear to be.
For example, a field in a script could be an op code, a signature, a public key, a hash, a text message, part of a binary file or some piece of data that is not easily identifiable.
Having a way to change viewing modes for these fields is useful for anyone interested in analyzing script usage as well as anyone who simply wants to learn how the system works.
(See the [Screen Shots](/docs/screen-shots.md) section for examples.)

#### Custom Projects

The REST API can be used as a back end for research projects which might focus on analysis of specific spend types, output types, script types, opcodes or anything else of interest.
A client application could be written in almost any language in a relatively short period of time and could be used to put large amounts of data into a database where it can be analyzed more thoroughly.
(See the [Blockchain Analysis](/docs/rest-api/v1/blockchain_analysis.md) section for examples.) The REST API can also be used as a back end for custom user interfaces.

## Caveats

#### Blocks slow to load

When identifying spend types, it is necessary to get the previous outputs for most inputs. For certain blocks, depending on what types of transactions exist in the block,
this can take longer than for other blocks. In most cases, it will only take seconds to load a block. But in some cases, it can take up to a few minutes.
It is not as much of a problem for automated scripts accessing the REST API than it is for the web-based block explorer.
In the future, settings will be added to help speed up the process by ignoring unneeded data.

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

1. Make sure the following Bitcoin Core settings are set. Make sure txindex is set to 1 and your IP addresses and port numbers are correct.
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
        
        ***********************************************************************
        *                                                                     *
        *                           SCANTOOL 0.1.1                            *
        *                                                                     *
        *                        Bitcoin Core 25.0.0                          *
        *                         192.168.1.99:9999                           *
        *                                                                     *
        *                            Web Access:                              *
        *                   http://192.168.1.77:8080/web/                     *
        *                                                                     *
        *                         REST API Example:                           *
        *  curl -X GET http://192.168.1.77:8080/rest/v1/current_block_height  *
        *                                                                     *
        ***********************************************************************


4. View the web interface in a browser.

        http://172.17.0.2/web/

5. Send a REST request from the command line.

        $ curl -X GET http://172.17.0.2/rest/v1/current_block_height
        {"Current_block_height":803131}

## Build

To build the executable, simply run the following command.

        $ go build ./scantool.go

## Documentation

### [Settings](/docs/app-settings.md)

### [Web Application](/docs/screen-shots.md)

### REST API

- [Block](/docs/rest-api/v1/block.md)
- [Transaction](/docs/rest-api/v1/tx.md)
- [Output Types](/docs/rest-api/v1/output_types.md)
- [Previous Output](/docs/rest-api/v1/previous_output.md)
- [Current Block Height](/docs/rest-api/v1/current_block_height.md)
- [Serialize Script Usage](/docs/rest-api/v1/serialized_script_usage.md)
- [Blockchain Analysis/Research](/docs/rest-api/v1/blockchain_analysis.md)

