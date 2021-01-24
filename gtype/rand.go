package gtype

import (
	"sync"
	"time"
)

type Rand interface {
	New() uint64
}

func NewRand(index uint64) Rand {
	return &randNumber{
		index: index,
	}
}

type randNumber struct {
	sync.Mutex

	id    uint64
	max   uint64
	index uint64
}

func (s *randNumber) New() uint64 {
	s.Lock()
	defer s.Unlock()

	s.id++
	if s.id > s.max {
		now := time.Now()
		sec := now.Unix() // 10
		idStart := s.index*1000000000000000 + uint64(sec)*100000

		s.id = idStart + 1
		s.max = idStart + 99999
	}

	return s.id
}
