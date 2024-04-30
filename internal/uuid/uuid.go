package uuid

import (
	"github.com/google/uuid"
	"github.com/nohns/eventale/internal"
)

type uuidImpl struct {
	id uuid.UUID
}

func (impl uuidImpl) String() string {
	return impl.String()
}

func (impl uuidImpl) Bytes() []byte {
	return impl.id[:]
}

type ider func() (internal.ID, error)

func (f ider) GenID() (internal.ID, error) {
	return f()
}

func Gen() (internal.ID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	return uuidImpl{id: id}, nil
}

var IDer internal.IDer = ider(func() (internal.ID, error) {
	return Gen()
})
