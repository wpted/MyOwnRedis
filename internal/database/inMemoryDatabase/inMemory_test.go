package inMemoryDatabase

import (
    "errors"
    "strconv"
    "testing"
)

func TestDb_Set(t *testing.T) {
    db := New()
    // Rare case, when key already exist but in listStorage.
    kLists := []struct {
        key  string
        list *StrNode
    }{
        {
            key: "foo_list",
            list: &StrNode{
                value: "1",
                next: &StrNode{
                    value: "2",
                    next:  nil,
                },
            },
        },
        {
            key: "bar_list",
            list: &StrNode{
                value: "3",
                next: &StrNode{
                    value: "4",
                    next:  nil,
                },
            },
        },
    }

    for _, kList := range kLists {
        db.listStorage[kList.key] = kList.list
    }

    testCases := []struct {
        key   string
        value string
    }{
        {"foo", "bar"},
        {"key", "value"},
        {"x", "1"},
        {"x", "r"},        // Should overwrite.
        {"foo_list", "1"}, // Should overwrite.
        {"bar_list", "2"}, // Should overwrite.¬
    }

    for _, tc := range testCases {
        db.Set(tc.key, tc.value)
        if db.stringStorage[tc.key] != tc.value {
            t.Errorf("Error setting value for key %s: expected value %s. got %s.\n", tc.key, tc.value, db.stringStorage[tc.key])
        }
    }

}

func TestDb_Get(t *testing.T) {
    db := New()
    t.Run("Test Get: Correct value type", func(t *testing.T) {
        kvs := []struct {
            key   string
            value string
        }{
            {key: "foo", value: "bar"},
            {key: "x", value: "1"},
            {key: "y", value: "2"},
            {key: "key", value: "value"},
        }

        for _, kv := range kvs {
            db.stringStorage[kv.key] = kv.value
        }

        for _, kv := range kvs {
            value, err := db.Get(kv.key)
            if err != nil {
                t.Errorf("Error getting value, got error %#v.\n", err)
            } else {
                if value != kv.value {
                    t.Errorf("Error getting value: expected %s, got %s.\n", kv.value, value)
                }
            }
        }

        value, err := db.Get("keyThatDoesntExist")
        if err != nil {
            t.Errorf("Error getting value, got error %#v.\n", err)
        }

        if value != Nil {
            t.Errorf("Error getting value: expected %s, got %s.\n", Nil, value)
        }

    })

    t.Run("Test Get: Incorrect value type", func(t *testing.T) {
        kLists := []struct {
            key  string
            list *StrNode
        }{
            {
                key: "foo_list",
                list: &StrNode{
                    value: "1",
                    next: &StrNode{
                        value: "2",
                        next:  nil,
                    },
                },
            },
            {
                key: "bar_list",
                list: &StrNode{
                    value: "3",
                    next: &StrNode{
                        value: "4",
                        next:  nil,
                    },
                },
            },
        }

        for _, kList := range kLists {
            db.listStorage[kList.key] = kList.list
        }
        for _, kList := range kLists {
            value, err := db.Get(kList.key)
            if !errors.Is(err, ErrNotString) {
                t.Errorf("Error incorrect error: expected %#v, got %#v.\n", ErrNotString, err)
            }
            if value != "" {
                t.Errorf("Error incorrect value: expected empty string, got %s.\n", value)
            }
        }
    })
}

func TestDb_Exists(t *testing.T) {
    db := New()
    kvs := []struct {
        key   string
        value string
    }{
        {key: "foo", value: "bar"},
        {key: "x", value: "1"},
        {key: "y", value: "2"},
        {key: "key", value: "value"},
    }

    for _, kv := range kvs {
        db.stringStorage[kv.key] = kv.value
    }

    kLists := []struct {
        key  string
        list *StrNode
    }{
        {
            key: "foo_list",
            list: &StrNode{
                value: "1",
                next: &StrNode{
                    value: "2",
                    next:  nil,
                },
            },
        },
        {
            key: "bar_list",
            list: &StrNode{
                value: "3",
                next: &StrNode{
                    value: "4",
                    next:  nil,
                },
            },
        },
    }

    for _, kList := range kLists {
        db.listStorage[kList.key] = kList.list
    }

    testCases := []struct {
        key    string
        exists bool
    }{
        {key: "foo", exists: true},
        {key: "x", exists: true},
        {key: "y", exists: true},
        {key: "key", exists: true},
        {key: "foo_list", exists: true},
        {key: "bar_list", exists: true},
        {key: "keyThatDoesntExist", exists: false},
        {key: "", exists: false},
    }

    for _, tc := range testCases {
        exists := db.Exists(tc.key)
        if exists != tc.exists {
            t.Errorf("Error key existence: expected %t, got %t.\n", tc.exists, exists)
        }
    }
}

