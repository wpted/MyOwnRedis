package redisObject

import (
    "errors"
    "strings"
)

const (
    NULL          = "null"
    SimpleStrings = "+"
    SimplerErrors = "-"
    Integers      = ":"
    BulkStrings   = "$"
    Arrays        = "*"
)

var ErrInvalidCommand = errors.New("error invalid command")

var cmdTable = map[string]int{
    "null":   1,
    "ping":   1,
    "echo":   2,
    "quit":   1,
    "set":    3,
    "get":    2,
    "exists": 2,
    "delete": 2,
    "incr":   2,
    "decr":   2,
    "save":   1,
    "bgsave": 1,
}

type RObj struct {
    Type    string
    Command string
    Content []string
}

// New deserializes the client request and creates a RObj.
func New(rType string, content []string, cmd string) *RObj {
    return &RObj{Type: rType, Content: content, Command: cmd}
}

// Deserialize decodes bytes into RObjs.
func Deserialize(req []byte) (*RObj, error) {
    if len(req) < 2 {
        return nil, ErrInvalidCommand
    }

    var robj = new(RObj)
    switch req[0] {
    // 1. SimpleErrors, SimpleStrings
    case '+', '-':
        robj.Type = string(req[0])
        robj.Content = []string{string(req[1 : len(req)-2])}

    // 2. Arrays, BulkStrings
    case '*', '$':
        content := make([]string, 0)
        if len(req) != 0 && string(req[1:3]) == "-1" {
            // a. null values
            robj.Type = NULL
            robj.Command = NULL
        } else {
            // b. Parse contents.
            robj.Type = string(req[0])

            switch robj.Type {
            case BulkStrings:
                msgLength, theRestOfTheInput, err := parseLength(req)
                if err != nil {
                    return nil, err
                }

                req = theRestOfTheInput

                msg, theRestOfTheInput, err := parseMessage(req)
                if err != nil {
                    return nil, err
                }
                req = theRestOfTheInput

                if len(msg) != msgLength {
                    return nil, ErrInvalidCommand
                }
                content = append(content, msg)
            case Arrays:
                // 1. How many elements do we have in the array
                elementNumber, theRestOfTheInput, err := parseLength(req)
                if err != nil {
                    return nil, err
                }
                req = theRestOfTheInput

                cmdLength, theRestOfTheInput, err := parseLength(req)
                if err != nil {
                    return nil, err
                }

                req = theRestOfTheInput

                cmd, theRestOfTheInput, err := parseMessage(req)
                if err != nil {
                    return nil, err
                }
                req = theRestOfTheInput

                if len(cmd) != cmdLength {
                    return nil, ErrInvalidCommand
                }

                robj.Command = strings.ToLower(cmd)
                if expectedArgs, ok := cmdTable[robj.Command]; !ok {
                    return nil, ErrInvalidCommand
                } else {
                    if elementNumber != expectedArgs {
                        return nil, ErrInvalidCommand
                    }
                }

                for len(req) != 0 {
                    msgLength, theRestOfTheInput, err := parseLength(req)
                    if err != nil {
                        return nil, err
                    }
                    req = theRestOfTheInput

                    msg, theRestOfTheInput, err := parseMessage(req)
                    if err != nil {
                        return nil, err
                    }
                    req = theRestOfTheInput

                    if len(msg) != msgLength {
                        return nil, ErrInvalidCommand
                    }
                    content = append(content, msg)
                }

                if len(content) != cmdTable[robj.Command]-1 {
                    return nil, ErrInvalidCommand
                }
            }

            robj.Content = content
        }

    default:
        return nil, ErrInvalidCommand
    }

    return robj, nil
}

func parseLength(input []byte) (int, []byte, error) {
    length := 0

    // Move the pointer to the next character from '$', '*' or ':' to the number part of the input.
    input = input[1:]

    // Loop until the character pointed to by p is '\r'
    for input[0] != '\r' {
        length = (length * 10) + int(input[0]-'0')
        input = input[1:]
    }

    // Remove '\r\n'.
    if len(input) < 2 {
        return 0, []byte{}, ErrInvalidCommand
    } else {
        theRestOfTheInput := input[2:]
        return length, theRestOfTheInput, nil
    }
}

func parseMessage(input []byte) (string, []byte, error) {
    messageByteArr := make([]byte, 0)
    for input[0] != '\r' {
        messageByteArr = append(messageByteArr, input[0])
        input = input[1:]
    }

    // Remove '\r\n'.
    if len(input) < 2 {
        return "", []byte{}, ErrInvalidCommand
    } else {
        theRestOfTheInput := input[2:]
        return string(messageByteArr), theRestOfTheInput, nil
    }
}
