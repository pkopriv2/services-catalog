package http

import (
	"fmt"
	"os"
	"testing"

	http "github.com/pkopriv2/golang-sdk/http/server"
	"github.com/pkopriv2/golang-sdk/lang/context"
	"github.com/pkopriv2/golang-sdk/lang/enc"
	"github.com/pkopriv2/golang-sdk/lang/errs"
	"github.com/pkopriv2/golang-sdk/lang/net"
	"github.com/pkopriv2/golang-sdk/lang/sql"
	"github.com/pkopriv2/services-catalog/core"
	sqlsvc "github.com/pkopriv2/services-catalog/sql"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	ctx := context.NewContext(os.Stdout, context.Info)
	defer ctx.Close()

	db, e := sql.NewSqlLiteDialer().Embed(ctx)
	if !assert.Nil(t, e) {
		return
	}

	store, err := sqlsvc.NewSqlStore(db, sql.NewSchemaRegistry("TEST"))
	if !assert.Nil(t, err) {
		return
	}

	server, err := http.Serve(ctx,
		http.Build(ServiceHandlers),
		http.WithListener(net.NewTCP4Network(), ":0"),
		http.WithDependency(StorageKey, store),
		http.WithMiddleware(http.TimerMiddleware),
		http.WithMiddleware(http.RouteMiddleware))
	if !assert.Nil(t, err) {
		return
	}

	transport := NewClient(server.Connect(), enc.Json)

	svc := core.NewService("name", "desc")
	if !t.Run("SaveService", func(t *testing.T) {
		svc, err = transport.SaveService(svc)
		assert.Nil(t, err)
	}) {
		return
	}

	if !t.Run("SaveService_Updated", func(t *testing.T) {
		svc, err = transport.SaveService(svc.Increment().SetDesc("desc2"))
		assert.Nil(t, err)
	}) {
		return
	}

	if !t.Run("SaveService_Conflict", func(t *testing.T) {
		_, err = transport.SaveService(svc)
		assert.True(t, errs.Is(err, core.ErrConflict))
	}) {
		return
	}

	// Run through various save version scenarios.
	v := core.NewVersion(svc.Id, "version1")
	if !t.Run("SaveVersion", func(t *testing.T) {
		v, err = transport.SaveVersion(v)
		assert.Nil(t, err)
	}) {
		return
	}
	if !t.Run("SaveVersion_NoService", func(t *testing.T) {
		_, err = transport.SaveVersion(core.NewVersion(uuid.NewV1(), "version1"))
		assert.True(t, errs.Is(err, core.ErrNoService))
	}) {
		return
	}
	if !t.Run("SaveVersion_Conflict", func(t *testing.T) {
		_, err = transport.SaveVersion(v)
		assert.True(t, errs.Is(err, core.ErrConflict))
	}) {
		return
	}

	// Run through some query tests.
	if !t.Run("ListServices_All", func(t *testing.T) {
		catalog, err := transport.ListServices(core.EmptyFilter, core.NewPage())
		if !assert.Nil(t, err) {
			return
		}

		assert.Equal(t, 1, len(catalog.Services))
		assert.Equal(t, svc, catalog.Services[0])
		assert.Equal(t, []core.Version{v}, catalog.Versions[svc.Id])
	}) {
		return
	}

	if !t.Run("ListServices_FilterByName_None", func(t *testing.T) {
		catalog, err := transport.ListServices(
			core.NewFilter(
				core.FilterByName("noexist")),
			core.NewPage())
		if !assert.Nil(t, err) {
			return
		}

		assert.Equal(t, 0, len(catalog.Services))
	}) {
		return
	}

	if !t.Run("ListServices_FilterByName_Equal", func(t *testing.T) {
		catalog, err := transport.ListServices(
			core.NewFilter(
				core.FilterByName(svc.Name)),
			core.NewPage())
		if !assert.Nil(t, err) {
			return
		}

		assert.Equal(t, 1, len(catalog.Services))
		assert.Equal(t, svc, catalog.Services[0])
	}) {
		return
	}

	if !t.Run("ListServices_FilterByName_Contains", func(t *testing.T) {
		catalog, err := transport.ListServices(
			core.NewFilter(
				core.FilterByName("nam")),
			core.NewPage())
		if !assert.Nil(t, err) {
			return
		}

		assert.Equal(t, 1, len(catalog.Services))
		assert.Equal(t, svc, catalog.Services[0])
	}) {
		return
	}

	if !t.Run("ListServices_FilterByDesc_None", func(t *testing.T) {
		catalog, err := transport.ListServices(
			core.NewFilter(
				core.FilterByDesc("noexist")),
			core.NewPage())
		if !assert.Nil(t, err) {
			return
		}

		assert.Equal(t, 0, len(catalog.Services))
	}) {
		return
	}

	if !t.Run("ListServices_FilterByDesc_Equal", func(t *testing.T) {
		catalog, err := transport.ListServices(
			core.NewFilter(
				core.FilterByDesc(svc.Desc)),
			core.NewPage())
		if !assert.Nil(t, err) {
			return
		}

		assert.Equal(t, 1, len(catalog.Services))
		assert.Equal(t, svc, catalog.Services[0])
	}) {
		return
	}

	if !t.Run("ListServices_FilterByDesc_Contains", func(t *testing.T) {
		catalog, err := transport.ListServices(
			core.NewFilter(
				core.FilterByDesc("desc")),
			core.NewPage())
		if !assert.Nil(t, err) {
			return
		}

		assert.Equal(t, 1, len(catalog.Services))
		assert.Equal(t, svc, catalog.Services[0])
	}) {
		return
	}

	if !t.Run("ListServices_Multiple", func(t *testing.T) {
		svc2 := core.NewService("name2", "desc2")
		svc3 := core.NewService("name3", "desc3")
		if !assert.Nil(t, store.SaveService(svc2)) {
			return
		}
		if !assert.Nil(t, store.SaveService(svc3)) {
			return
		}
		v21 := core.NewVersion(svc2.Id, "version21")
		v22 := core.NewVersion(svc2.Id, "version22")
		if !assert.Nil(t, store.SaveVersion(v21)) {
			return
		}
		if !assert.Nil(t, store.SaveVersion(v22)) {
			return
		}

		catalog, err := transport.ListServices(core.EmptyFilter, core.NewPage())
		if !assert.Nil(t, err) {
			return
		}

		assert.Equal(t, 3, len(catalog.Services))
		assert.Equal(t, svc, catalog.Services[0])
		assert.Equal(t, svc2, catalog.Services[1])
		assert.Equal(t, svc3, catalog.Services[2])
		fmt.Println(enc.Json.MustEncodeString(catalog))
	}) {
		return
	}
}
