# GoFreeLeh

![Unit Test](https://github.com/FreeLeh/GoFreeLeh/actions/workflows/unit_test.yml/badge.svg)

<div>
    <h2 align="center">
        Excited to start your personal project, but too lazy to setup your server, database, KV store, etc.?
        <br>
        <br>
        <i>We feel you!</i>
    </h2>
</div>

`GoFreeLeh` is a Golang library providing common and familiar interfaces on top of common free services we have access to.

## Why do you need this library?

Our main goal is to make developers who want to **just start their small personal projects so much easier without thinking too much about the setup required to get started**. We can leverage a bunch of well known free services available to us like Google Sheets and Telegram. We want to use these services as our **easy-to-setup and "managed" database or even a message queue**.

`GoFreeLeh` is just the beginning. It is very likely we will explore other languages (e.g. Python, Java, Kotlin, Swift, etc.) to support in the future.

## What kind of interfaces/abstractions can this library provide?

Here are a few things we have developed so far:

1. A simple key-value store on top of Google Sheets.
2. A simple row based database on top of Google Sheets.

There are other ideas we have in our backlog:

1. A simple message queue on top of Google Sheets.
2. A simple message queue on top of Telegram Channels.

We are quite open to knowing any other free services we can leverage on.<br>
Please suggest your ideas in the [issues]() page!

## What can I do with these interfaces/abstractions?

The primary target for this library is **small personal projects with low QPS and no high performance requirement**. A project that is not too complex, only doing simple queries, and only used by the project owner is usually a good candidate.

Here are a few ideas we thought of:

1. **A simple personalised expenses tracker.**
    - A simple mobile app is sufficient. The app is specifically configured and set up just for the author.
    - Mobile app is not distributed via Google Play Store or Apple App Store.
    - The expenses can be **tracked using the simple row based database on top of Google Sheets**.
    - The data can be further manipulated through Google Sheets manually (e.g. summarise it using pivot table).

2. **A simple home automation controller.**
    - You may want to build a simple mobile app controlling your Raspberry Pi.
    - However, you cannot connect to your Raspberry Pi easily (there are tools for it, but it's usually not free).
    - You can make the mobile app publish an event to Google Sheets and let the Raspberry Pi listen to such events and act accordingly.

# Table of Contents

* [Key Value Store](#key-value-store)
    * [Google Sheets Key Value Store](#google-sheets-key-value-store)
        * [Key Value Store Interface](#key-value-store-interface)
        * [Key Value Store Modes](#key-value-store-modes)
            * [Default Mode](#default-mode)
            * [Append Only Mode](#append-only-mode)
* [Row Store](#row-store)
   * [Google Sheets Row Store](#google-sheets-row-store)
       * [Row Store Interface](#row-store-interface)
* [Google Credentials](#google-credentials)
    * [OAuth2 Flow](#oauth2-flow)
    * [Service Account Flow](#service-account-flow)
    * [Custom HTTP Client](#custom-http-client)
* [Limitations](#limitations)
* [Disclaimer](#disclaimer)
* [License](#license)
    

# Key Value Store

## Google Sheets Key Value Store

```go
// If using Google Service Account.
auth, _ := auth.NewServiceFromFile(
    "<path_to_service_account_json>",
    []string{auth.GoogleSheetsReadWrite},
    auth.ServiceConfig{},
)

// If using Google OAuth2 Flow.
auth, err := auth.NewOAuth2FromFile(
    "<path_to_client_secret_json>",
    "<path_to_cached_credentials_json>",
    []string{auth.GoogleSheetsReadWrite},
    auth.OAuth2Config{},
)

// Below are the same regardless of the auth client chosen above.
kv := freeleh.NewGoogleSheetKVStore(
    auth,
    "<spreadsheet_id>",
    "<sheet_name>",
    freeleh.GoogleSheetKVStoreConfig{Mode: freeleh.KVSetModeAppendOnly},
)
defer kv.Close(context.Background())

value, _ := kv.Get(context.Background(), "k1")
_ = kv.Set(context.Background(), "k1", []byte("some_value"))
_ = kv.Delete(context.Background(), "k1")
```

Getting started is very simple (error handling ignored for brevity).
You only need 3 information to get started:

1. A Google credentials (the `auth` variable). Read below for more details how to get this.
2. The Google Sheets `spreadsheet_id` to use as your database.
3. The Google Sheets `sheet_name` to use as your database.

If you want to compare the above concept with a Redis server, the `spreadsheet_id` is the Redis host and port,
while a `sheet_name` is the Redis database that you can select using the [Redis `SELECT` command](https://redis.io/commands/select/).

### Key Value Store Interface

#### `Get(ctx context.Context, key string) ([]byte, error)`

- `Get` tries to retrieve the value associated to the given key.
- If the key exists, this method will return the value.
- Otherwise, `ErrKeyNotFound` will be returned.

#### `Set(ctx context.Context, key string, value []byte) error`

- `Set' performs an upsert operation on the key.
- If the key exists, this method will update the value for that key.
- Otherwise, it will create a new entry and sets the value accordingly.

#### `Delete(ctx context.Context, key string) error`

- `Delete` removes the key from the database.
- If the key exists, this method will remove the key from the database.
- Otherwise, this method will do nothing.

> ### ⚠️ ⚠️ Warning
> Please note that only `[]byte` values are supported at the moment.

### Key Value Store Modes

There are 2 different modes supported:

1. Default mode.
2. Append only mode.

```go
// Default mode
kv := freeleh.NewGoogleSheetKVStore(auth, "<spreadsheet_id>", "<sheet_name>", freeleh.GoogleSheetKVStoreConfig{Mode: freeleh.KVModeDefault})

// Append only mode
kv := freeleh.NewGoogleSheetKVStore(auth, "<spreadsheet_id>", "<sheet_name>", freeleh.GoogleSheetKVStoreConfig{Mode: freeleh.KVModeAppendOnly})
```

#### Default Mode

The default mode works just like a normal key value store. The behaviours are as follows.

##### `Get(ctx context.Context, key string) ([]byte, error)`

- Returns `ErrKeyNotFound` if the key is not in the store.
- Use a simple `VLOOKUP` formula on top of the data table.
- Does not support concurrent operations.

##### `Set(ctx context.Context, key string, value []byte) error`

- If the key is not in the store, `Set` will create a new row and store the key value pair there.
- If the key is in the store, `Set` will update the previous row with the new value and timestamp.
- There are exactly 2 API calls behind the scene: getting the row for the key and creating/updating with the given key value data.
- Does not support concurrent operations.

##### `Delete(ctx context.Context, key string) error`

- If the key is not in the store, `Delete` will not do anything.
- If the key is in the store, `Delete` will remove that row.
- There are up to 2 API calls behind the scene: getting the row for the key and remove the row (if the key exists).
- Does not support concurrent operations.

![Default Mode Screenshot](docs/img/default_mode.png?raw=true)

You can see that each key (the first column) only appears at most once.

Some additional notes to understand the default mode better:

1. Default mode is easier to manage as the concept is very similar to common key value store out there. You can think of it like a normal `map[string]string` in Golang.
2. Default mode is slower for most operations as its `Set` and `Delete` operation need up to 2 API calls.
3. Default mode uses less rows as it updates in place.
4. Default mode does not support concurrent operations.

#### Append Only Mode

The append only mode works by only appending changes to the end of the sheet. The behaviours are as follows.

##### `Get(ctx context.Context, key string) ([]byte, error)`

- Returns `ErrKeyNotFound` if the key is not in the store.
- Use a simple `VLOOKUP` with `SORT` (sort the 3rd column, the timestamp) formula on top of the data table.
- Support concurrent operations as long as the `GoogleSheetKV` instance is not shared between goroutines.

##### `Set(ctx context.Context, key string, value []byte) error`

- `Set` always creates a new row at the bottom of the sheet with the latest value and timestamp.
- There is only 1 API call behind the scene.
- Support concurrent operations as long as the `GoogleSheetKV` instance is not shared between goroutines.

##### `Delete(ctx context.Context, key string) error`

- `Delete` also creates a new row at the bottom of the sheet with a tombstone value and timestamp.
- `Get` will recognise the tombstone value and decide that the key has been deleted.
- There is only 1 API call behind the scene.
- Support concurrent operations as long as the `GoogleSheetKV` instance is not shared between goroutines.

![Append Only Mode Screenshot](docs/img/append_only_mode.png?raw=true)

You can see that a specific key can have multiple rows. The row with the latest timestamp would be seen as the latest value for that specific key.

Some additional notes to understand the append only mode better:

1. Append only mode is faster for most operations as all methods are only calling the API once.
2. Append only mode may use more rows as it does not do any compaction of the old rows (unlike SSTable concept).
3. Append only mode support concurrent operations as long as the `GoogleSheetKV` instance is not shared between goroutines.

# Row Store

## Google Sheets Row Store

```go
// If using Google Service Account.
auth, _ := auth.NewServiceFromFile(
    "<path_to_service_account_json>",
    []string{auth.GoogleSheetsReadWrite},
    auth.ServiceConfig{},
)

// If using Google OAuth2 Flow.
auth, err := auth.NewOAuth2FromFile(
    "<path_to_client_secret_json>",
    "<path_to_cached_credentials_json>",
    []string{auth.GoogleSheetsReadWrite},
    auth.OAuth2Config{},
)

// Below are the same regardless of the auth client chosen above.
store := freeleh.NewGoogleSheetsRowStore(
    auth,
    "<spreadsheet_id>",
    "<sheet_name>",
    freeleh.GoogleSheetRowStoreConfig{Columns: []string{"name", "age"}},
)
defer store.Close(context.Background())

type Person struct {
	Name string
	Age int
}

// Inserts a bunch of rows.
// Note that the here matters, and it should follow the GoogleSheetRowStoreConfig.Columns settings.
_ = store.RawInsert(
    []interface{}{"name1", 10},
    []interface{}{"name2", 11},
    []interface{}{"name3", 12},
).Exec(context.Background())

// Updates the name column for rows with age = 10
_ = store.Update(map[string]interface{}{"name": "name4"}).Where("age = ?", 10).Exec(context.Background())

// Deletes rows with age = 11
_ = store.Delete().Where("age = ?", 11).Exec(context.Background())

// Returns rows with just the name values for rows with name = name4 or age = 12
var results []Person
_ = store.Select(&results, "name").Where("name = ? OR age = ?", "name4", 12).Exec(context.Background())
```

Getting started is very simple (error handling ignored for brevity).
You only need 3 information to get started:

1. A Google credentials (the `auth` variable). Read below for more details how to get this.
2. The Google Sheets `spreadsheet_id` to use as your database.
3. The Google Sheets `sheet_name` to use as your database.
4. A list of strings to define the columns in your database (note that the ordering matters!).

## Row Store Interface

For all the examples in this section, we assume we have a table of 2 columns: name (column A) and age (column B).

> ### ⚠️ ⚠️ Warning
> Please note that the row store implementation does not support any ACID guarantee.
> Concurrency is not a primary consideration and there is no such thing as a "transaction" concept anywhere.
> Each statement may trigger multiple APIs and those API executions are not atomic in nature.

### `Select(output interface{}, columns ...string) *googleSheetSelectStmt`

- `Select` returns a statement to perform the actual select operation. You can think of this operation like the normal SQL select statement (with limitations).
- If `columns` is an empty list, all columns will be returned.
- If a column is not found in the provided list of columns in `GoogleSheetRowStoreConfig.Columns`, that column will be ignored.
- The `output` argument must be a pointer to a slice.
- We are using the [`mapstructure`](https://pkg.go.dev/github.com/mitchellh/mapstructure) package to perform the conversion from a raw `map[string]interface{}` into the `output` argument. Any struct tag provided by `mapstructure` should work as well.

#### `googleSheetSelectStmt`

##### `Where(condition string, args ...interface{}) *googleSheetSelectStmt`

- The values in `condition` string must be replaced using a placeholder.
- The actual values used for each placeholder (ordering matters) are provided via the `args` parameter.
- The purpose of doing this is because we need to replace each column name registered in `GoogleSheetRowStoreConfig.Columns` into the column name in Google Sheet (i.e. `A` for the first column, `B` for the second column, and so on).
- All conditions supported by Google Sheet `QUERY` function are supported by this library. You can read the full information in this [Google Sheets Query docs](https://developers.google.com/chart/interactive/docs/querylanguage#where).
- This function returns a reference to the statement for chaining.

Examples:

```go
// SELECT * WHERE A = "bob" AND B = 12
store.Select(&result).Where("name = ? AND age = ?", "bob", 12)

// SELECT * WHERE A like "b%" OR B >= 10
store.Select(&result).Where("name like ? OR age >= ?", "b%", 10) 
```

##### `OrderBy(colToOrdering map[string]OrderBy) *googleSheetSelectStmt`

- The `colToOrdering` argument decides which column should have what kind of ordering.
- The library provides 2 `OrderBy` constants: `OrderByAsc` and `OrderByDesc`.
- And empty `colToOrdering` map will result in no operation.
- This function will translate into the `ORDER BY` clause as stated in this [Google Sheets Query docs](https://developers.google.com/chart/interactive/docs/querylanguage#order-by).
- This function returns a reference to the statement for chaining.

Examples:

```go
// SELECT * WHERE A = "bob" AND B = 12 ORDER BY A ASC, B DESC
store.Select(&result).Where("name = ? AND age = ?", "bob", 12).OrderBy(map[string]OrderBy{"name": OrderByAsc, "age": OrderByDesc)

// SELECT * ORDER BY A ASC
store.Select(&result).OrderBy(map[string]OrderBy{"name": OrderByAsc})
```

##### `Limit(limit uint64) *googleSheetSelectStmt`

- This function limits the number of returned rows.
- This function will translate into the `LIMIT` clause as stated in this [Google Sheets Query docs](https://developers.google.com/chart/interactive/docs/querylanguage#limit).
- This function returns a reference to the statement for chaining.

Examples:

```go
// SELECT * WHERE A = "bob" AND B = 12 LIMIT 10
store.Select(&result).Where("name = ? AND age = ?", "bob", 12).Limit(10)
```

##### `Offset(offset uint64) *googleSheetSelectStmt`

- This function skips a given number of first rows.
- This function will translate into the `OFFSET` clause as stated in this [Google Sheets Query docs](https://developers.google.com/chart/interactive/docs/querylanguage#offset).
- This function returns a reference to the statement for chaining.

Examples:

```go
// SELECT * WHERE A = "bob" AND B = 12 OFFSET 10
store.Select(&result).Where("name = ? AND age = ?", "bob", 12).Offset(10)
```

##### `Exec(ctx context.Context) error`

- This function will actually execute the `SELECT` statement and inject the resulting rows into the provided `output` argument in the `Select` function.
- There is only one API call involved in this function.

Examples:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
err := store.Select(&result).Where("name = ? AND age = ?", "bob", 12).Exec(ctx)
```
 
### `RawInsert(rows ...[]interface{}) *googleSheetRawInsertStmt`

- `RawInsert` returns a statement to perform the actual insert operation.
- The `rows` argument is a slice of an `interface{}` slice.
- The ordering of the values inside each slice of `interface{}` matters as it is the ordering that this library will use when inserting into the Google Sheet.

> This function is called `RawInsert` because the library is not really concerned with how the values in each row is formed.
> There is also no type checking involved.
> In the future, we are thinking of adding an `Insert` function that will provide a simple type checking mechanism.

#### `googleSheetRawInsertStmt`

##### `Exec(ctx context.Context) error`

- This function will actually execute the `INSERT` statement.
- This works by appending new rows into Google Sheets.
- There is only one API call involved in this function.

### `Update(colToValue map[string]interface{}) *googleSheetUpdateStmt`
 
- `Update` returns a statement to perform the actual update operation.
- The `colToValue` map tells the library which column should be updated to what value.
- Note that the column in `colToValue` must exist in the `GoogleSheetRowStoreConfig.Columns` definition.
- The value is not type checked at the moment.

#### `googleSheetUpdateStmt`

##### `Where(condition string, args ...interface{}) *googleSheetUpdateStmt`

This works exactly the same as the `googleSheetSelectStmt.Where` function. You can refer to the above section for more details.

##### `Exec(ctx context.Context) error`

- This function will actually execute the `UPDATE` statement.
- There are two API calls involved: one for figuring out which rows are affected and another for actually updating the values.
  
### `Delete() *googleSheetDeleteStmt`

- `Delete` returns a statement to perform the actual delete operation.

#### `googleSheetDeleteStmt`

##### `Where(condition string, args ...interface{}) *googleSheetDeleteStmt`

This works exactly the same as the `googleSheetSelectStmt.Where` function. You can refer to the above section for more details.

##### `Exec(ctx context.Context) error`

- This function will actually execute the `DELETE` statement.
- There are two API calls involved: one for figuring out which rows are affected and another for actually deleting the rows.

# Google Credentials

There are 2 modes of authentication that we support:

1. OAuth2 flow.
2. Service account flow.

## OAuth2 Flow

```go
auth, err := auth.NewOAuth2FromFile(
    "<path_to_client_secret_json>",
    "<path_to_cached_credentials_json>",
    scopes,
    auth.OAuth2Config{},
)
```

**Explanations:**

1. The `client_secret_json` can be obtained by creating a new OAuth2 credentials in [Google Developers Console](https://console.cloud.google.com/apis/credentials). You can put any link for the redirection URL field.
2. The `cached_credentials_json` will be created automatically once you have authenticated your Google Account via the normal OAuth2 flow. This file will contain the access token and refresh token.
3. The `scopes` tells Google what your application can do to your spreadsheets (`auth.GoogleSheetsReadOnly`, `auth.GoogleSheetsWriteOnly`, or `auth.GoogleSheetsReadWrite`).

During the OAuth2 flow, you will be asked to click a generated URL in the terminal.

1. Click the link and authenticate your Google Account.
2. You will eventually be redirected to another link which contains the authentication code (not the access token yet).
3. Copy and paste that final redirected URL into the terminal to finish the flow.

If you want to understand the details, you can start from this [Google OAuth2 page](https://developers.google.com/identity/protocols/oauth2/web-server).

## Service Account Flow

```go
auth, err := auth.NewServiceFromFile(
    "<path_to_service_account_json>",
    scopes,
    auth.ServiceConfig{},
)
```

**Explanations:**

1. The `service_account_json` can be obtained by following the steps in this [Google OAuth2 page](https://developers.google.com/identity/protocols/oauth2/service-account#creatinganaccount). The JSON file of interest is **the downloaded file after creating a new service account key**.
2. The `scopes` tells Google what your application can do to your spreadsheets (`auth.GoogleSheetsReadOnly`, `auth.GoogleSheetsWriteOnly`, or `auth.GoogleSheetsReadWrite`).

If you want to understand the details, you can start from this [Google Service Account page](https://developers.google.com/identity/protocols/oauth2/service-account).

> ### ⚠️ ⚠️ Warning
> Note that a service account is just like an account. The email in the `service_account_json` must be allowed to read/write into the Google Sheet itself just like a normal email address.
> If you don't do this, you will get an authorization error.

## Custom HTTP Client

We are using HTTP to connect to Google server to handle the authentication flow done by the [Golang OAuth2](golang.org/x/oauth2) library internally.
By default, it will be using the default HTTP client provided by `net/http`: `http.DefaultClient`.

For most simple use cases, `http.DefaultClient` should be sufficient.
However, if you want to use a custom `http.Client` instance, you can do that too.

```go
customHTTPClient := &http.Client{Timeout: time.Second*10}

auth, err := auth.NewOAuth2FromFile(
    "<path_to_client_secret_json>",
    "<path_to_cached_credentials_json>",
    scopes,
    auth.OAuth2Config{HTTPClient: customHTTPClient},
)

auth, err := auth.NewServiceFromFile(
    "<path_to_service_account_json>",
    scopes,
    auth.ServiceConfig{HTTPClient: customHTTPClient},
)
```

# Limitations

1. If you want to manually edit the Google Sheet, you can do it, but you need to understand the value encoding scheme.
2. It is not easy to support concurrent operations. Only few modes or abstractions allow concurrent operations.
3. Performance is not a high priority for this project.
4. `GoFreeLeh` does not support OAuth2 flow that spans across frontend and backend yet.

### (Google Sheets Key Value) Exclamation Mark `!` Prefix

1. We prepend an exclamation mark `!` in front of the value automatically.
2. This is to differentiate a client provided value of `#N/A` from the `#N/A` returned by the Google Sheet formula.
3. Hence, if you are manually updating the values via Google Sheets directly, you need to ensure there is an exclamation mark `!` prefix.

### (Google Sheets Row) Value Type in Cell

1. Note that we do not do any type conversion when inserting values into Google cells.
2. Values are marshalled using JSON internally by the Google Sheets library.
3. Values are interpreted automatically by the Google Sheet itself (unless you have changed the cell value type intentionally and manually). Let's take a look at some examples.
    - The literal string value of `"hello"` will automatically resolve into a `string` type for that cell.
    - The literal integer value of `1` will automatically resolve into a `number` type for that cell.
    - The literal string value of `"2000-1-1"`, however, will automatically resolve into a `date` type for that cell.
    - Note that this conversion is automatically done by Google Sheet.
    - Querying such column will have to consider the automatic type inference for proper querying. You can read here for [more details](https://developers.google.com/chart/interactive/docs/querylanguage#language-elements).
4. It may be possible to build a more type safe system in the future.
    - For example, we can store the column value type and store everything as strings instead.
    - During the data retrieval, we can read the column value type and perform explicit conversion.

# Disclaimer

- Please note that this library is in its early work.
- The interfaces provided are still unstable and we may change them at any point in time before it reaches v1.
- In addition, since the purpose of this library is for personal projects, we are going to keep it simple.
- Please use it at your own risk.

# License

This project is [MIT licensed](https://github.com/FreeLeh/GoFreeLeh/blob/main/LICENSE).
