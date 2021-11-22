# Services Catalog

This project implements a very simple services catalog. It supports
pluggable storage and transport layers, but comes shipped with a SQL
storage implementation and an HTTP transport implementation.

## Assumptions

* Multiple services with same name allowed
* Services may have [0,n] versions
* No constraints on versions types (e.g. semver)
* Versions are immutable
* Services are mutable

## Getting Started

This project can run a standalone server and client. To start the server,
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

You can filter results by name, description or id:
```
go run main.go list --name "example"
go run main.go list --desc "example"
go run main.go list --id 269f1872-4be9-11ec-8acb-9801a796f7a7
```

You can view the versions for services by supplying a `-v` flag, e.g.:
```
go run main.go list -v
```

Paging options can be supplied:
```
go run main.go list --offset 10 -n 10" --orderBy name
```

## Project Organization

```
* cli - Command line command definitions
* core - Core data types and libraries (see core/api.go) <-- This is the best place to start
* http - HTTP client & server
* sql - SQL storage implementation
* main.go - Main entrypoint
```

## Design Considerations

### Data/Storage Layer

I began the design thinking about the data model. I originally went down the road of trying 
a single table design. Each row would encode both a service and a version. To list all the
versions, simply collect all the instances of a row with a given service id. To get a service
given an identifier, select the latest of all rows with the given id.

However, this proved deficient when attempting to accommodate deletes of versions (didn't
actually implement deletes, just wanted a design that would grow with future functionality).
I probably could have forged ahead and encoded add/delete operations to the versions 
themselves but this seemed needlessly complex.

Instead, I decided to split the data into two objects: 1) Service and 2) Version, where the
versions point back to their services. Services are allowed to evolve (maybe descriptions 
are updated or names changed) so I included a simple versioning scheme that allows us to
both prevent concurrent updates and maintain a history of services. Versions on the other
handle are immutable once they hit the storage layer. No additional versioning mechanisms
required since only a single instance of a version will be available. Concurrent/simultaneous
inserts of the same version are prevented through the use of a unique index on version 
(service\_id, name). The complete schema can be seen here:

* https://github.com/pkopriv2/services-catalog/blob/main/sql/store.go#L17-L33

Once the design was settled, the actual implementation of the storage layer wasn't too 
difficult. Saving both services and versions are straight forward. The only additional
constraints to consider were ensuring that services' version numbers were incremented
properly and that the corresponding service actually existed before inserting a new
version. 

Listing services was definitely the most interesting aspect. I briefly considered
two separate list methods, one for services and one for versions (selected by a slice 
of service ids) but to perform a join across two separate transactions could produce
inconsistent results. For example, a service may have been deleted before the corresponding
versions could be gathered, resulting in an inconsistent view. Instead, I decided to 
perform the join in a single query. You can view the query here:

* https://github.com/pkopriv2/services-catalog/blob/main/sql/store.go#L129-L159

And lastly, the storage is exposed by a technology-agnostic API. New implementations
can be injected at runtime. You can view the API here:

* https://github.com/pkopriv2/services-catalog/blob/main/core/api.go#L121-L136

### REST API

The REST API very closely resembles the storage layer in terms of transactional
boundaries. Users can add/update services and add versions. The only real
decisions were how to encode the method calls and arguments. I landed on the 
following endpoints:

 * PUT /v1/services
 * PUT /v1/versions
 * GET /v1/services?name=<>&desc=<>&id=<>&offset=<>&limit=<>

I was on the fence between PUT vs. POST for the updates, but ultimately landed
on PUT since they encapsulate both update and create semantics for /v1/services.
I could very easily be talked into POST for both of these. The client and server 
implementation can be found here:

* https://github.com/pkopriv2/services-catalog/blob/main/http/client.go
* https://github.com/pkopriv2/services-catalog/blob/main/http/server.go

### Testing

The project does include some minimal automated testing. This occurred
at both the sql and http layers. The provided tests are not exhaustive
but cover a majority of the functionality. They helped flush out the obvious
bugs and greatly sped up development. If this were a bit closer to production
quality, I would definitely beef up the test cases.

## Dependencies

This project makes heavy use of `github.com/pkopriv2/golang-sdk`. These 
are a set of golang libraries that I have been using to prototype and 
and build services. It provides a number of useful libraries but it
was brought in for its support in constructing HTTP servers/clients 
and its SQL libraries, e.g.:

* github.com/pkopriv2/golang-sdk/lang/sql
* github.com/pkopriv2/golang-sdk/lang/http/server
* github.com/pkopriv2/golang-sdk/lang/http/client

NOTE: This dependency makes use of sqlite3, which uses CGO under the
covers. 
