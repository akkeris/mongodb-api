# Broker for MongoDB in Akkeris

## Synopis

Runs an http server with REST to provision MongoDB database instance in a shared MongoDB server

## Details

Listens on port 4040

* GET /v1/mongodb/plans
* POST /v1/mongodb/instance/ JSON body with plan and billingcode
* GET /v1/mongodb/instance/:name
* DELETE /v1/mongodb/instance/:name
* GET /v1/mongodb/:name
* GET /v1/mongodb/url/:name
* GET /v1/mongodb

## Runtime Environment Variables

* VAULT_ADDR
* VAULT_TOKEN
* NAME_PREFIX
* MONGODB_SECRET=
* MONGODB_API_RUNTIME
* PORT

## Build

* make dep
* make test
* make docker 
