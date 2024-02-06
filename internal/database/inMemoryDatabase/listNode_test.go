package inMemoryDatabase

import "testing"

func TestStrNode_LeftPush(t *testing.T) {
    testCases := []struct {
        node             *StrNode
        valuesToPushLeft []string
        expectedArr      []string
    }{
        {
            node:             nil,
            valuesToPushLeft: []string{"1"},
            expectedArr:      []string{"1"},
        },
        {
            node:             nil,
            valuesToPushLeft: []string{"1", "2"},
            expectedArr:      []string{"2", "1"},
        },
        {
            node: &StrNode{
                value: "1",
                next:  nil,
            },
            valuesToPushLeft: []string{"2", "3"},
            expectedArr:      []string{"3", "2", "1"},
        },
        {
            node: &StrNode{
                value: "1",
                next: &StrNode{
                    value: "2",
                    next:  nil,
                },
            },
            valuesToPushLeft: []string{"3", "4"},
            expectedArr:      []string{"4", "3", "1", "2"},
        },
    }

    for _, tc := range testCases {
        newHead := tc.node.LeftPush(tc.valuesToPushLeft)
        gotArr := make([]string, 0)
        temp := newHead
        for temp != nil {
            gotArr = append(gotArr, temp.value)
            temp = temp.next
        }

        if len(gotArr) != len(tc.expectedArr) {
            t.Errorf("Error linked list length: expected %d, got %d.\n", len(tc.expectedArr), len(gotArr))
        }

        for n, ele := range gotArr {
            if ele != tc.expectedArr[n] {
                t.Errorf("Error element in linked list: expected %s, got %s.\n", tc.expectedArr[n], ele)
            }
        }
    }
}

func TestStrNode_RightPush(t *testing.T) {
    testCases := []struct {
        node              *StrNode
        valuesToPushRight []string
        expectedArr       []string
    }{
        {
            node:              nil,
            valuesToPushRight: []string{"1"},
            expectedArr:       []string{"1"},
        },
        {
            node:              nil,
            valuesToPushRight: []string{"1", "2"},
            expectedArr:       []string{"1", "2"},
        },
        {
            node: &StrNode{
                value: "1",
                next:  nil,
            },
            valuesToPushRight: []string{"2", "3"},
            expectedArr:       []string{"1", "2", "3"},
        },
        {
            node: &StrNode{
                value: "1",
                next: &StrNode{
                    value: "2",
                    next:  nil,
                },
            },
            valuesToPushRight: []string{"3", "4"},
            expectedArr:       []string{"1", "2", "3", "4"},
        },
    }

    for _, tc := range testCases {
        newHead := tc.node.RightPush(tc.valuesToPushRight)
        gotArr := make([]string, 0)
        temp := newHead
        for temp != nil {
            gotArr = append(gotArr, temp.value)
            temp = temp.next
        }

        if len(gotArr) != len(tc.expectedArr) {
            t.Errorf("Error linked list length: expected %d, got %d.\n", len(tc.expectedArr), len(gotArr))
        }

        for n, ele := range gotArr {
            if ele != tc.expectedArr[n] {
                t.Errorf("Error element in linked list: expected %s, got %s.\n", tc.expectedArr[n], ele)
            }
        }
    }

}

func TestStrNode_Len(t *testing.T) {
    testCases := []struct {
        node   *StrNode
        length int
    }{
        {
            node:   nil,
            length: 0,
        },
        {
            node: &StrNode{
                value: "1",
                next: &StrNode{
                    value: "2",
                    next: &StrNode{
                        value: "3",
                        next: &StrNode{
                            value: "4",
                        },
                    },
                },
            },
            length: 4,
        },
    }

    for _, tc := range testCases {
        if tc.node.Len() != tc.length {
            t.Errorf("Error node length: expected %d, got %d.\n", tc.length, tc.node.Len())
        }
    }
}
