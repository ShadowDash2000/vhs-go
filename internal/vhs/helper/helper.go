package helper

import "github.com/pocketbase/pocketbase/core"

type Fn1[T comparable, V any] func(V) T
type Fn2[T any] func(T) error

func VerifyFields[T comparable, V any](list []V, cmp T, fn1 Fn1[T, V], fn2 Fn2[V]) error {
	for _, item := range list {
		v := fn1(item)
		if v != cmp {
			continue
		}
		err := fn2(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateRecordFromOther(r1 *core.Record, r2 *core.Record, fields ...string) {
	for _, field := range fields {
		r1.Set(field, r2.Get(field))
	}
}
