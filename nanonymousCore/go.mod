module nanonymousCore

go 1.18

require nanoKeyManager v1.0.0

require (
	filippo.io/edwards25519 v1.0.0-rc.1 // indirect
	golang.org/x/crypto v0.0.0-20220321153916-2c7772ba3064 // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
)

replace nanoKeyManager => ./nanoKeyManager
