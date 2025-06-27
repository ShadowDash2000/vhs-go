package tests

import (
	"github.com/pocketbase/pocketbase/core"
	"vhs/internal/vhs"
)

type VideoBaseMock struct {
	vhs.VideoBase
}

func NewVideoBaseMockFromRecord(record *core.Record) *VideoBaseMock {
	v := &VideoBaseMock{}
	v.SetProxyRecord(record)

	return v
}

func (v *VideoBaseMock) Save() error {
	return PocketBase.Save(v)
}

func (v *VideoBaseMock) Refresh() error {
	return nil
}
