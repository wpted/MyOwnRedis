package redisObject

import (
    "errors"
    "testing"
)

func Test_Deserialize(t *testing.T) {
    t.Run("Test Deserialize Invalid commands", func(t *testing.T) {
        testCases := []struct {
            input []byte
            err   error
        }{
            {
                input: []byte("$123\r"), // Arrays.
                err:   ErrInvalidCommand,
            },
            {
                input: []byte("*1\r\n$3\r\nPNG\r\n"), // Arrays.
                err:   ErrInvalidCommand,
            },
        }

        for _, tc := range testCases {
            _, err := Deserialize(tc.input)
            if !errors.Is(err, ErrInvalidCommand) {
                t.Errorf("error incorrect error: expected %v, got %v.\n", ErrInvalidCommand, err)
            }
        }
    })

    t.Run("Test Deserialize Valid commands", func(t *testing.T) {
        testCases := []struct {
            input  []byte
            result *RObj
        }{ //RESP
            {
                input:  []byte("$-1\r\n"), // NULL value.
                result: &RObj{Type: NULL, Command: NULL},
            },
            {
                input:  []byte("*-1\r\n"), // NULL value.
                result: &RObj{Type: NULL, Command: NULL},
            },
            {
                input:  []byte("-Error message\r\n"), // Errors.
                result: &RObj{Type: SimplerErrors, Content: []string{"Error message"}},
            },
            {
                input:  []byte("+OK\r\n"), // Simple strings.
                result: &RObj{Type: SimpleStrings, Content: []string{"OK"}},
            },
            {
                input:  []byte("+hello world\r\n"),
                result: &RObj{Type: SimpleStrings, Content: []string{"hello world"}},
            },
            {
                input:  []byte("$0\r\n\r\n"), // Empty string encoding -> ""
                result: &RObj{Type: BulkStrings, Content: []string{""}},
            },
            {
                input:  []byte("*1\r\n$4\r\nPING\r\n"), // Arrays: PING.
                result: &RObj{Type: Arrays, Command: "ping"},
            },
            {
                input:  []byte("*2\r\n$4\r\necho\r\n$11\r\nhello world\r\n"), // Arrays: ECHO.
                result: &RObj{Type: Arrays, Command: "echo", Content: []string{"hello world"}},
            },
            {
                input:  []byte("*2\r\n$3\r\nget\r\n$3\r\nkey\r\n"), // Arrays: GET.
                result: &RObj{Type: Arrays, Command: "get", Content: []string{"key"}},
            },
            {
                input: []byte("*3\r\n$3\r\nset\r\n$5\r\nmykey\r\n$1\r\n1\r\n"), // Arrays: SET.
                result: &RObj{Type: Arrays, Command: "set", Content: []string{"mykey", "1"},
                },
            },
            {
                input: []byte("*5\r\n$3\r\nset\r\n$5\r\nmykey\r\n$1\r\n1\r\n$2\r\nEX\r\n$2\r\n12\r\n"), // Arrays: SET with EX.
                result: &RObj{Type: Arrays, Command: "set", Content: []string{"mykey", "1", "EX", "12"},
                },
            },
            {
                input: []byte("*5\r\n$3\r\ndel\r\n$5\r\nmykey\r\n$1\r\n1\r\n$5\r\nhello\r\n$3\r\nfoo\r\n"), // Arrays: DEL keys.
                result: &RObj{Type: Arrays, Command: "del", Content: []string{"mykey", "1", "hello", "foo"},
                },
            },
            {
                input: []byte("*2\r\n$6\r\nexists\r\n$1\r\nx\r\n"), // Arrays: EXISTS.
                result: &RObj{Type: Arrays, Command: "exists", Content: []string{"x"},
                },
            },
            {
                input: []byte("*2\r\n$4\r\nincr\r\n$5\r\nmykey\r\n"), // Arrays: INCR.
                result: &RObj{Type: Arrays, Command: "incr", Content: []string{"mykey"},
                },
            },
            {
                input: []byte("*2\r\n$4\r\ndecr\r\n$5\r\nmykey\r\n"), // Arrays: DECR.
                result: &RObj{Type: Arrays, Command: "decr", Content: []string{"mykey"},
                },
            },
            {
                input: []byte("*4\r\n$5\r\nlpush\r\n$5\r\nmykey\r\n$1\r\n1\r\n$5\r\nhello\r\n"), // Arrays: LPUSH.
                result: &RObj{Type: Arrays, Command: "lpush", Content: []string{"mykey", "1", "hello"},
                },
            },
            {
                input: []byte("*4\r\n$5\r\nrpush\r\n$5\r\nmykey\r\n$1\r\n1\r\n$5\r\nhello\r\n"), // Arrays: RPUSH.
                result: &RObj{Type: Arrays, Command: "rpush", Content: []string{"mykey", "1", "hello"},
                },
            },
            {
                input: []byte("*4\r\n$6\r\nlrange\r\n$5\r\nmykey\r\n$1\r\n1\r\n$1\r\n2\r\n"), // Arrays: LRANGE.
                result: &RObj{Type: Arrays, Command: "lrange", Content: []string{"mykey", "1", "2"},
                },
            },
        }

        for _, tc := range testCases {
            robj, err := Deserialize(tc.input)
            if err != nil {
                t.Errorf("error deserializing: got error %v.\n", err)
            }

            // Check the command type.
            if robj.Type != tc.result.Type {
                t.Errorf("error deserializing - incorrect type: expected %s, got %s.\n", tc.result.Type, robj.Type)
            }

            // Check the command.
            if robj.Command != tc.result.Command {
                t.Errorf("error deserializing - incorrect command: expected %s, got %s.\n", tc.result.Command, robj.Command)
            }

            // Check if the length of the content is as expected.

            if tc.result.Command != "" && len(robj.Content) != len(tc.result.Content) {
                t.Errorf("error deserializing - incorrect content length: expected %d, got %d.\n", len(tc.result.Content), len(robj.Content))
            } else {
                // Compare the contents in robj.
                for n, c := range robj.Content {
                    if c != tc.result.Content[n] {
                        t.Errorf("error deserializing - incorrect content: expected %s, got %s.\n", tc.result.Content[n], c)
                    }
                }
            }
        }
    })
}

