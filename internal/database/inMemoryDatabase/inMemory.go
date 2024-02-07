package inMemoryDatabase

import (
    "errors"
    "strconv"
    "sync"
)

const NIL = "nil"

var (
    ErrNotString  = errors.New("error fetched value is not a string")
    ErrNotInteger = errors.New("error fetched value is not an integer")
    ErrNotList    = errors.New("error fetched value is not a list")
)

// Db instance.
type Db struct {
    stringStorage map[string]string
    listStorage   map[string]*StrNode
    mu            sync.RWMutex
}

// New creates a new Db.
func New() *Db {
    return &Db{
        stringStorage: make(map[string]string),
        listStorage:   make(map[string]*StrNode),
        mu:            sync.RWMutex{},
    }
}

// Set sets key to hold the string value.
// If key already holds a value, it is overwritten, regardless of its type.
func (d *Db) Set(key string, value string) {
    d.mu.Lock()
    defer d.mu.Unlock()

    d.stringStorage[key] = value
}

// Get returns the string value of the key. If the key doesn't exist, "nil" is returned.
// An error is returned if the value stored at key is not a string, because Get only handles string values.
func (d *Db) Get(key string) (string, error) {
    // With RLock, all goroutines can read concurrently without blocking each other.
    d.mu.RLock()
    defer d.mu.RUnlock()

    value, okInStringStorage := d.stringStorage[key]

    _, okInListStorage := d.listStorage[key]

    if okInListStorage {
        return "", ErrNotString
    }

    if okInStringStorage {
        return value, nil
    } else {
        return NIL, nil
    }
}

// Exists determines whether a key exists.
func (d *Db) Exists(key string) bool {
    d.mu.RLock()
    defer d.mu.RUnlock()

    _, strOk := d.stringStorage[key]
    _, listOk := d.listStorage[key]
    return strOk || listOk
}

// Delete removes the specified key. The action is ignored if the key doesn't exists.
func (d *Db) Delete(key string) {
    d.mu.Lock()
    defer d.mu.Unlock()

    // delete is a no-op if the key doesn't exist in the map.
    delete(d.stringStorage, key)
    delete(d.listStorage, key)
}

// Increment increments the number stored at key by 1. If the key doesn't exists, it is set to 0 before performing the operation.
// An error is returned if the key contains a value of the wrong type or contains a string that cannot be represented as integer.
// The whole function should be a string operation because Redis does not have a dedicated integer type ( where all values in Redis are stored in their strings ).
func (d *Db) Increment(key string) error {
    d.mu.Lock()
    defer d.mu.Unlock()

    // Key exist but in
    _, inListStorage := d.listStorage[key]
    value, inStringStorage := d.stringStorage[key]

    if !inStringStorage && !inListStorage {
        d.stringStorage[key] = "1"
    } else if inListStorage && !inStringStorage {
        // Key exist but in wrong storage -> Error value type.
        return ErrNotInteger
    } else {
        intValue, err := strconv.Atoi(value)
        if err != nil {
            return ErrNotInteger // Value not an Integer string.
        }
        intValue++
        d.stringStorage[key] = strconv.Itoa(intValue)
    }
    return nil
}

// Decrement decrements the number stored at key by one. If the key doesn't exists, it is set to 0 before performing the operation.
// An error is returned if the key contains a value of the wrong type or contains a string that cannot be represented as integer.
// The whole function should be a string operation because Redis does not have a dedicated integer type ( where all values in Redis are stored in their strings ).
func (d *Db) Decrement(key string) error {
    d.mu.Lock()
    defer d.mu.Unlock()
    _, inListStorage := d.listStorage[key]
    value, inStringStorage := d.stringStorage[key]

    if !inStringStorage && !inListStorage {
        d.stringStorage[key] = "-1"
    } else if inListStorage && !inStringStorage {
        return ErrNotInteger
    } else {
        intValue, err := strconv.Atoi(value)
        if err != nil {
            return ErrNotInteger
        }
        intValue--
        d.stringStorage[key] = strconv.Itoa(intValue)
    }
    return nil
}

// LRange returns the specified elements of the list stored at key.
// The offsets start and stop are zero-based indexes, with 0 being the first element of the list ( the head of the list ), 1 being the next element and so on.
// These offsets can also be negative numbers indicating offsets starting at the end of the list.
// For example, -1 is the last element of the list, -2 the penultimate, and so on.
// Out of range indexes will not produce an error, if start is larger than the end of the list, an empty array is returned.
// If stop is larger than the actual end of the list, LRange will treat it like the last element of the list.
// If the key doesn't exist, returns an empty list.
func (d *Db) LRange(key string, start, stop int) ([]string, error) {
    d.mu.RLock()
    defer d.mu.RUnlock()

    if _, ok := d.stringStorage[key]; ok {
        // Values having wrong type.
        return nil, ErrNotList
    } else if _, ok := d.listStorage[key]; !ok {
        // Keys that doesn't exist.
        return nil, nil
    } else {
        return d.listStorage[key].LRange(start, stop), nil
    }
}

// LeftPush prepends one or more elements to the list.
// The return value is the Len of the list after the push operation.
// If key doesn't exist, it is created as empty list before performing the push operations.
// When key holds a value that is not a list, an error is returned.
// It is possible to push multiple elements using a single command call just specifying multiple arguments at the end of the command.
// Elements are inserted one after the other to the head of the list, from the leftmost element to the rightmost element.
// So for instance the command `LPUSH myList a b c` will result into a list containing `c` as first element, `b` as second element and `a` as third element.
func (d *Db) LeftPush(key string, values ...string) (int, error) {
    d.mu.Lock()
    defer d.mu.Unlock()

    _, inStringStorage := d.stringStorage[key]
    if inStringStorage {
        return 0, ErrNotList
    }

    _, inListStorage := d.listStorage[key]
    if !inListStorage {
        d.listStorage[key] = nil
    }

    // Push the values to the current node.
    newHead := d.listStorage[key].LeftPush(values)
    // Assign newHead as the new corresponding list.
    d.listStorage[key] = newHead

    return d.listStorage[key].Len(), nil
}

// RightPush appends one or more elements to a list.
// The return value is the Len of the list after the push operation.
// If key doesn't exist, it is created as empty list before performing the push operations.
// When key holds a value that is not a list, an error is returned.
// It is possible to push multiple elements using a single command call just specifying multiple arguments at the end of the command.
// Elements are inserted one after the other to the tail of the list, from the leftmost element to the rightmost element.
// So for instance the command `RPUSH myList a b c` will result into a list containing `a` as first element, `b` as second element and `c` as third element.
func (d *Db) RightPush(key string, values ...string) (int, error) {
    d.mu.Lock()
    defer d.mu.Unlock()

    _, inStringStorage := d.stringStorage[key]
    if inStringStorage {
        return 0, ErrNotList
    }

    _, inListStorage := d.listStorage[key]
    if !inListStorage {
        d.listStorage[key] = nil
    }
    newHead := d.listStorage[key].RightPush(values)
    d.listStorage[key] = newHead

    return d.listStorage[key].Len(), nil
}

func (d *Db) SaveDatabase() error {
    return nil
}

func (d *Db) LoadDatabase() error {
    return nil
}