func TestDb_Delete(t *testing.T) {
    db := New()
    kvs := []struct {
        key   string
        value string
    }{
        {key: "foo", value: "bar"},
        {key: "x", value: "1"},
        {key: "y", value: "2"},
        {key: "key", value: "value"},
    }

    for _, kv := range kvs {
        db.stringStorage[kv.key] = kv.value
    }

    kLists := []struct {
        key  string
        list *StrNode
    }{
        {
            key: "foo_list",
            list: &StrNode{
                value: "1",
                next: &StrNode{
                    value: "2",
                    next:  nil,
                },
            },
        },
        {
            key: "bar_list",
            list: &StrNode{
                value: "3",
                next: &StrNode{
                    value: "4",
                    next:  nil,
                },
            },
        },
    }

    for _, kList := range kLists {
        db.listStorage[kList.key] = kList.list
    }

    keysToDelete := []string{
        "foo",
        "foo_list",
        "keyThatDoesntExist", // no-op
        "",                   // no-op
    }

    deletedKeys := db.Delete(keysToDelete...)

    // Check if the number of the deleted key matches.
    if deletedKeys != 2 {
        t.Errorf("Error number of keys that were removed: expected %d, got %d.\n", 2, deletedKeys)
    }

    testCases := []struct {
        key    string
        exists bool
    }{
        {key: "foo", exists: false},
        {key: "x", exists: true},
        {key: "y", exists: true},
        {key: "key", exists: true},
        {key: "foo_list", exists: false},
        {key: "bar_list", exists: true},
        {key: "keyThatDoesntExist", exists: false}, // Should be a no-op and shouldn't have exists.
        {key: "", exists: false},                   // Should be a no-op and shouldn't have exists.
    }

    for _, tc := range testCases {
        exists := db.Exists(tc.key)
        if exists != tc.exists {
            t.Errorf("Error key existence: expected %t, got %t.\n", tc.exists, exists)
        }
    }
}

func TestDb_Increment(t *testing.T) {
    db := New()

    t.Run("Test Increment: Correct input", func(t *testing.T) {
        kvs := []struct {
            key   string
            value string
        }{
            {key: "x", value: "1"},
            {key: "y", value: "2"},
        }
        for _, kv := range kvs {
            db.stringStorage[kv.key] = kv.value
        }

        testCases := []struct {
            input         string
            expectedValue string
        }{
            {input: "x", expectedValue: "2"},
            {input: "y", expectedValue: "3"},
            {input: "z", expectedValue: "1"}, // key that doesn't exist should create the key-value pair with a default value 0, then add 1 to it.
        }

        for _, tc := range testCases {
            re, err := db.Increment(tc.input)
            if err != nil {
                t.Errorf("Error incrementing, got error: %#v.\n", err)
            }

            if strconv.Itoa(re) != tc.expectedValue {
                t.Errorf("Error incrementing result: expected %s, got %d.\n", tc.expectedValue, re)
            }
        }
    })

    t.Run("Test Increment: Incorrect input", func(t *testing.T) {
        kvs := []struct {
            key   string
            value string
        }{
            {key: "foo", value: "bar"},
            {key: "key", value: "value"},
        }

        for _, kv := range kvs {
            db.stringStorage[kv.key] = kv.value
        }

        kLists := []struct {
            key  string
            list *StrNode
        }{
            {
                key: "foo_list",
                list: &StrNode{
                    value: "1",
                    next: &StrNode{
                        value: "2",
                        next:  nil,
                    },
                },
            },
            {
                key: "bar_list",
                list: &StrNode{
                    value: "3",
                    next: &StrNode{
                        value: "4",
                        next:  nil,
                    },
                },
            },
        }

        for _, kList := range kLists {
            db.listStorage[kList.key] = kList.list
        }

        incorrectKeys := []string{"foo", "key", "foo_list", "bar_list"}
        for _, incorrectKey := range incorrectKeys {
            _, err := db.Increment(incorrectKey)
            if !errors.Is(err, ErrNotInteger) {
                t.Errorf("Error incorrect error: expected %#v, got %#v.\n", ErrNotInteger, err)
            }
        }
    })
}

