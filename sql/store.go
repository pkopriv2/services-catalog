package sql

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pkopriv2/golang-sdk/lang/errs"
	"github.com/pkopriv2/golang-sdk/lang/sql"
	"github.com/pkopriv2/services-catalog/core"
	uuid "github.com/satori/go.uuid"
)

// The schema mirrors the core type definitions.  Essentially,
// there are two tables, services and versions.  When listing
// the catalog, versions are joined to services and a composite
// type is returned.
var (
	SchemaService = sql.NewSchema("service", 0).
		WithStruct(core.Service{}).
		WithIndices(
			sql.NewUniqueIndex("idx_service_uniq", "id", "version"),
			sql.NewIndex("idx_service_name", "name"),  // index for searching on name
			sql.NewIndex("idx_service_desc", "desc")). // index for searching on description
		Build()
)

var (
	SchemaVersion = sql.NewSchema("version", 0).
		WithStruct(core.Version{}).
		WithIndices(
			sql.NewUniqueIndex("idx_version_uniq", "service_id", "name")).
		Build()
)

var emptyId = uuid.UUID{}

type SqlServiceStore struct {
	db sql.Driver
}

func NewSqlStore(db sql.Driver, schemas sql.SchemaRegistry) (ret core.Storage, err error) {
	if err = sql.InitSchemas(db, schemas,
		SchemaService,
		SchemaVersion); err != nil {
		return
	}

	ret = &SqlServiceStore{db}
	return
}

func (s *SqlServiceStore) SaveService(service core.Service) (err error) {
	if service.Id == emptyId {
		err = errors.Wrapf(core.ErrState, "Id must not be empty")
		return
	}
	if service.Name == "" {
		err = errors.Wrapf(core.ErrState, "Name must not be empty")
		return
	}

	defer func() {
		switch {
		case errs.Is(err, sql.ErrSqliteUnique): // not portable
			err = core.ErrConflict
		case errs.Is(err, sql.ErrNone):
			err = errors.Wrapf(core.ErrNoService, "No such service [%v]", service.Id)
		}
	}()

	// If this is the first version, just go ahead and insert.  If a concurrent
	// insert is happening, the unique constraint will prevent one from winning.
	if service.Version <= 0 {
		return s.db.Do(sql.Exec(SchemaService.Insert(service)))
	}

	return s.db.Do(
		sql.ExpectOne(
			SchemaService.SelectAs("s").
				Where("s.id = ?", service.Id).
				Where("s.version = ?", service.Version-1)).
			ThenExec(SchemaService.Insert(service)))
}

func (s *SqlServiceStore) SaveVersion(version core.Version) (err error) {
	if version.ServiceId == emptyId {
		err = errors.Wrapf(core.ErrState, "ServiceId must not be empty")
		return
	}
	if version.Name == "" {
		err = errors.Wrapf(core.ErrState, "Name must not be empty")
		return
	}

	defer func() {
		switch {
		case errs.Is(err, sql.ErrSqliteUnique): // not portable
			err = core.ErrConflict
		case errs.Is(err, sql.ErrNone):
			err = errors.Wrapf(core.ErrNoService, "No such service [%v]", version.ServiceId)
		}
	}()

	return s.db.Do(
		sql.ExpectOne(
			SchemaService.SelectAs("s").
				Where("s.id = ?", version.ServiceId).
				Where(latestService("s"))).
			ThenExec(SchemaVersion.Insert(version)))
}

func (s *SqlServiceStore) ListServices(filter core.Filter, page core.Page) (ret core.Catalog, err error) {

	// Need to validate the order field since this will be part of the query
	// itself (not one of the bindings). This is to protect against sql injection.
	switch page.OrderBy {
	default:
		err = errors.Wrapf(core.ErrState, "Invalid order by field [%v]. Must be one of [name, desc, updated]", page.OrderBy)
		return
	case
		"name",
		"desc",
		"updated":
	}

	// In order to implement proper pagination, we need to use an inner
	// select which is not handled well by the sql query builder.  Just
	// use raw sql to construct this query.
	query := `
select
	s.id,
	s.name,
	s.desc,
	s.version,
	s.updated,
	v.service_id,
	v.name,
	v.created
from
	(
		select
			*
		from
			service as s
		where
			not exists (
				select
					1
				from
					service as o
				where
					o.id = s.id
					and o.version > s.version
			)
			%v
		order by s.%v, s.id limit %v offset %v
	) as s
left join version as v on v.service_id = s.id
order by s.%v, s.id, v.created
`

	// Add filter arguments
	inner, binds := "", []interface{}{}
	if filter.NameContains != nil {
		inner += "and s.name like ?"
		binds = append(binds, "%"+*filter.NameContains+"%")
	}

	if filter.DescContains != nil {
		inner += "and s.desc like ?"
		binds = append(binds, "%"+*filter.DescContains+"%")
	}

	if filter.ServiceId != nil {
		inner += "and s.id = ?"
		binds = append(binds, *filter.ServiceId)
	}

	// Finally, compile the real query
	query = fmt.Sprintf(query,
		inner,
		page.OrderBy,
		page.Limit,
		page.Offset,
		page.OrderBy)

	type row struct {
		core.Service
		core.Version
	}
	var results []row

	err = s.db.Do(
		sql.Scan(
			sql.Raw(query, binds...),
			sql.Slice(&results, sql.MultiStruct)))
	if err != nil {
		return
	}

	services := make([]core.Service, 0, len(results))
	versions := make(map[uuid.UUID][]core.Version)
	for _, r := range results {
		// add the service if we haven't seen it before
		if len(services) == 0 || services[len(services)-1].Id != r.Service.Id {
			services = append(services, r.Service)
		}

		// add the version if one exists.
		if r.Version.ServiceId != emptyId {
			versions[r.Version.ServiceId] =
				append(versions[r.Version.ServiceId], r.Version)
		}
	}

	ret = core.Catalog{
		Offset:   page.Offset,
		Limit:    page.Limit,
		Versions: versions,
		Services: services}
	return
}

func latestService(alias string) string {
	return fmt.Sprintf(`
		not exists (
			select
				1
			from
				service as o
			where
				o.id = %v.id
				and o.version > %v.version
		)`, alias, alias)
}
