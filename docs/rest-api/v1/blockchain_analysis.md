# Blockchain Analysis

Client applications can easily be created in any programming language.
Such an application could be used to gather data over a period of time or a specific range of blocks, or even the entire history of the blockchain.

As a test case, two programs were written in C++, one to analyze the types and contents of ordinals, and the other to analyze multisig transactions that use serialized scripts.
The multisig program was written in 121 lines of code. That includes comments and blank lines.

## Ordinals Test Project

The ordinals program was written in C++, in barely more than 100 lines of code not including comments and whitespace.
The data gathered was written to a PostgreSQL database.
For the test, 392 arbitrarily chosen blocks were analyzed. They were all between block 777000 (February 2023) and block 800019 (July 2023).
A total of 587171 ordinals were found, averaging about 1497 ordinals per block during a peak period of ordinal creation.

The ordinals were divided into 4 categories:
- Standard ordinals, such as the BRC-20 standard
- Ordinals that encode binary files, including certain text file types.
- Ordinals defined by only a text string
- Ordinals that did not fall into any of the above categories. A total of 181 of these were discovered.

### Standard Ordinals

Standard ordinals accounted for slightly more than 93.5% of the ordinals in our sample.
All of them were type text/plain except for 2.96% of them which were application/json.

It appears as though there are several different applications that use very similar but slightly different formats for creating ordinals.
The standard ordinals provide a JSON object with metadata about the ordinal. The "p" field indicates which standard is being used. The "op" field is the operation being performed.

Here are the different standards found in this test and the number of ordinals found for each standard.

Standard | Count | %
---|---:|---:
brc-20 | 542265 | 98.74
orc-20 | 2525 | 0.46
sns | 2408 | 0.44
orc-cash | 652 | 0.12
brc-721 | 458 | 0.08
grc-721 | 293 | 0.05
brc20-s | 155 | 0.03
nft-brc-721 | 128 | 0.02
orc-721 | 102 | 0.02
brc-20c | 102 | 0.02
grc-20 | 73 | 0.01
Ordinals | 11 | 0
drc-20 | 10 | 0
grc-137 | 5 | 0
erc-20 | 5 | 0
.bitter | 3 | 0
orcns | 2 | 0
gen-brc-721 | 2 | 0
ons | 2 | 0
urc-20 | 1 | 0
Brc-20 | 1 | 0
Others-20 | 1 | 0
src-20 | 1 | 0
bitclub | 1 | 0

When we add the operations performed by each standard ordinal type, we see that the vase majority are mint operations.

Standard | Operation | Count | %
:---:|:---:|---:|---:
brc-20 | mint | 527007 | 95.96
brc-20 | transfer | 14671 | 2.67
orc-20 | mint | 2431 | 0.44
sns | reg | 2407 | 0.44
orc-cash | mint | 647 | 0.12
brc-20 | deploy | 587 | 0.11
brc-721 | mint | 456 | 0.08
grc-721 | mint | 293 | 0.05
nft-brc-721 | mint | 128 | 0.02
brc20-s | deposit | 105 | 0.02
brc-20c | mint | 102 | 0.02
orc-721 | mint | 94 | 0.02
orc-20 | send | 76 | 0.01
grc-20 | mint | 73 | 0.01
brc20-s | mint | 18 | 0.00
brc20-s | withdraw | 15 | 0.00
orc-20 | deploy | 14 | 0.00
brc20-s | deploy | 12 | 0.00
Ordinals | mint | 11 | 0.00
drc-20 | mint | 10 | 0.00
orc-721 | deploy | 8 | 0.00
erc-20 | mint | 5 | 0.00
brc20-s | transfer | 5 | 0.00
orc-cash | deploy | 4 | 0.00
grc-137 | register | 3 | 0.00
gen-brc-721 | mint | 2 | 0.00
grc-137 | deploy | 2 | 0.00
brc-721 | deploy | 2 | 0.00
ons | post | 2 | 0.00
orc-20 | cancel | 2 | 0.00
orc-20 | upgrade | 2 | 0.00
orcns | reg | 2 | 0.00
.bitter | post | 2 | 0.00
orc-cash | upgrade | 1 | 0.00
src-20 | mint | 1 | 0.00
bitclub | clubreg | 1 | 0.00
Others-20 | Mint | 1 | 0.00
sns | ns |      1 | 0.00
urc-20 | transfer | 1 | 0.00
Brc-20 | transfer | 1 | 0.00
.bitter | reply | 1 | 0.00



## Binary Files in Ordinals

There were 6187 binary files (including javascript, css and markdown files) encoded among the ordinals analyzed. They ranged in file size from 35 bytes to 396484 bytes.
By far the most common file types found were images. The content of the images was not examined.
The number of each file type found is shown here.

File Type | Count | %
---|---:|---:
image/png | 3944 | 63.75
image/webp | 803 | 12.98
image/jpeg | 529 | 8.55
image/svg+xml | 468 | 7.56
image/avif | 298 | 4.82
image/gif | 121 | 1.96
text/javascript | 13 | 0.21
application/x-gzip | 4 | 0.06
application/octet-stream | 3 | 0.05
text/css | 2 | 0.03
text/markdown | 1 | 0.02
audio/mpeg | 1 | 0.02

## Text Strings in Ordinals

A total of 41597 ordinals were encoded as simple text strings. Of these, approximately 2440 were HTML. The content of the HTML was not examined.
There were 1703 that began with the @ character which appeared to be online handles of some sort.

