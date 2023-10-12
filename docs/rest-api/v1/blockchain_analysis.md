# Blockchain Analysis

Client applications can easily be created in any programming language.
Such an application could be used to gather data about the blockchain over a period of time or a specific range of blocks, or even the entire history of the blockchain.

As an example, two programs were written in C++, one to analyze the types and contents of ordinals, and the other to analyze multisig transactions that use serialized scripts.

## Ordinals Example Project

The ordinals analysis program was written in C++. (Similar programs could be written in many different languages.) The data gathered were written to a PostgreSQL database where they could be analyzed further.

For the test, 392 arbitrarily chosen blocks were analyzed. They were all between block 777000 (February 2023) and block 800019 (July 2023).
A total of 587171 ordinals were found, averaging about 1497 ordinals per block during a peak period of ordinal creation.

The ordinals were divided into 4 categories:
- Standard ordinals, such as the BRC-20 standard
- Ordinals that encode binary files, including certain text file types.
- Ordinals defined by only a text string
- Ordinals that did not fall into any of the above categories. A total of 181 of these were discovered. For this example, they were removed from the data set.

### Standard Ordinals

Standard ordinals accounted for slightly more than 93.5% of the ordinals in our sample.
They were type text/plain except for 2.96% of them which were application/json.

It appears as though there are several different applications that use very similar but slightly different formats for creating ordinals.
The standard ordinals provide a JSON object with metadata about the ordinal. The "p" field indicates which protocol is being used. The "op" field is the operation being performed.

Here are the different protocols found in this test and the number of ordinals found for each protocol.

Protocol | Count | %
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

Here are the operations found in this test.

Operation | Count | %
---|---:|---:
mint | 531278 | 96.74
transfer | 14678 | 2.67
reg | 2409 | 0.44
deploy | 629 | 0.11
deposit | 105 | 0.02
send | 76 | 0.01
withdraw | 15 | 0.00
post | 4 | 0.00
register | 3 | 0.00
upgrade | 3 | 0.00
cancel | 2 | 0.00
clubreg | 1 | 0.00
ns | 1 | 0.00
reply | 1 | 0.00
Mint | 1 | 0.00

### Binary Files in Ordinals

There were 6187 binary files (including javascript, css and markdown files) encoded among the ordinals analyzed. They ranged in file size from 35 bytes to 396484 bytes.
The vast majority of files found were images. The content of the images was not examined.
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
audio/mpeg | 1 | 0.02
text/markdown | 1 | 0.02

### Text Strings in Ordinals

A total of 41597 ordinals were encoded as simple text strings. Of these, approximately 2440 were HTML. The content of the HTML was not examined.
There were 1703 that began with the @ character which appeared to be online handles of some sort.

## Multisig Example Project

The multisig analysis program was written in C++. (Similar programs could be written in many different languages.) The data gathered were written to a PostgreSQL database where they could be analyzed further.

For the test, all blocks between and including blocks 798000 and 798299 were analyzed, a total of 300 blocks.
The test identified 120817 multisig serialized scripts, averaging about 403 per block.

The results are shown below.

Multisig Type | Count | %
---|---:|---:
2 of 3 | 72847 | 60.30
2 of 2 | 19705 | 16.31
1 of 1 | 18622 | 15.41
3 of 4 | 6227 | 5.15
2 of 4 | 3023 | 2.50
3 of 5 | 159 | 0.13
2 of 5 | 90 | 0.07
5 of 7 | 51 | 0.04
1 of 3 | 26 | 0.02
1 of 2 | 25 | 0.02
3 of 3 | 14 | 0.01
1 of 4 | 10 | 0.00
3 of 7 | 4 | 0.00
4 of 6 | 4 | 0.00
3 of 6 | 4 | 0.00
2 of 6 | 3 | 0.00
4 of 4 | 2 | 0.00
4 of 5 | 1 | 0.00

