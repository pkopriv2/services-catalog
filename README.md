# Services Catalog

This project implements a very simple services catalog.  It supports
pluggable storage and transport layers, but comes shipped with a SQL
storage implementation and an HTTP transport implementation.

## Assumptions

* Multiple services with same name allowed
* Services may have [0,n] versions
* No constraints on versions types (e.g. semver)

## Dependencies

This project makes heavy use of `github.com/pkopriv2/golang-sdk`.  These 
are a set of golang libraries that I have been using to prototype and 
and build services.  It provides a number of useful libraries but it
was brought in for its support in constructing HTTP servers/clients 
and its SQL libraries, e.g.:

* github.com/pkopriv2/golang-sdk/lang/sql
* github.com/pkopriv2/golang-sdk/lang/http/server
* github.com/pkopriv2/golang-sdk/lang/http/client

NOTE: This dependency makes use of sqlite3, which uses CGO under the
covers. 

## Getting Started

This project can run a standalone server and client.  To start the server,
open a shell and run:

```
go run main.go start
```

To seed the server with some data, run:
```
go run main.go load
```

To begin listing the catalog, run:
```
go run main.go list
```

## Project Organization

```
* cli  - CLI commands
* core - Core data types and libraries (see core/api.go)
* http - HTTP client & server
* sql  - SQL storage implementation
* main.go - main entrypoint
```

## Design Overview


