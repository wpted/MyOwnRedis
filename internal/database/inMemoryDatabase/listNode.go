package inMemoryDatabase

// StrNode is a linked list that stores string value.
type StrNode struct {
    value string
    next  *StrNode
}

// NewStrNode creates a new instance of a StrNode.
func NewStrNode(value string) *StrNode {
    return &StrNode{value: value}
}

// LeftPush pushes values to the left of a linked list in reverse order.
// Say we left push ['a', 'b'] to a linked list 'c' -> 'd', the result will be 'b' -> 'a' -> 'c' -> 'd'.
func (s *StrNode) LeftPush(values []string) *StrNode {
    newHead := NewStrNode(values[len(values)-1])

    temp := newHead

    if len(values) >= 1 {
        for i := len(values) - 2; i >= 0; i-- {
            currNode := NewStrNode(values[i])
            temp.next = currNode
            // Advance
            temp = temp.next
        }
    }

    // Assign the original list to the tail.
    if temp == nil {
        temp = s
    } else {
        temp.next = s
    }
    return newHead
}

// RightPush pushes values to the right of a linked list.
// Say we left push ['a', 'b'] to a linked list 'c' -> 'd', the result will be 'c' -> 'd' -> 'a' -> 'b'.
func (s *StrNode) RightPush(values []string) *StrNode {
    // Create a list from values.
    tail := NewStrNode(values[0])
    tempTail := tail
    if len(values) > 1 {
        for _, value := range values[1:] {
            currNode := NewStrNode(value)
            tempTail.next = currNode
            tempTail = tempTail.next
        }
    }

    if s != nil {
        temp := s
        for temp.next != nil {
            temp = temp.next
        }
        temp.next = tail
        return s
    }
    return tail
}

// Len returns the expectedLength of a StrNode linked list.
func (s *StrNode) Len() int {
    var count int

    temp := s
    for temp != nil {
        temp = temp.next
        count++
    }
    return count
}

// arr returns a string array representation of a StrNode.
func (s *StrNode) arr() []string {
    re := make([]string, 0)
    temp := s
    for temp != nil {
        re = append(re, temp.value)
        temp = temp.next
    }
    return re
}

// LRange returns the elements from offsets start and stop.
func (s *StrNode) LRange(start, stop int) []string {
    arr := s.arr()
    listLength := len(arr)

    // Check start index and modify it.
    if start < 0 {
        if start < -listLength {
            start = 0
        } else {
            start += listLength
        }
    }

    if start > listLength {
        return []string{}
    }

    // Check the stop index and modify it.
    if stop < 0 {
        if stop < -listLength {
            stop = 0
        } else {
            stop += listLength
        }
    }

    if stop > listLength {
        stop = listLength - 1
    }

    if stop < start {
        return []string{}
    }

    return arr[start : stop+1]
}
