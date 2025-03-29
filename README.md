# Nanonymous
Front and backend code for nanonymous.cc (Servers down indefinately)

## Disclaimer
This repo was never meant as a community project. It is merely an archive of my project in the hopes that it might help someone. As such, there are quite a few oddities and hurdles to getting it set up. The code is provided as-is for posterity; I will not be helping you troubleshoot.

There ARE two smaller repos that I was using in Nanonymous that are intended to be used by any nano go progamers. They are the [nanoKeyManager](https://github.com/Tumbler/nanoKeyManager) and [nanoyTypes](https://github.com/Tumbler/nanoTypes).

## Breakdown
  * Frontend
  * Backend (core)
  * Scritps

### Frontend
There's not too much here other than your typcial HTML/javascript/css and some libraries that I used and did not make myself. The only really technically interesting part is calculating the fees in [transaction.js](https://github.com/Tumbler/nanonymous/blob/main/nanonymousFrontEnd/script/transaction.js). This essentially had to work backwards from the logic in the core and had to perfectly match so people could make precise payments even with the fee.

### Backend
This is the meat and potatoes of the project. Written in Go and Postgres.

#### Setup
  * `git clone git@github.com:Tumbler/nanonymous.git`
  * `cd nanonymous/nanonymousCore`
  * `go get github.com/c-sto/encembed` (This is an embedding library that I used to embed and encrypt the configuruation data)
  * Add "embed.txt" to the directory.
  * `go generate` (This must be done every time there is a change to `embed.txt`)

At this point the core should be ready to run, with the caveat that you must have it talking to operating node or it will hang.
