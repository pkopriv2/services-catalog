package http

import (
	"time"

	"github.com/pkg/errors"
	"github.com/pkopriv2/golang-sdk/http/headers"
	http "github.com/pkopriv2/golang-sdk/http/server"
	"github.com/pkopriv2/golang-sdk/lang/enc"
	"github.com/pkopriv2/golang-sdk/lang/errs"
	"github.com/pkopriv2/golang-sdk/lang/mime"
	"github.com/pkopriv2/services-catalog/core"
	uuid "github.com/satori/go.uuid"
)

const (
	StorageKey = "storage.services"
)

// Uses the dependency injector to retrieve the storage implementation
func getStorage(env http.Environment) (ret core.Storage) {
	env.Assign(StorageKey, &ret)
	return
}

// Register all the server handlers
func ServiceHandlers(svc *http.Service) {
	var emptyId = uuid.UUID{}

	svc.Register(http.Put("/v1/services"),
		func(env http.Environment, req http.Request) (ret http.Response) {
			logger, storage := env.Logger(), getStorage(env)

			var svc core.Service
			if err := http.RequireStruct(req, enc.DefaultRegistry, &svc); err != nil {
				ret = http.BadRequest(err)
				return
			}

			// Do some basic validation.  Would need to better understand
			// product requirements to constrain these fields further.
			// For now, just make sure that none of the required fields
			// are empty.
			if ret = http.First(
				http.AssertTrue(svc.Name != "", "Invalid epoch"),
				http.AssertTrue(svc.Desc != "", "Invalid description"),
				http.AssertTrue(svc.Version >= 0, "Invalid version"),
			); ret != nil {
				return
			}

			// Basic support for handling multiple encoding types.
			accept := mime.Json
			if _, err := http.ParseHeader(req, headers.Accept, http.String, &accept); err != nil {
				ret = http.BadRequest(err)
				return
			}

			ok, enc := enc.DefaultRegistry.FindByMime(accept)
			if !ok {
				ret = http.BadRequest(errors.Errorf("Invalid accept type: %v", accept)) // TODO: Is this the right response type?
				return
			}

			logger.Debug("Adding service [name=%v,version=%v]", svc.Name, svc.Version)
			if svc.Version <= 0 {
				svc = svc.SetId(uuid.NewV1())
			} else {
				svc = svc.Update()
			}

			if err := storage.SaveService(svc); err != nil {
				if errs.Is(err, core.ErrConflict) {
					ret = http.Conflict(err)
					return
				}

				ret = http.Panic(err)
				return
			}

			ret = http.Ok(enc, svc)
			return
		})

	svc.Register(http.Put("/v1/versions"),
		func(env http.Environment, req http.Request) (ret http.Response) {
			logger, storage := env.Logger(), getStorage(env)

			var v core.Version
			if err := http.RequireStruct(req, enc.DefaultRegistry, &v); err != nil {
				ret = http.BadRequest(err)
				return
			}

			// Do some basic validation.  Would need to better understand
			// product requirements to constrain these fields further.
			// For now, just make sure that none of the required fields
			// are empty.
			if ret = http.First(
				http.AssertTrue(v.ServiceId != emptyId, "Invalid service id"),
				http.AssertTrue(v.Name != "", "Invalid name"),
			); ret != nil {
				return
			}

			// Basic support for handling multiple encodings.
			accept := mime.Json
			if _, err := http.ParseHeader(req, headers.Accept, http.String, &accept); err != nil {
				ret = http.BadRequest(err)
				return
			}

			ok, enc := enc.DefaultRegistry.FindByMime(accept)
			if !ok {
				ret = http.BadRequest(errors.Errorf("Invalid accept type: %v", accept)) // TODO: Is this the right response type?
				return
			}

			v = v.SetCreated(time.Now().UTC())
			logger.Debug("Adding version [service=%v,name=%v]", v.ServiceId, v.Name)

			if err := storage.SaveVersion(v); err != nil {
				ret = http.Panic(err)
				return
			}

			ret = http.Ok(enc, v)
			return
		})

	// Considered making this a POST /v1/services_list that included a request body.
	// Instead just made it a simple GET and encoding the various request elements
	// in the query parameters
	svc.Register(http.Get("/v1/services"),
		func(env http.Environment, req http.Request) (ret http.Response) {
			storage := getStorage(env)

			filter := core.NewFilter()
			if err := http.ParseQueryParams(req,
				http.Param("name", http.String, &filter.NameContains),
				http.Param("desc", http.String, &filter.DescContains),
				http.Param("id", http.UUID, &filter.ServiceId),
			); err != nil {
				ret = http.BadRequest(err)
				return
			}

			page := core.NewPage()
			if err := http.ParseQueryParams(req,
				http.Param("offset", http.Uint64, &page.Offset),
				http.Param("limit", http.Uint64, &page.Limit),
				http.Param("order", http.String, &page.OrderBy),
			); err != nil {
				ret = http.BadRequest(err)
				return
			}

			// Do some basic validation.  Would need to better understand
			// product requirements to constrain these fields further.
			// For now, just make sure that none of the required fields
			// are empty.
			if ret = http.AssertTrue(page.Limit <= 1024, "Invalid limit. Must be <= 1024"); ret != nil {
				return
			}

			// Basic support for handling multiple encodings
			accept := mime.Json
			if _, err := http.ParseHeader(req, headers.Accept, http.String, &accept); err != nil {
				ret = http.BadRequest(err)
				return
			}

			ok, enc := enc.DefaultRegistry.FindByMime(accept)
			if !ok {
				ret = http.BadRequest(errors.Errorf("Invalid accept type: %v", accept)) // TODO: Is this the right response type?
				return
			}

			catalog, err := storage.ListServices(filter, page)
			if err != nil {
				ret = http.Panic(err)
				return
			}

			ret = http.Ok(enc, catalog)
			return
		})
}
