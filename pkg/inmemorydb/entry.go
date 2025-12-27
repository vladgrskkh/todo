package inmemorydb

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

// action represents the type of operation (Put or Delete) in the database log.
type action string

const (
	Put action = "put"
	Del action = "del"
)

var (
	ErrBadFormat    = errors.New("inmemorydb: bad line format")
	ErrCannotDecode = errors.New("inmemorydb: cannot decode element")
)

type entry struct {
	action action
	key    string
	value  []byte
}

func newEntry(action action, key string, value []byte) *entry {
	return &entry{
		action: action,
		key:    key,
		value:  value,
	}
}

// newEntryFromLine parses a database file line and returns the corresponding entry.
// Lines must be in the format: action,base64(key),base64(value)
func newEntryFromLine(line string) (*entry, error) {
	elements := strings.Split(line, ",")
	if len(elements) != 3 {
		return nil, ErrBadFormat
	}

	key, err := base64.StdEncoding.DecodeString(elements[1])
	if err != nil {
		return nil, fmt.Errorf("inmemorydb: unable to decode key: %w", err)
	}

	value, err := base64.StdEncoding.DecodeString(elements[2])
	if err != nil {
		return nil, fmt.Errorf("inmemorydb: unable to decode value: %w", err)
	}

	return &entry{
		action: action(elements[0]),
		key:    string(key),
		value:  value,
	}, nil
}

func (e *entry) toBytes() []byte {
	return fmt.Appendf(nil, "%s,%s,%s\n", e.action, base64.StdEncoding.EncodeToString([]byte(e.key)), base64.StdEncoding.EncodeToString([]byte(e.value)))
}
