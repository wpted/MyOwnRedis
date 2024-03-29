package inMemoryDatabase

import (
    "encoding/csv"
    "errors"
    "io"
    "os"
    "strconv"
    "sync"
)

const Nil = "nil"
const (
    Tmp      = "tmp/"
    DumpFile = "tmp/dump.csv"
)

const (
    TypeString = "<String>"
    TypeList   = "<List>"
)

var (
    ErrNotString  = errors.New("error fetched value is not a string")
    ErrNotInteger = errors.New("error fetched value is not an integer")
    ErrNotList    = errors.New("error fetched value is not a list")
)

// Db instance.
type Db struct {
    stringStorage map[string]string
    listStorage   map[string]*StrNode
    sync.RWMutex
}

// New creates a new Db.
// If there's existing dump.csv, load the data instead.
func New() *Db {
    // Create directory 'tmp' and 'tmp/dump.csv' if not exist.
    if _, err := os.Stat(Tmp); os.IsNotExist(err) {
        if err = os.Mkdir(Tmp, os.ModeDir|os.ModePerm); err != nil {
            panic(err)
        }
        _, err = os.OpenFile(DumpFile, os.O_CREATE, 0644)
        if err != nil {
            panic(err)
        }
        return &Db{
            stringStorage: make(map[string]string),
            listStorage:   make(map[string]*StrNode),
        }
    }

    db, err := loadDatabase()
    if err != nil {
        panic(err)
    }
    return db

}

// Set sets key to hold the string value.
// If key already holds a value, it is overwritten, regardless of its type.
func (d *Db) Set(key string, value string) {
    d.Lock()
    defer d.Unlock()
    // Check if key is in listStorage.
    if _, ok := d.listStorage[key]; ok {
        // Delete the key-value pair if existed.
        delete(d.listStorage, key)
    }

    d.stringStorage[key] = value
}

// Get returns the string value of the key. If the key doesn't exist, "nil" is returned.
// An error is returned if the value stored at key is not a string, because Get only handles string values.
func (d *Db) Get(key string) (string, error) {
    // With RLock, all goroutines can read concurrently without blocking each other.
    d.RLock()
    defer d.RUnlock()

    value, okInStringStorage := d.stringStorage[key]

    _, okInListStorage := d.listStorage[key]

    if okInListStorage {
        return "", ErrNotString
    }

    if okInStringStorage {
        return value, nil
    } else {
        return Nil, nil
    }
}

// GetAllKeys returns all the keys within the database.
func (d *Db) GetAllKeys() []string {
    allKeys := make([]string, 0)

    d.RLock()
    defer d.RUnlock()

    for k := range d.stringStorage {
        allKeys = append(allKeys, k)
    }

    for k := range d.listStorage {
        allKeys = append(allKeys, k)
    }

    return allKeys
}

// Exists determines whether a key exists.
func (d *Db) Exists(key string) bool {
    d.RLock()
    defer d.RUnlock()

    _, strOk := d.stringStorage[key]
    _, listOk := d.listStorage[key]
    return strOk || listOk
}

// Delete removes the specified key and returns numbers of the actual deleted key.
// The action is ignored if the keys doesn't exist.
func (d *Db) Delete(keys ...string) int {
    d.Lock()
    defer d.Unlock()

    var deletedKeys int
    // delete is a no-op if the key doesn't exist in the map.
    for _, key := range keys {
        _, inStringStorage := d.stringStorage[key]
        _, inListStorage := d.listStorage[key]
        if inStringStorage || inListStorage {
            deletedKeys++
        }

        delete(d.stringStorage, key)
        delete(d.listStorage, key)
    }

    return deletedKeys
}

// Increment increments the number stored at key by 1. If the key doesn't exist, it is set to 0 before performing the operation.
// An error is returned if the key contains a value of the wrong type or contains a string that cannot be represented as integer.
// The whole function should be a string operation because Redis does not have a dedicated integer type ( where all values in Redis are stored in their strings ).
func (d *Db) Increment(key string) (int, error) {
    d.Lock()
    defer d.Unlock()

    // Key exist but in listStorage.
    _, inListStorage := d.listStorage[key]
    value, inStringStorage := d.stringStorage[key]

    var result int
    var err error

    if !inStringStorage && !inListStorage {
        d.stringStorage[key] = "1"
        result = 1
    } else if inListStorage && !inStringStorage {
        // Key exist but in wrong storage -> Error value type.
        return 0, ErrNotInteger
    } else {
        result, err = strconv.Atoi(value)
        if err != nil {
            return 0, ErrNotInteger // Value not an Integer string.
        }
        result++
        d.stringStorage[key] = strconv.Itoa(result)
    }
    return result, nil
}

