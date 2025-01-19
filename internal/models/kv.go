package models

import "errors"

// KVMode defines the mode of the key value store.
// For more details, please read the README file.
type KVMode int

const (
	KVModeDefault    KVMode = 0
	KVModeAppendOnly KVMode = 1

	NAValue    = "#N/A"
	ErrorValue = "#ERROR!"
)

// ErrKeyNotFound is returned only for the key-value store and when the key does not exist.
var (
	ErrKeyNotFound = errors.New("error key not found")
)
