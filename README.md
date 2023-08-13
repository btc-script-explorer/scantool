### Bitcoin Script Analytics Project

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

