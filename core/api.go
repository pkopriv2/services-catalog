package core

import (
	"errors"
	"time"

	uuid "github.com/satori/go.uuid"
)

// Standard error types.  Returned by transport and storage apis
var (
	ErrState     = errors.New("Core:ErrState")
	ErrConflict  = errors.New("Core:ErrConflict")
	ErrNoService = errors.New("Core:ErrNoService")
)

// This defines the core service data type.
//
// Multiple services with the same name are allowed. In fact, they may describe
// completely different services. As such, we're going to use a simple unique
// identifier to address the service. There is an additional versioning column
// that allows services to be updated, and which is also used for concurrency
// control.  The unique key for a service is then (id, version).
type Service struct {
	Id      uuid.UUID `json:"id,omitempty"`
	Name    string    `json:"name"`
	Desc    string    `json:"desc"`
	Version int       `json:"version,omitempty"`
	Updated time.Time `json:"updated,omitempty"`
}

func NewService(name, desc string) Service {
	return Service{
		Id:      uuid.NewV1(),
		Name:    name,
		Desc:    desc,
		Version: 0,
		Updated: time.Now().UTC(),
	}
}

func (s Service) Update(fns ...func(*Service)) (ret Service) {
	ret = s
	for _, fn := range fns {
		fn(&ret)
	}
	ret.Updated = time.Now().UTC()
	return
}

// Increments the version of the service.  Exposed to consumers
// for greater control.
func (s Service) Increment() (ret Service) {
	return s.Update(func(s *Service) {
		s.Version = s.Version + 1
	})
}

// Set the id of the service.  This is typically done in the business logic of the
// service. Essentially, we're not going to allow consumers to generate their own ids
func (s Service) SetId(id uuid.UUID) (ret Service) {
	return s.Update(func(s *Service) {
		s.Id = id
	})
}

// Set the description of the service.
func (s Service) SetDesc(desc string) (ret Service) {
	return s.Update(func(s *Service) {
		s.Desc = desc
	})
}

// This describes a service version.  Services may contain many versions.  They may
// also contain none.  This wasn't explicitly discussed so definitely taking a liberty
// here.  If this feature is wrong, then some of the following APIs may be a little
// off.
//
// Because we're not dictating any constraints on the names of versions (e.g. semantic
// versions), we can't provide any native sorting techniques. Added a created timestamp
// to allow for sorting at the presentation layer. For all intents and purposes versions
// are immutable.
//
type Version struct {
	ServiceId uuid.UUID `json:"service_id"`
	Name      string    `json:"name"`
	Created   time.Time `json:"created"`
}

func NewVersion(serviceId uuid.UUID, name string) Version {
	return Version{
		ServiceId: serviceId,
		Name:      name,
		Created:   time.Now().UTC(),
	}
}

func (v Version) Update(fns ...func(*Version)) (ret Version) {
	ret = v
	for _, fn := range fns {
		fn(&ret)
	}
	return
}

func (v Version) SetCreated(time time.Time) Version {
	return v.Update(func(v *Version) {
		v.Created = time
	})
}

// A simple aggregate type that represents a page of services and their
// associated versions.
type Catalog struct {
	Services map[uuid.UUID]Service   `json:"services"`
	Versions map[uuid.UUID][]Version `json:"versions"`
	Offset   uint64                  `json:"offset"`
	Limit    uint64                  `json:"limit"`
}

// This is the primary storage interface.  This project will come shipped with a SQL
// implementation but others may be swapped at deploy time
type Storage interface {

	// Saves a service.  Concurrency control is expected to operate on the version
	// field, such that only a single version of a given service is allowed.
	// In order to update a service, a client must increment the previous version.
	SaveService(Service) error

	// Adds a version.  Implementations must verify that the associated service exists.
	SaveVersion(Version) error

	// List services. May provide filtering and paging options.  An empty filter will allow
	// essentially be equivalent to a list all.
	ListServices(Filter, Page) (Catalog, error)
}

// This is the primary client interface.  This project will come shipped with an HTTP client transport.
type Transport interface {

	// Adds/updates a service.  Multiple services of the same name
	// are allowed.
	SaveService(Service) (Service, error)

	// Adds a version. Multiple versions of the same name and service are not allowed.
	// The corresponding service must exist.
	SaveVersion(Version) (Version, error)

	// List services. May provide filtering and paging options.  An empty filter will
	// be equivalent to "list all". Implementations may implement additional constraints
	// on the input paging options.
	ListServices(Filter, Page) (Catalog, error)
}
