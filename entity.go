package go_Weather_ITUR

import (
	"sync/atomic"
)

var (
	idInc uint64
)

type BasicEntity struct {
	// Entity ID.
	id       uint64
}

type Identifier interface {
	ID() uint64
}

func (e BasicEntity) ID() uint64 {
	return e.id
}

func NewBasic() BasicEntity {
	return BasicEntity{id: atomic.AddUint64(&idInc, 1)}
}

func (e *BasicEntity) GetBasicEntity() *BasicEntity {
	return e
}

type BasicFace interface {
	GetBasicEntity() *BasicEntity
}