func TestDb_Decrement(t *testing.T) {
    db := New()

    t.Run("Test Decrement: Correct input", func(t *testing.T) {
        kvs := []struct {
            key   string
            value string
        }{
            {key: "x", value: "1"},
            {key: "y", value: "2"},
        }
        for _, kv := range kvs {
            db.stringStorage[kv.key] = kv.value
        }

        testCases := []struct {
            input         string
            expectedValue string
        }{
            {input: "x", expectedValue: "0"},
            {input: "y", expectedValue: "1"},
            {input: "z", expectedValue: "-1"}, // key that doesn't exist should create the key-value pair with a default value 0, then decrement it by 1.
        }

        for _, tc := range testCases {
            re, err := db.Decrement(tc.input)
            if err != nil {
                t.Errorf("Error incrementing, got error: %#v.\n", err)
            }
            if strconv.Itoa(re) != tc.expectedValue {
                t.Errorf("Error incrementing result: expected %s, got %d.\n", tc.expectedValue, re)
            }
        }
    })

    t.Run("Test Decrement: Incorrect input", func(t *testing.T) {
        kvs := []struct {
            key   string
            value string
        }{
            {key: "foo", value: "bar"},
            {key: "key", value: "value"},
        }

        for _, kv := range kvs {
            db.stringStorage[kv.key] = kv.value
        }

        kLists := []struct {
            key  string
            list *StrNode
        }{
            {
                key: "foo_list",
                list: &StrNode{
                    value: "1",
                    next: &StrNode{
                        value: "2",
                        next:  nil,
                    },
                },
            },
            {
                key: "bar_list",
                list: &StrNode{
                    value: "3",
                    next: &StrNode{
                        value: "4",
                        next:  nil,
                    },
                },
            },
        }

        for _, kList := range kLists {
            db.listStorage[kList.key] = kList.list
        }

        incorrectKeys := []string{"foo", "key", "foo_list", "bar_list"}
        for _, incorrectKey := range incorrectKeys {
            _, err := db.Decrement(incorrectKey)
            if !errors.Is(err, ErrNotInteger) {
                t.Errorf("Error incorrect error: expected %#v, got %#v.\n", ErrNotInteger, err)
            }
        }
    })
}

func TestDb_LRange(t *testing.T) {
    db := New()
    db.stringStorage["x"] = "1"

    // We only test for the incorrect ones here. The correct ones were tested already.
    incorrectKey := "x"
    // number is random since we're not testing the actual ranging.
    _, err := db.LRange(incorrectKey, 0, 0)
    if !errors.Is(err, ErrNotList) {
        t.Errorf("Error incorrect error: expected %#v, got %#v.\n", ErrNotList, err)
    }
}

