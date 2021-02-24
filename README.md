# OCM Backend

This is a bespoke backend for fetching blockchain data from the [One Click Miner](https://github.com/vertcoin-project/one-click-miner-vnext).

This is work in progress and lacking documentation

## Installation

Either compile the go binary or build the Docker container using the Dockerfile

Then, set up a PostgreSQL server and create a database. In that database run `database.sql` from the repository to create the table structure

## Running

Set the following environment variables (either in a `docker-compose` if you're using the Docker container, or just plain `export` if you're running the go binary):

| Setting | Meaning | Example |
|---------|---------|----------|
| `PGSQL_CONNECTION` | The connection string for the database | `postgres://user:pass@server/database?sslmode=disable` |
| `RPCHOST` | The host where vertcoind listens for RPC commands | `mainnet:5888` |
| `RPCUSER` | The rpc user to authenticate with | `rpc` |
| `RPCPASS` | The password to authenticate with | `4rlkjjkasdlfkj2` |
| `OCM_BACKEND_STARTHEIGHT` | The height at which to begin indexing. Will start at genesis if omitted. Outputs received before startheight will not be shown as part of the balance | `1300000` |
| `OCM_BACKEND_APIONLY` | Set this to 1 for backup API endpoints. The software does not support running the indexer twice against the same database, but setting this to 1 allows a backup endpoint for serving API requests and loadbalance it that way | `1` |

# Donations

If you want to reward this work you can donate some coins here:

* Vertcoin : `VnGYRfD65XFab4gE3feJwWVUuaxK7EgFPK` 
* Bitcoin  : `37E56moXYRE2RU9ddXnhu8gTLk7Rqq45eh`