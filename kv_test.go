package freedb

import (
	"context"
	"fmt"

	"github.com/FreeLeh/GoFreeDB/google/auth"
)

func ExampleGoogleSheetKVStore() {
	googleAuth, err := auth.NewServiceFromFile(
		"<path_to_file>",
		GoogleAuthScopes,
		auth.ServiceConfig{},
	)
	if err != nil {
		panic(err)
	}

	store := NewGoogleSheetKVStore(
		googleAuth,
		"spreadsheet_id",
		"sheet_name",
		GoogleSheetKVStoreConfig{
			Mode: KVModeDefault,
		},
	)

	val, err := store.Get(context.Background(), "key1")
	if err != nil {
		panic(err)
	}
	fmt.Println("get key", val)

	err = store.Set(context.Background(), "key1", []byte("value1"))
	if err != nil {
		panic(err)
	}

	err = store.Delete(context.Background(), "key1")
	if err != nil {
		panic(err)
	}
}

func ExampleGoogleSheetKVStoreV2() {
	googleAuth, err := auth.NewServiceFromFile(
		"<path_to_file>",
		GoogleAuthScopes,
		auth.ServiceConfig{},
	)
	if err != nil {
		panic(err)
	}

	storeV2 := NewGoogleSheetKVStoreV2(
		googleAuth,
		"spreadsheet_id",
		"sheet_name",
		GoogleSheetKVStoreV2Config{
			Mode: KVModeDefault,
		},
	)

	val, err := storeV2.Get(context.Background(), "key1")
	if err != nil {
		panic(err)
	}
	fmt.Println("get key", val)

	err = storeV2.Set(context.Background(), "key1", []byte("value1"))
	if err != nil {
		panic(err)
	}

	err = storeV2.Delete(context.Background(), "key1")
	if err != nil {
		panic(err)
	}
}
