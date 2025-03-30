# Nanonymous
<img src="nanonymousFrontEnd/images/No_words_logo.png" width="250"><br>
Front and backend code for nanonymous.cc (Servers down indefinately)

Nanonymous was a tool to help increase [nano](nano.org) anonymity by breaking the direct chain of sender -> receiver on the blockchain. It did so by maintaining many wallets to send from alongside a hashed blacklist to ensure future transactions never related addresses that were part of previous transactions.

## Disclaimer
This repo was never meant as a community project. It is merely an archive of my project in the hopes that it might help someone. As such, there are quite a few oddities and hurdles to getting it set up. The code is provided as-is for posterity; I will not be helping you troubleshoot.

There ARE two smaller repos that I was using in Nanonymous that are intended to be used by any nano go developers. They are the [nanoKeyManager](https://github.com/Tumbler/nanoKeyManager) and [nanoTypes](https://github.com/Tumbler/nanoTypes).

## Breakdown
  * [Frontend](#frontend)
  * [Backend](#backend) (core)
  * [Scripts](#scripts)

## Frontend
There's not too much here other than your typcial HTML/javascript/css and some libraries that I used and did not make myself. The only really technically interesting part is calculating the fees in [transaction.js](https://github.com/Tumbler/nanonymous/blob/main/nanonymousFrontEnd/script/transaction.js). This essentially had to work backwards from the logic in the core and had to perfectly match so people could make precise payments even with the fee.

## Backend
This is the meat and potatoes of the project. Usually referred to as the "Core" in the code. Written in Go and Postgres.

### Setup
  * Postgres must be setup and configured on the machine. (Then use [databaseSetup](https://github.com/Tumbler/nanonymous/blob/main/scripts/databaseSetup.sql))
  * `git clone git@github.com:Tumbler/nanonymous.git`
  * `cd nanonymous/nanonymousCore`
  * `go get github.com/c-sto/encembed` (This is an embedding library that I used to embed and encrypt the configuruation data)
  * Add "embed.txt" to the directory.
  * `go generate` (This must be done every time there is a change to `embed.txt`)

At this point the core should be ready to run, with the caveat that you must have it talking to an operating node or it will hang. <br><br>

### Behavior
The default behavior of the core is to run in listen mode where it is simply waiting for calls from the front end to respond to. There is also a CLI that can be accessed with the launch option `-c`. <br><br>

The CLI has three basic modes:
 1. A CLI Wallet for sending/receiving funds from the addresses in the database
 2. An RCP client for the connected node
 3. A very basic database viewer

### Launch options
 * `-c` Launch the CLI instead of the listener
 * `-s` Go through database and check if there are any unrecieved funds and receive them. (Scans only last two seeds unless `-a` is also specified)
 * `-r` Compiles a report about the last weeks transactions and emails them to the specified email in embed.txt
 * `-v` Prints version information
 * `-h` Prings help message
 * `-beta` Runs with no fees
 * If the final argument is a number (no dash) the output verbosity is changed to that number (1-10)

## Scripts
Some small tools to help with server maintenance.

 * `backupdatabase.sh` Backups the postgres database in case of corruption. (Designed to be called from a cron job)
 * `cleardatabase.sql` Purges the database **CAUTION IRREVERSABLE**
 * `coreRetireSeed.sh` Tells the current running instance to stop generating addresses from the current seed. (Desinged to be called from a cron job)
 * `coreSafeExit.sh` Tells the current running instance to finsh whatever it's doing and exit. (This prevents inturrupting an active tranaction and losing customre funds)
 * `databaseSetup.sql` Initializes the database with the necessary tables to run. Won't override already exsisting tables. (Use `cleardatabase.sql` first if you need to wipe and start over)
 * `generateTLSCert.sh` Generates a self signed TLS certificate and key that can be used between your own servers.
 * `generateTLSCertNoPass.sh` Same as above but doesn't prompt for a password (In case you need to do this programmatically)
 * `resetTestDatabase.sql` Resets the test database to a known (non-empty) state so that all our unit tests can have known success cases.
