# Application Settings

It is recommended to use a config file, although all of the settings shown here can be used on the command line by prepending -- to the front of them.

When using a config file, the **--config-file** command line parameter is required. It can be a full path or a relative path to the config file.

## Bitcoin Core Settings

***

### bitcoin-core-addr

The IP address used to connect to the JSON RPC API of Bitcoin Core.
This should be the IP address part of one of the **rpcbind** Bitcoin Core settings.

**Default**: 127.0.0.1<br>
**Required** if using Bitcoin Core

***

### bitcoin-core-port

The port number used to connect to the JSON RPC API of Bitcoin Core.
This should be the port part of one of the **rpcbind** Bitcoin Core settings.

**Default**: 8332<br>
**Required** if using Bitcoin Core

***

### bitcoin-core-username

The username used to connect to the JSON RPC API of Bitcoin Core.
This should be the **rpcuser** Bitcoin Core setting.

**Default**: None<br>
**Required** if using Bitcoin Core

***

### bitcoin-core-password

The password used to connect to the JSON RPC API of Bitcoin Core.
This should be the **rpcpassword** Bitcoin Core setting.

**Default**: None<br>
**Required** if using Bitcoin Core

## HTTP Server Settings

### addr

The IP address of the website.

**Default**: 127.0.0.1<br>
**Required** if **no-web** is not set

***

### port

The port number of the website.

**Default**: 8080<br>
**Required** if **no-web** is not set

***

### no-web

Turns off the web server. It does not take a value. If it is present, the web server will be turned off. Otherwise, it will be turned on.

**Default**: off<br>
**Optional**

