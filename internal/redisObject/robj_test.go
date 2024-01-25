package redisObject

import (
    "errors"
    "testing"
)

func Test_Deserialize(t *testing.T) {
    t.Run("Test Deserialize - Invalid commands", func(t *testing.T) {
        testCases := []struct {
            input []byte
            err   error
        }{
            {},
        }

        for _, tc := range testCases {
            _, err := Deserialize(tc.input)
            if !errors.Is(err, ErrInvalidCommand) {
                t.Errorf("error incorrect error: expected %v, got %v.\n", ErrInvalidCommand, err)
            }
        }
    })

    t.Run("Test Deserialize - Valid commands", func(t *testing.T) {
        testCases := []struct {
            input  []byte
            result RObj
        }{ //RESP
            {
                input:  []byte(`$-1\r\n`), // NULL value.
                result: RObj{Type: NULLS, Content: "NULL", nextRObj: nil},
            },
            {
                input:  []byte(`*-1\r\n`), // NULL value.
                result: RObj{Type: NULLS, Content: "NULL", nextRObj: nil},
            },
            {
                input:  []byte(`-Error message\r\n`), // Errors.
                result: RObj{Type: SIMPLEERRORS, Content: "Error message", nextRObj: nil},
            },
            {
                input:  []byte(`+OK\r\n`), // Simple strings.
                result: RObj{Type: SIMPLESTRINGS, Content: "OK", nextRObj: nil},
            },
            {
                input:  []byte(`+hello world\r\n`),
                result: RObj{Type: SIMPLESTRINGS, Content: "hello world", nextRObj: nil},
            },
            {
                input:  []byte(`$0\r\n`), // Integers.
                result: RObj{Type: INTEGERS, Content: "0", nextRObj: nil},
            },
            {
                input: []byte(`*1\r\n\$4\r\nPING\r\n`), // Arrays.
                result: RObj{
                    Type: ARRAYS,
                    nextRObj: &RObj{
                        Type:     SIMPLESTRINGS,
                        Content:  "PING",
                        nextRObj: nil,
                    },
                },
            },
            {
                input: []byte(`*2\r\n$4\r\necho\r\n$11\r\nhello world\r\n`),
                result: RObj{
                    Type: ARRAYS,
                    nextRObj: &RObj{
                        Type:    SIMPLESTRINGS,
                        Content: "echo",
                        nextRObj: &RObj{
                            Type:    SIMPLESTRINGS,
                            Content: "hello world",
                        },
                    },
                },
            },
            {
                input: []byte(`*2\r\n$3\r\nget\r\n$3\r\nkey\r\n`),
                result: RObj{
                    Type: ARRAYS,
                    nextRObj: &RObj{
                        Type:    SIMPLESTRINGS,
                        Content: "get",
                        nextRObj: &RObj{
                            Type:    SIMPLESTRINGS,
                            Content: "key",
                        },
                    },
                },
            },
            {
                input: []byte(`*3\r\n\$3\r\nset\r\n\$5\r\nmykey\r\n\$1\r\n1\r\n`),
                result: RObj{
                    Type: ARRAYS,
                    nextRObj: &RObj{
                        Type:    SIMPLESTRINGS,
                        Content: "set",
                        nextRObj: &RObj{
                            Type:    SIMPLESTRINGS,
                            Content: "mykey",
                            nextRObj: &RObj{
                                Type:    INTEGERS,
                                Content: "1",
                            },
                        },
                    },
                },
            },
        }

        for _, tc := range testCases {
            _, err := Deserialize(tc.input)
            if err != nil {
                t.Errorf("error unmarshalling: got error %v.\n", err)
            }
        }
    })
}
