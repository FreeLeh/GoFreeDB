# Generating Mocks

- Use `go generate ./...` to generate all mocks.
- Alternatively, you can just run the `mockgen` per `go:generate` line.

Note that we add `-build_constraint=gofreedb_test` to the `mockgen` command to all our mock files.
- We cannot use `_test.go` for the mock files because we want the mock files to be as
  close as possible to the source code files.
- If we use `_test.go` in the original package, test files in other packages
  cannot import the mock files.
- If we use `_mock.go` only, these files will be included in the main build for production use cases,
  unnecessarily including the mock files.
- As the mock files are only used for testing purposes, we use the `-build_constraint` flag. This ensures
  we don't include these mock files on production build (unless users intentionally include such
  build constraints).
- The `go test` script we are using already includes this build tag.

> If you are using Goland, you may need to add `gofreedb_test` into the `custom build tags` option in Goland.
> Otherwise, you may find the mock functions as "not found".