module nanonymousCore

go 1.18

require (
	github.com/jackc/pgconn v1.11.0
	github.com/jackc/pgerrcode v0.0.0-20220416144525-469b46aa5efa
	github.com/jackc/pgx/v4 v4.15.0
	github.com/shopspring/decimal v1.2.0
	golang.org/x/crypto v0.0.0-20220321153916-2c7772ba3064
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2
	nanoKeyManager v1.0.0
	nanoTypes v1.0.0
)

require (
	filippo.io/edwards25519 v1.0.0-rc.1 // indirect
	github.com/hectorchu/gonano v0.1.17 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.2.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.10.0 // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
	golang.org/x/text v0.3.6 // indirect
)

replace nanoKeyManager => ./nanoKeyManager
replace nanoTypes => ./nanoTypes
