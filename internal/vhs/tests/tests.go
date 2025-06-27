package tests

import (
	"github.com/pocketbase/pocketbase/tests"
	"os"
	"vhs/pkg/collections"

	_ "vhs/migrations"
)

const pbBasePath = "pb_data_test"

var (
	PocketBase  *tests.TestApp
	Collections *collections.Collections
)

func init() {
	var err error
	err = os.MkdirAll(pbBasePath, 0755)
	if err != nil {
		panic(err)
	}

	PocketBase, err = tests.NewTestApp(pbBasePath)
	if err != nil {
		panic(err)
	}

	Collections = collections.NewCollections(PocketBase)
}
