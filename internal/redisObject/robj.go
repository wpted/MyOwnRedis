package redisObject

import (
    "errors"
    "fmt"
    "strconv"
    "strings"
)

const (
    NULL          = "null"
    SimpleStrings = "+"
    SimpleErrors  = "-"
    Integers      = ":"
    BulkStrings   = "$"
    Arrays        = "*"
)

var ErrInvalidCommand = errors.New("error invalid command")

const (
    FIX      = "fix"
    OPTIONAL = "optional"
    MULTIPLE = "multiple"
)

var cmdTable = map[string]struct {
    cmdType      string
    expectedArgs int
}{
    "null":   {},
    "ping":   {cmdType: FIX, expectedArgs: 0},
    "echo":   {cmdType: MULTIPLE, expectedArgs: -1},
    "quit":   {cmdType: FIX, expectedArgs: 0},
    "get":    {cmdType: FIX, expectedArgs: 1},
    "exists": {cmdType: FIX, expectedArgs: 1},
    "incr":   {cmdType: FIX, expectedArgs: 1},
    "decr":   {cmdType: FIX, expectedArgs: 1},
    "save":   {cmdType: FIX, expectedArgs: 0},
    "load":   {cmdType: FIX, expectedArgs: 0},
    "lrange": {cmdType: FIX, expectedArgs: 3},
    "set":    {cmdType: OPTIONAL, expectedArgs: -1},
    "del":    {cmdType: MULTIPLE, expectedArgs: -1},
    "lpush":  {cmdType: MULTIPLE, expectedArgs: -1},
    "rpush":  {cmdType: MULTIPLE, expectedArgs: -1},
}

type RObj struct {
    Type       string
    Command    string
    TimeToLive int
    Content    []string
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
                currentCmd, ok := cmdTable[robj.Command]
                if !ok {
                    // Command doesn't exist.
                    return nil, ErrInvalidCommand
                }

                switch cmdTable[robj.Command].cmdType {
                case FIX:
                    // The numbers of element - 1 should be exactly the same as expectedArgs.
                    if elementNumber-1 != currentCmd.expectedArgs {
                        return nil, ErrInvalidCommand
                    }
                    if currentCmd.expectedArgs != 0 {
                        content, err = parseContent(req)
                        if err != nil {
                            return nil, err
                        }

                        // The content number should be exactly the same as expectedArgs.
                        if len(content) != currentCmd.expectedArgs {
                            return nil, ErrInvalidCommand
                        }
                    }
                case MULTIPLE:
                    content, err = parseContent(req)
                    if err != nil {
                        return nil, err
                    }
                    // There should be at least one element in the contents of a MULTIPLE type command.
                    if len(content) == 0 {
                        return nil, ErrInvalidCommand
                    }

                case OPTIONAL:
                    content, err = parseContent(req)
                    if err != nil {
                        return nil, err
                    }

                    if len(content) == 2 || len(content) == 4 {
                        // For set commands there's optional tags like EX, PX, EAXT, PXAT...
                        if len(content) == 4 {
                            // Check the optional tags
                            optionalCmd := strings.ToLower(content[2])
                            timeArg, err := strconv.Atoi(content[3])
                            if err != nil {
                                // The given argument after tag isn't a string.
                                return nil, ErrInvalidCommand
                            }

                            switch optionalCmd {
                            case "ex":
                                robj.TimeToLive = timeArg
                            case "px":
                            case "eaxt":
                            case "pxat":
                            default:
                                return nil, ErrInvalidCommand
                            }
                        }
                    } else {
                        return nil, ErrInvalidCommand
                    }
                }
            default:
                return nil, ErrInvalidCommand
            }
            robj.Content = content
        }
    default:
        return nil, ErrInvalidCommand
    }

    return robj, nil
}

// parseContent is a helper function that should be called after parsing the command from a RESP format byte.
// The function parses the req two steps a time, with the first step parsing the length of the element and the second step parsing the actual element.
func parseContent(req []byte) ([]string, error) {
    content := make([]string, 0)
    for len(req) != 0 {
        // 1. Parse the length of the current element.
        msgLength, theRestOfTheInput, err := parseLength(req)
        if err != nil {
            return nil, err
        }
        req = theRestOfTheInput

        // 2. Parse the actual element.
        msg, theRestOfTheInput, err := parseMessage(req)
        if err != nil {
            return nil, err
        }
        req = theRestOfTheInput

        // Check if the length of the element matches the given length.
        if len(msg) != msgLength {
            return nil, ErrInvalidCommand
        }
        content = append(content, msg)
    }

    return content, nil
}

// parseMessage reads length from a request.
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

// parseMessage reads message from a request.
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

func Serialize(responseType string, data ...string) []byte {
    var re []byte
    switch responseType {
    case SimpleStrings:
        re = []byte(SimpleStrings)
        for i := 0; i < len(data)-1; i++ {
            // Add spaces between contents if length of data != 1.
            re = append(re, fmt.Sprintf("%s ", data[i])...)
        }

        // Append the last element with the delimiter '\r\n'.
        re = append(re, fmt.Sprintf("%s\r\n", data[len(data)-1])...)
    case SimpleErrors:
        re = []byte(fmt.Sprintf("%s%s\r\n", SimpleErrors, data[0]))
    case Arrays:
        // Count the elements in the array.
        re = []byte(fmt.Sprintf("%s%d\r\n", Arrays, len(data)))
        for _, ele := range data {
            // Count the length of the ele.
            re = append(re, fmt.Sprintf("$%d\r\n%s\r\n", len(ele), ele)...)
        }
    case Integers:
        re = []byte(fmt.Sprintf("%s%s\r\n", Integers, data[0]))
    case BulkStrings:
        re = []byte(fmt.Sprintf("%s%d\r\n%s\r\n", BulkStrings, len(data[0]), data[0]))
    }
    return re
}
