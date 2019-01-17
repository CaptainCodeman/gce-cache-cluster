package main

import (
	"context"

	"github.com/golang/groupcache"
)

func NewSourceCache(storage Storage, size int64) *groupcache.Group {
	fill := func(ctx groupcache.Context, key string, dest groupcache.Sink) error {
		data, err := storage.GetBlob(context.TODO(), key)
		if err != nil {
			logger.Debugf("err %v", err)
			return err
		}

		return dest.SetBytes(data)
	}
	return groupcache.NewGroup("source", size<<20, groupcache.GetterFunc(fill))
}
