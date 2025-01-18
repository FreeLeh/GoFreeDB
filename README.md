# GoFreeDB
<br />

<div align="center">
	<picture>
		<source media="(prefers-color-scheme: dark)" srcset="docs/img/logo_dark.png">
		<img width=200 src="docs/img/logo_light.png">
	</picture>
	<h3><i>Ship Faster with Google Sheets as a Database!</i></h3>
</div>

<p align="center">
	<code>GoFreeDB</code> is a Golang library that provides common and simple database abstractions on top of Google Sheets.
</p>

<br />

<div align="center">

  ![Unit Test](https://github.com/FreeLeh/GoFreeDB/actions/workflows/unit_test.yml/badge.svg)
  ![Integration Test](https://github.com/FreeLeh/GoFreeDB/actions/workflows/full_test.yml/badge.svg)
![Coverage](https://img.shields.io/badge/Coverage-82.2%25-brightgreen)
  [![Go Report Card](https://goreportcard.com/badge/github.com/FreeLeh/GoFreeDB)](https://goreportcard.com/report/github.com/FreeLeh/GoFreeDB)
  [![Go Reference](https://pkg.go.dev/badge/github.com/FreeLeh/GoFreeDB.svg)](https://pkg.go.dev/github.com/FreeLeh/GoFreeDB)

</div>

## Features

1. Provide a straightforward **key-value** and **row based database** interfaces on top of Google Sheets.
2. Serve your data **without any server setup** (by leveraging Google Sheets infrastructure).
3. Support **flexible enough query language** to perform various data queries.
4. **Manually manipulate data** via the familiar Google Sheets UI (no admin page required).

> For more details, please read [our analysis](https://github.com/FreeLeh/docs/blob/main/freedb/alternatives.md#why-should-you-choose-freedb)
> on other alternatives and how it compares with `FreeDB`.

## Table of Contents

* [Protocols](#protocols)
* [Getting Started](#getting-started)
  * [Installation](#installation)
  * [Pre-requisites](#pre-requisites)
* [Row Store](#row-store)
  * [Querying Rows](#querying-rows)
  * [Counting Rows](#counting-rows)
  * [Inserting Rows](#inserting-rows)
  * [Updating Rows](#updating-rows)
  * [Deleting Rows](#deleting-rows)
  * [Struct Field to Column Mapping](#struct-field-to-column-mapping)
* [KV Store](#kv-store)
  * [Get Value](#get-value)
  * [Set Key](#set-key)
  * [Delete Key](#delete-key)
  * [Supported Modes](#supported-modes)
* [KV Store V2](#kv-store-v2)
  * [Get Value](#get-value-v2)
  * [Set Key](#set-key-v2)
  * [Delete Key](#delete-key-v2)
  * [Supported Modes](#supported-modes-v2)

## Protocols

Clients are strongly encouraged to read through the **[protocols document](https://github.com/FreeLeh/docs/blob/main/freedb/protocols.md)** to see how things work
under the hood and **the limitations**.

## Getting Started

### Installation

```
go get github.com/FreeLeh/GoFreeDB
```

### Pre-requisites

1. Obtain a Google [OAuth2](https://github.com/FreeLeh/docs/blob/main/google/authentication.md#oauth2-flow) or [Service Account](https://github.com/FreeLeh/docs/blob/main/google/authentication.md#service-account-flow) credentials.
2. Prepare a Google Sheets spreadsheet where the data will be stored.

## Row Store

Let's assume each row in the table is represented by the `Person` struct.

```go
type Person struct {
	Name string `db:"name"`
	Age  int    `db:"age"`
}
```

Please read the [struct field to column mapping](#struct-field-to-column-mapping) section
to understand the purpose of the `db` struct field tag.

```go
import (
	"github.com/FreeLeh/GoFreeDB"
	"github.com/FreeLeh/GoFreeDB/google/auth"
)

// If using Google Service Account.
auth, err := auth.NewServiceFromFile(
	"<path_to_service_account_json>", 
	freedb.FreeDBGoogleAuthScopes, 
	auth.ServiceConfig{},
)

// If using Google OAuth2 Flow.
auth, err := auth.NewOAuth2FromFile(
	"<path_to_client_secret_json>", 
	"<path_to_cached_credentials_json>", 
	freedb.FreeDBGoogleAuthScopes, 
	auth.OAuth2Config{},
)

store := freedb.NewGoogleSheetsRowStore(
	auth, 
	"<spreadsheet_id>", 
	"<sheet_name>", 
	freedb.GoogleSheetRowStoreConfig{Columns: []string{"name", "age"}},
)
defer store.Close(context.Background())
```

### Querying Rows

```go
// Output variable
var output []Person

// Select all columns for all rows
err := store.
	Select(&output).
	Exec(context.Background())

// Select a few columns for all rows (non-selected struct fields will have default value)
err := store.
	Select(&output, "name").
	Exec(context.Background())

// Select rows with conditions
err := store.
	Select(&output).
	Where("name = ? OR age >= ?", "freedb", 10).
	Exec(context.Background())

// Select rows with sorting/order by
ordering := []freedb.ColumnOrderBy{
	{Column: "name", OrderBy: freedb.OrderByAsc},
	{Column: "age", OrderBy: freedb.OrderByDesc},
}
err := store.
	Select(&output).
	OrderBy(ordering).
	Exec(context.Background())

// Select rows with offset and limit
err := store.
	Select(&output).
	Offset(10).
	Limit(20).
	Exec(context.Background())
```

### Counting Rows

```go
// Count all rows
count, err := store.
	Count().
	Exec(context.Background())

// Count rows with conditions
count, err := store.
	Count().
	Where("name = ? OR age >= ?", "freedb", 10).
	Exec(context.Background())
```

### Inserting Rows

```go
err := store.Insert(
	Person{Name: "no_pointer", Age: 10}, 
	&Person{Name: "with_pointer", Age: 20},
).Exec(context.Background())
```

### Updating Rows

```go
colToUpdate := make(map[string]interface{})
colToUpdate["name"] = "new_name"
colToUpdate["age"] = 12

// Update all rows
err := store.
	Update(colToUpdate).
	Exec(context.Background())

// Update rows with conditions
err := store.
	Update(colToUpdate).
	Where("name = ? OR age >= ?", "freedb", 10).
	Exec(context.Background())
```

### Deleting Rows

```go
// Delete all rows
err := store.
	Delete().
	Exec(context.Background())

// Delete rows with conditions
err := store.
	Delete().
	Where("name = ? OR age >= ?", "freedb", 10).
	Exec(context.Background())
```

### Struct Field to Column Mapping

The struct field tag `db` can be used for defining the mapping between the struct field and the column name.
This works just like the `json` tag from [`encoding/json`](https://pkg.go.dev/encoding/json).

Without `db` tag, the library will use the field name directly (case-sensitive).

```go
// This will map to the exact column name of "Name" and "Age".
type NoTagPerson struct {
	Name string
	Age  int
}

// This will map to the exact column name of "name" and "age" 
type WithTagPerson struct {
	Name string  `db:"name"`
	Age  int     `db:"age"`
}
```

## KV Store

> Please use `KV Store V2` as much as possible, especially if you are creating a new storage.

```go
import (
	"github.com/FreeLeh/GoFreeDB"
	"github.com/FreeLeh/GoFreeDB/google/auth"
)

// If using Google Service Account.
auth, err := auth.NewServiceFromFile(
	"<path_to_service_account_json>", 
	freedb.FreeDBGoogleAuthScopes, 
	auth.ServiceConfig{},
)

// If using Google OAuth2 Flow.
auth, err := auth.NewOAuth2FromFile(
	"<path_to_client_secret_json>", 
	"<path_to_cached_credentials_json>", 
	freedb.FreeDBGoogleAuthScopes, 
	auth.OAuth2Config{},
)

kv := freedb.NewGoogleSheetKVStore(
	auth, 
	"<spreadsheet_id>", 
	"<sheet_name>", 
	freedb.GoogleSheetKVStoreConfig{Mode: freedb.KVSetModeAppendOnly},
)
defer kv.Close(context.Background())
```

### Get Value

If the key is not found, `freedb.ErrKeyNotFound` will be returned.

```go
value, err := kv.Get(context.Background(), "k1")
```

### Set Key

```go
err := kv.Set(context.Background(), "k1", []byte("some_value"))
```

### Delete Key

```go
err := kv.Delete(context.Background(), "k1")
```

### Supported Modes

> For more details on how the two modes are different, please read the [protocol document](https://github.com/FreeLeh/docs/blob/main/freedb/protocols.md).

There are 2 different modes supported:

1. Default mode.
2. Append only mode.

```go
// Default mode
kv := freedb.NewGoogleSheetKVStore(
	auth,
	"<spreadsheet_id>",
	"<sheet_name>",
	freedb.GoogleSheetKVStoreConfig{Mode: freedb.KVModeDefault},
)

// Append only mode
kv := freedb.NewGoogleSheetKVStore(
	auth,
	"<spreadsheet_id>",
	"<sheet_name>",
	freedb.GoogleSheetKVStoreConfig{Mode: freedb.KVModeAppendOnly},
)
```

## KV Store V2

The KV Store V2 is implemented internally using the row store.

> The original `KV Store` was created using more complicated formulas, making it less maintainable.
> You can still use the original `KV Store` implementation, but we strongly suggest using this new `KV Store V2`.

You cannot use an existing sheet based on `KV Store` with `KV Store V2` as the sheet structure is different. 
- If you want to convert an existing sheet, just add an `_rid` column and insert the first key-value row with `1`
  and increase it by 1 until the last row.
- Remove the timestamp column as `KV Store V2` does not depend on it anymore. 

```go
import (
	"github.com/FreeLeh/GoFreeDB"
	"github.com/FreeLeh/GoFreeDB/google/auth"
)

// If using Google Service Account.
auth, err := auth.NewServiceFromFile(
	"<path_to_service_account_json>", 
	freedb.FreeDBGoogleAuthScopes, 
	auth.ServiceConfig{},
)

// If using Google OAuth2 Flow.
auth, err := auth.NewOAuth2FromFile(
	"<path_to_client_secret_json>", 
	"<path_to_cached_credentials_json>", 
	freedb.FreeDBGoogleAuthScopes, 
	auth.OAuth2Config{},
)

kv := freedb.NewGoogleSheetKVStoreV2(
	auth, 
	"<spreadsheet_id>", 
	"<sheet_name>", 
	freedb.GoogleSheetKVStoreV2Config{Mode: freedb.KVSetModeAppendOnly},
)
defer kv.Close(context.Background())
```

### Get Value V2

If the key is not found, `freedb.ErrKeyNotFound` will be returned.

```go
value, err := kv.Get(context.Background(), "k1")
```

### Set Key V2

```go
err := kv.Set(context.Background(), "k1", []byte("some_value"))
```

### Delete Key V2

```go
err := kv.Delete(context.Background(), "k1")
```

### Supported Modes V2

> For more details on how the two modes are different, please read the [protocol document](https://github.com/FreeLeh/docs/blob/main/freedb/protocols.md).

There are 2 different modes supported:

1. Default mode.
2. Append only mode.

```go
// Default mode
kv := freedb.NewGoogleSheetKVStoreV2(
	auth,
	"<spreadsheet_id>",
	"<sheet_name>",
	freedb.GoogleSheetKVStoreV2Config{Mode: freedb.KVModeDefault},
)

// Append only mode
kv := freedb.NewGoogleSheetKVStoreV2(
	auth,
	"<spreadsheet_id>",
	"<sheet_name>",
	freedb.GoogleSheetKVStoreV2Config{Mode: freedb.KVModeAppendOnly},
)
```

## License

This project is [MIT licensed](https://github.com/FreeLeh/GoFreeDB/blob/main/LICENSE).


Lark:
- Spreadsheet Token = Spreadsheet ID
- Sheet ID -> not sure whether this is actually needed, but this is basically sheet name.
- 