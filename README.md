# GoFreeLeh

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
    * [Google Sheets](#google-sheets)
        * [Key value Interface](#key-value-interface)
        * [Key Value Store Modes](#key-value-store-modes)
            * [Default Mode](#default-mode)
            * [Append Only Mode](#append-only-mode)
    * [Google Credentials](#google-credentials)
        * [OAuth2 Flow](#oauth2-flow)
        * [Service Account Flow](#service-account-flow)
        * [Custom HTTP Client](#custom-http-client)
* [Limitations](#limitations)
* [Disclaimer](#disclaimer)
* [Acknowledgement](#acknowledgement)
* [License](#license)
    

# Key Value Store

## Google Sheets

```go
// If using Google Service Account.
auth, _ := auth.NewService(
    "<path_to_service_account_json>",
    []string{auth.GoogleSheetsReadWrite},
    auth.ServiceConfig{},
)

// If using Google OAuth2 Flow.
auth, err := auth.NewOAuth2(
    "<path_to_client_secret_json>",
    "<path_to_cached_credentials_json>",
    []string{auth.GoogleSheetsReadWrite},
    auth.OAuth2Config{},
)

// Below are the same regardless of the auth client chosen above.
kv := freeleh.NewGoogleSheetKeyValue(
    auth,
    "<spreadsheet_id>",
    "<sheet_name>",
    freeleh.GoogleSheetKVConfig{Mode: freeleh.KVSetModeAppendOnly},
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

### Interface

- `Get(ctx context.Context, key string) ([]byte, error)`

- `Set(ctx context.Context, key string, value []byte) error`

- `Delete(ctx context.Context, key string) error`

> ### ⚠️ ⚠️ Warning
> Please note that only `[]byte` values are supported at the moment.

### Key Value Store Modes

There are 2 different modes supported:

1. Default mode.
2. Append only mode.

```go
// Default mode
kv := freeleh.NewGoogleSheetKeyValue(auth, "<spreadsheet_id>", "<sheet_name>", freeleh.GoogleSheetKVConfig{Mode: freeleh.KVModeDefault})

// Append only mode
kv := freeleh.NewGoogleSheetKeyValue(auth, "<spreadsheet_id>", "<sheet_name>", freeleh.GoogleSheetKVConfig{Mode: freeleh.KVModeAppendOnly})
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

## Google Credentials

There are 2 modes of authentication that we support:

1. OAuth2 flow.
2. Service account flow.

### OAuth2 Flow

```go
auth, err := auth.NewOAuth2(
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

### Service Account Flow

```go
auth, err := auth.NewService(
    "<path_to_service_account_json>",
    scopes,
    auth.ServiceConfig{},
)
```

**Explanations:**

1. The `service_account_json` can be obtained by following the steps in this [Google OAuth2 page](https://developers.google.com/identity/protocols/oauth2/service-account#creatinganaccount). The JSON file of interest is **the downloaded file after creating a new service account key**.
2. The `scopes` tells Google what your application can do to your spreadsheets (`auth.GoogleSheetsReadOnly`, `auth.GoogleSheetsWriteOnly`, or `auth.GoogleSheetsReadWrite`).

If you want to understand the details, you can start from this [Google Service Account page](https://developers.google.com/identity/protocols/oauth2/service-account).

### Custom HTTP Client

We are using HTTP to connect to Google server to handle the authentication flow done by the [Golang OAuth2](golang.org/x/oauth2) library internally.
By default, it will be using the default HTTP client provided by `net/http`: `http.DefaultClient`.

For most simple use cases, `http.DefaultClient` should be sufficient.
However, if you want to use a custom `http.Client` instance, you can do that too.

```go
customHTTPClient := &http.Client{Timeout: time.Second*10}

auth, err := auth.NewOAuth2(
    "<path_to_client_secret_json>",
    "<path_to_cached_credentials_json>",
    scopes,
    auth.OAuth2Config{HTTPClient: customHTTPClient},
)

auth, err := auth.NewService(
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

<details>
<summary>Notable Protocols</summary>
<p>

### Exclamation Mark `!` Prefix

1. We prepend an exclamation mark `!` in front of the value automatically.
2. This is to differentiate a client provided value of `#N/A` from the `#N/A` returned by the Google Sheet formula.
3. Hence, if you are manually updating the values via Google Sheets directly, you need to ensure there is an exclamation mark `!` prefix.

</p>
</details>

# Disclaimer

This project is still a **Work in Progress**. The interfaces and APIs are prone to changes in the foreseeable future.

# Acknowledgement

Thanks to @fata.nugraha for inspiring me to start this project and discussing the details together.

# License

This project is [MIT licensed](https://github.com/FreeLeh/GoFreeLeh/blob/main/LICENSE).