// Decrement decrements the number stored at key by one. If the key doesn't exists, it is set to 0 before performing the operation.
// An error is returned if the key contains a value of the wrong type or contains a string that cannot be represented as integer.
// The whole function should be a string operation because Redis does not have a dedicated integer type ( where all values in Redis are stored in their strings ).
func (d *Db) Decrement(key string) (int, error) {
    d.Lock()
    defer d.Unlock()
    _, inListStorage := d.listStorage[key]
    value, inStringStorage := d.stringStorage[key]

    var result int
    var err error

    if !inStringStorage && !inListStorage {
        d.stringStorage[key] = "-1"
        result = -1
    } else if inListStorage && !inStringStorage {
        return 0, ErrNotInteger
    } else {
        result, err = strconv.Atoi(value)
        if err != nil {
            return 0, ErrNotInteger
        }
        result--
        d.stringStorage[key] = strconv.Itoa(result)
    }
    return result, nil
}

// LRange returns the specified elements of the list stored at key.
// The offsets start and stop are zero-based indexes, with 0 being the first element of the list ( the head of the list ), 1 being the next element and so on.
// These offsets can also be negative numbers indicating offsets starting at the end of the list.
// For example, -1 is the last element of the list, -2 the penultimate, and so on.
// Out of range indexes will not produce an error, if start is larger than the end of the list, an empty array is returned.
// If stop is larger than the actual end of the list, LRange will treat it like the last element of the list.
// If the key doesn't exist, returns an empty list.
func (d *Db) LRange(key string, start, stop int) ([]string, error) {
    d.RLock()
    defer d.RUnlock()

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
    d.Lock()
    defer d.Unlock()

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
    d.Lock()
    defer d.Unlock()

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

// SaveDatabase persists all data to 'tmp/dump.csv'
func (d *Db) SaveDatabase() error {
    // Lock the database from writing new data.
    d.RLock()
    defer d.RUnlock()

    // Open a csv file.
    file, err := os.OpenFile(DumpFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
    if err != nil {
        return err
    }

    defer func() {
        if err = file.Close(); err != nil {
            panic(err)
        }
    }()

    // Create csv writer.
    w := csv.NewWriter(file)

    // Clean underlying buffer in w.
    defer w.Flush()

    record := make([][]string, 0)

    // Write the string section.
    for key, value := range d.stringStorage {
        record = append(record, []string{TypeString, key, value})
    }

    // Write the list section.
    for key, list := range d.listStorage {
        curRow := []string{TypeList, key}
        temp := list
        for temp != nil {
            curRow = append(curRow, temp.value)
            temp = temp.next
        }
        record = append(record, curRow)
    }

    err = w.WriteAll(record)
    if err != nil {
        return err
    }

    return nil
}

// loadDatabase loads from 'tmp/dump.csv'
func loadDatabase() (*Db, error) {
    db := &Db{
        stringStorage: make(map[string]string),
        listStorage:   make(map[string]*StrNode),
    }

    // Read from dump.csv and store to d.Db
    file, err := os.OpenFile(DumpFile, os.O_CREATE|os.O_RDONLY, 0644)
    if err != nil {
        return nil, err
    }

    defer func() {
        err = file.Close()
        if err != nil {
            panic(err)
        }
    }()

    // Create csv reader.
    r := csv.NewReader(file)

    // Disable record length test in the CSV reader.
    r.FieldsPerRecord = -1
    var record []string
    for {
        record, err = r.Read()
        if err != nil {
            if errors.Is(err, io.EOF) {
                break
            }
            return nil, err
        }

        if len(record) != 0 {
            switch record[0] {
            case TypeString:
                db.Set(record[1], record[2])
            case TypeList:
                _, _ = db.RightPush(record[1], record[2:]...)
            }
        }
    }
    return db, nil
}