func Test_parseLength(t *testing.T) {
    testCases := []struct {
        input         []byte
        result        int
        residualBytes []byte
    }{
        {input: []byte("$123\r\n"), result: 123},
        {input: []byte("*3\r\n"), result: 3},
        {input: []byte("*1\r\nPING\r\n"), result: 1, residualBytes: []byte("PING\r\n")},
        {input: []byte("*2\r\n$3\r\nget\r\n$3\r\nkey\r\n"), result: 2, residualBytes: []byte("$3\r\nget\r\n$3\r\nkey\r\n")},
    }

    for _, tc := range testCases {
        re, theRestOfTheInput, err := parseLength(tc.input)
        if err != nil {
            t.Errorf("Error parsing length: got error %v.\n", err)
        }
        if re != tc.result {
            t.Errorf("Error parsing length: expected %d, got %d.\n", tc.result, re)
        }
        if len(theRestOfTheInput) != len(tc.residualBytes) {
            t.Errorf("Error re-slicing, expected length %d, go %d.\n", len(tc.residualBytes), len(theRestOfTheInput))
        } else {
            for n, char := range theRestOfTheInput {
                if char != tc.residualBytes[n] {
                    t.Errorf("Error re-slicing, incorrect characters: expected %s, got %s.\n", string(tc.residualBytes[n]), string(char))
                }
            }
        }
    }
}

func Test_parseMessage(t *testing.T) {
    testCases := []struct {
        input         []byte
        result        string
        residualBytes []byte
    }{
        {input: []byte("get\r\n$3\r\nkey\r\n"), result: "get", residualBytes: []byte("$3\r\nkey\r\n")},
        {input: []byte("1\r\n"), result: "1"},
        {input: []byte("PING\r\n"), result: "PING"},
        {input: []byte("echo\r\n$11\r\nhello world\r\n"), result: "echo", residualBytes: []byte("$11\r\nhello world\r\n")},
    }

    for _, tc := range testCases {
        msg, theRestOfTheInput, err := parseMessage(tc.input)
        if err != nil {
            t.Errorf("Error parsing message: got error %v.\n", err)
        }
        if msg != tc.result {
            t.Errorf("Error parsing message: expected %s, got %s.\n", tc.result, msg)
        }

        if len(theRestOfTheInput) != len(tc.residualBytes) {
            t.Errorf("Error re-slicing, expected length %d, go %d.\n", len(tc.residualBytes), len(theRestOfTheInput))
        } else {
            for n, char := range theRestOfTheInput {
                if char != tc.residualBytes[n] {
                    t.Errorf("Error re-slicing, incorrect characters: expected %s, got %s.\n", string(tc.residualBytes[n]), string(char))
                }
            }
        }
    }
}
