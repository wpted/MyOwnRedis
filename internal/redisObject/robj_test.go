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
                input: []byte(`$-1\r\n`), // NULL value
            },
            {
                input: []byte(`*-1\r\n`), // NULL value
            },
            {
                input: []byte(`-Error message\r\n`), // Errors.
            },
            {
                input: []byte(`+OK\r\n`), // Simple strings.
            },
            {
                input: []byte(`+hello world\r\n`),
            },
            {
                input: []byte(`$0\r\n`), // Integers.
            },
            {
                input: []byte(`*1\r\n\$4\r\nPING\r\n`), // Arrays.
            },
            {
                input: []byte(`*2\r\n$4\r\necho\r\n$11\r\nhello world\r\n`),
            },
            {
                input: []byte(`*2\r\n$3\r\nget\r\n$3\r\nkey\r\n`),
            },
            {
                input: []byte(`*3\r\n\$3\r\nset\r\n\$5\r\nmykey\r\n\$1\r\n1\r\n`),
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
