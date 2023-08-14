# Bitcoin Script Analytics Project

[![AGPL licensed](https://img.shields.io/badge/license-AGPL-blue.svg)](https://github.com/btc-script-explorer/explorer/blob/master/LICENSE)

### Features

The application provides 2 main services.

- Web Site (Block/Tx Explorer)
  - Displays serialized scripts (redeem scripts, witness scripts and tap scripts) which are displayed only in hex by most block explorers.
  - Displays Spend Types of inputs in addition to Output Types.
  - Allows the user to view script fields and segregated witness fields as hex, text or data types.
  - Displays both blocks and transactions. (Address searches are not currently supported.)
  - Can be configured to serve web pages on any network interface and port.

- REST API
  - Returns data in JSON format, some of which is not provided by ordinary bitcoin nodes.

### Motivation

##### Serialized Scripts

When people discuss Transaction Types in the bicoin blockchain, they are usually referring to Output Types.
However, outputs are relatively uninteresting from an analytical perspective. The only thing an output does is store a value and reveal a method, or a set of possible methods, (but not the actual data)
required to redeem that value.

Inputs are the workhorses of the bitcoin blockchain. They do the heavy lifting required to transfer funds.
Many modern transactions are redeemed using serialized scripts. Visualizing these scripts is the best way to understand how they work, but most block explorers display them only as hex fields, the same way
they would display a signature or a public key. When we examine serialized scripts, we can see that there are a few more standard input types than there are output types.

As serialized scripts themselves have become standardized, we now have what could be called a standard within a standard, a concept that first appeared with the wrapped P2SH redemption methods
described in BIP 143, where the redeem script used to redeem a standard P2SH output was itself a standard P2WSH output script. Among the redeem scripts that redeem legacy (non-segwit) P2SH outputs
and all witness scripts (which redeem P2SH-P2WSH and P2WSH outputs), more than 90% are standard Multisig transactions.
Ordinals (the newest standard serialized script type) account for nearly 99% of all tap scripts.

In order to analyze how the bitcoin system works, and the ways that transation types are being used, it requires a system that can both identify script types and display them.

In this discussion, we use the term Spend Type to describe the various standard methods that inputs use to redeem funds.
Each Spend Type can redeem exactly one Output Type, but some Output Types are redeemable by multiple Spend Types.
Bitcoin Core recognizes 7 standard Output Types. The table below shows the 10 standard Spend Types that can be used to redeem the standard Output Types.
the required contents of the input script and segregated witness are included in the table.
The Output Types are listed by the names assigned to them by Bitcoin Core. The Spend Types are listed by their commonly-used "P2" names.
Since these are all standard methods for redeeming funds, the data in both the inputs and the outputs must be exactly as shown in the table below, otherwise one or both will be considered non-standard.

![Spend Types](/assets/images/spend-type-table.jpg)

##### Hex, Text & Data Types

Another important feature missing from most block explorers is the ability to easily change the way that script and segregated witness fields are interpretted and displayed.
A field in a script could be an op code, a signature, a public key, a hash, a text message, part of a binary file or some piece of data that is not easily identifiable.
Providing a way to view these fields as hex or text, or have the system attempt to determine what types of data they are, is a feature that most block explorers do not have,
but it is very useful for anyone interested in analyzing script usage as well as anyone who simply wants to learn how the system works.

##### Custom Projects

The REST API makes it easy to create individualized research projects which might focus on analysis of specific spend types, output types, script types, opcode usage or anything else.
Such a program would, in most cases, be simple to write and would make it easy to put large of amounts of data into a database where it could be analyzed more thoroughly.
The REST API could also be used as a back end for custom user interfaces.

### Screen Shots

### Usage

##### Requirements

1. Access to a bitcoin node that has transaction indexing enabled.

##### Download

##### Quick Start (Bitcoin Core)

Obviously, the following ip addresses, ports, username and password should be replaced by the ones used in your specific setup. Do not use the ones shown here.

1. Bitcoin Core Config Settings

        txindex=1
        rpcbind=192.168.1.99:9999
        rpcallowip=192.168.1.77
        rpcuser=btc_node_username
        rpcpassword=btc_node_password

2. Explorer Config Settings

        bitcoin-core-addr=192.168.1.99
        bitcoin-core-port=9999
        bitcoin-core-username=btc_node_username
        bitcoin-core-password=btc_node_password
        url=192.168.1.77
        port=8080

3. Run Explorer

        $ ./explorer --config-file=./explorer.conf 
        
        ************************************************
        *      Node: 192.168.1.99:9999 (Bitcoin Core)  *
        *  Explorer: 192.168.1.77:8080                 *
        ************************************************

4. View the Web User Interface in a Browser

        http://127.0.0.1:8080

4. Send a REST Request from the Command Line

        $ curl -X GET http://127.0.0.1:8080/rest/v1/current_block_height
        {"Current_block_height":803131}

### Building


