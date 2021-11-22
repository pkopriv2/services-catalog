package sql

import (
	"os"
	"testing"

	"github.com/pkopriv2/golang-sdk/lang/context"
	"github.com/pkopriv2/golang-sdk/lang/errs"
	"github.com/pkopriv2/golang-sdk/lang/sql"
	"github.com/pkopriv2/services-catalog/core"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestServiceStore(t *testing.T) {
	ctx := context.NewContext(os.Stdout, context.Debug)
	defer ctx.Close()

	db, e := sql.NewSqlLiteDialer().Embed(ctx)
	if !assert.Nil(t, e) {
		return
	}

	store, err := NewSqlStore(db, sql.NewSchemaRegistry("TEST"))
	if !assert.Nil(t, err) {
		return
	}

	// Run through the various save service methods.
	svc := core.NewService("name", "description")
	if !t.Run("SaveService", func(t *testing.T) {
		assert.Nil(t, store.SaveService(svc))
	}) {
		return
	}

	if !t.Run("SaveService_Updated", func(t *testing.T) {
		svc = svc.Increment().SetDesc("description2")
		assert.Nil(t, store.SaveService(svc))
	}) {
		return
	}

	if !t.Run("SaveService_Conflict", func(t *testing.T) {
		assert.Equal(t, core.ErrConflict, store.SaveService(svc))
	}) {
		return
	}

	// Run through various save version scenarios.
	v := core.NewVersion(svc.Id, "version1")
	if !t.Run("SaveVersion", func(t *testing.T) {
		assert.Nil(t, store.SaveVersion(v))
	}) {
		return
	}
	if !t.Run("SaveVersion_NoService", func(t *testing.T) {
		assert.True(t, errs.Is(store.SaveVersion(core.NewVersion(uuid.NewV1(), "version1")), core.ErrNoService))
	}) {
		return
	}
	if !t.Run("SaveVersion_Conflict", func(t *testing.T) {
		assert.Equal(t, core.ErrConflict, store.SaveVersion(v))
	}) {
		return
	}

	// Run through some query tests.
	if !t.Run("ListServices_All", func(t *testing.T) {
		catalog, err := store.ListServices(core.EmptyFilter, core.NewPage())
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
		catalog, err := store.ListServices(
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
		catalog, err := store.ListServices(
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
		catalog, err := store.ListServices(
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
		catalog, err := store.ListServices(
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
		catalog, err := store.ListServices(
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
		catalog, err := store.ListServices(
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
		svc2 := core.NewService("name2", "description2")
		svc3 := core.NewService("name3", "description3")
		if !assert.Nil(t, store.SaveService(svc2)) {
			return
		}
		if !assert.Nil(t, store.SaveService(svc3)) {
			return
		}

		catalog, err := store.ListServices(core.EmptyFilter, core.NewPage())
		if !assert.Nil(t, err) {
			return
		}

		assert.Equal(t, 3, len(catalog.Services))
		assert.Equal(t, svc, catalog.Services[0])
		assert.Equal(t, svc2, catalog.Services[1])
		assert.Equal(t, svc3, catalog.Services[2])
	}) {
		return
	}
}
