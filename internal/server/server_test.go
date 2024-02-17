package server

import (
    "MyOwnRedis/internal/database/inMemoryDatabase"
    "bytes"
    "errors"
    "io"
    "net"
    "testing"
)

const TestAddr = "localhost:6380"

func TestRedisServer_Run(t *testing.T) {
    db := inMemoryDatabase.New()
    rs := New(TestAddr, db)
    go func() {
        err := rs.Run()
        if err != nil {
            panic(err)
        }
        defer func() {
            if err = rs.Close(); err != nil {
                panic(err)
            }
        }()
    }()

    clientConn, err := net.Dial(TCP, TestAddr)
    if err != nil {
        t.Errorf("error cannot connect to server: %#v\n", err)
    }

    testCases := []struct {
        request  []byte
        response []byte
    }{
        {
            request:  []byte("*1\r\n$4\r\nPING\r\n"),
            response: []byte("+PONG\r\n"),
        },
        {
            request:  []byte("*3\r\n$4\r\necho\r\n$5\r\nhello\r\n$5\r\nworld\r\n"),
            response: []byte("+hello world\r\n"),
        },
    }

    for _, tc := range testCases {
        if _, err = clientConn.Write(tc.request); err != nil {
            t.Errorf("error receive request error: %#v.\n", err)
        }

        resp := make([]byte, 1024)
        if _, err = clientConn.Read(resp); err != nil {
            if !errors.Is(err, io.EOF) {
                t.Errorf("error reading from connection:%#v.\n", err)
            }
        } else {
            if bytes.Compare(resp[:len(tc.response)], tc.response) != 0 {
                t.Errorf("error response didn't match, expected %s, got %s.\n", tc.response, resp[:len(tc.response)])
            }
        }
    }

    err = clientConn.Close()
    if err != nil {
        panic(err)
    }
}