func TestDb_LeftPush(t *testing.T) {
    db := New()
    t.Run("Test LeftPush: Correct input", func(t *testing.T) {
        testCases := []struct {
            key            string
            list           *StrNode
            inputValues    []string
            expectedLength int
            expectedArr    []string
        }{
            {
                key:            "bar_list",
                list:           nil,
                inputValues:    []string{"hello world", "1"},
                expectedLength: 2,
                expectedArr:    []string{"1", "hello world"},
            },
            {
                key: "foo_list",
                list: &StrNode{
                    value: "1",
                    next: &StrNode{
                        value: "2",
                        next:  nil,
                    },
                },
                expectedLength: 4,
                inputValues:    []string{"a", "b"},
                expectedArr:    []string{"b", "a", "1", "2"},
            },
        }

        for _, tc := range testCases {
            if tc.list != nil {
                // Since we want to test lpush to a key that doesn't exist, we don't create key in the listStorage if tc.list == nil.
                db.listStorage[tc.key] = tc.list
            }
        }

        for _, tc := range testCases {
            length, err := db.LeftPush(tc.key, tc.inputValues...)
            if err != nil {
                t.Errorf("Error left pushing values to key, got error %#v.\n", err)
            }

            if length != tc.expectedLength {
                t.Errorf("Error list expectedLength: expected %d, got %d.\n", tc.expectedLength, length)
            } else {
                head := db.listStorage[tc.key]
                temp := head
                gotArr := make([]string, 0)
                for temp != nil {
                    gotArr = append(gotArr, temp.value)
                    temp = temp.next
                }

                for n, ele := range gotArr {
                    if ele != tc.expectedArr[n] {
                        t.Errorf("Error left pushing values: expected %s, got %s.\n", tc.expectedArr[n], ele)
                    }
                }
            }
        }
    })

    t.Run("Test LeftPush: Incorrect input", func(t *testing.T) {
        // Insert some key-value in the database and left push to the key.
        incorrectKVs := []struct {
            key   string
            value string
        }{
            {key: "foo", value: "bar"},
            {key: "x", value: "1"},
        }

        for _, kv := range incorrectKVs {
            db.stringStorage[kv.key] = kv.value
        }

        for _, kv := range incorrectKVs {
            _, err := db.LeftPush(kv.key, kv.value)
            if !errors.Is(err, ErrNotList) {
                t.Errorf("Error incorrect error: expected %#v, got %#v.\n", ErrNotList, err)
            }
        }
    })
}

func TestDb_RightPush(t *testing.T) {
    db := New()
    t.Run("Test RightPush: Correct input", func(t *testing.T) {
        testCases := []struct {
            key            string
            list           *StrNode
            inputValues    []string
            expectedLength int
            expectedArr    []string
        }{
            {
                key:            "bar_list",
                list:           nil,
                inputValues:    []string{"hello world", "1"},
                expectedLength: 2,
                expectedArr:    []string{"hello world", "1"},
            },
            {
                key: "foo_list",
                list: &StrNode{
                    value: "1",
                    next: &StrNode{
                        value: "2",
                        next:  nil,
                    },
                },
                expectedLength: 4,
                inputValues:    []string{"a", "b"},
                expectedArr:    []string{"1", "2", "a", "b"},
            },
        }

        for _, tc := range testCases {
            if tc.list != nil {
                // Since we want to test lpush to a key that doesn't exist, we don't create key in the listStorage if tc.list == nil.
                db.listStorage[tc.key] = tc.list
            }
        }

        for _, tc := range testCases {
            length, err := db.RightPush(tc.key, tc.inputValues...)
            if err != nil {
                t.Errorf("Error right pushing values to key, got error %#v.\n", err)
            }

            if length != tc.expectedLength {
                t.Errorf("Error list expectedLength: expected %d, got %d.\n", tc.expectedLength, length)
            } else {
                head := db.listStorage[tc.key]
                temp := head
                gotArr := make([]string, 0)
                for temp != nil {
                    gotArr = append(gotArr, temp.value)
                    temp = temp.next
                }

                for n, ele := range gotArr {
                    if ele != tc.expectedArr[n] {
                        t.Errorf("Error left pushing values: expected %s, got %s.\n", tc.expectedArr[n], ele)
                    }
                }
            }
        }
    })

    t.Run("Test RightPush: Incorrect input", func(t *testing.T) {
        // Insert some key-value in the database and left push to the key.
        incorrectKVs := []struct {
            key   string
            value string
        }{
            {key: "foo", value: "bar"},
            {key: "x", value: "1"},
        }

        for _, kv := range incorrectKVs {
            db.stringStorage[kv.key] = kv.value
        }

        for _, kv := range incorrectKVs {
            _, err := db.RightPush(kv.key, kv.value)
            if !errors.Is(err, ErrNotList) {
                t.Errorf("Error incorrect error: expected %#v, got %#v.\n", ErrNotList, err)
            }
        }
    })
}
