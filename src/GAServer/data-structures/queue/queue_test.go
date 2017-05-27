package queue

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	q := New()

	if !q.isEmpty() ||
		q.len != 0 ||
		q.Len() != 0 {
		t.Error()
	}

	q.Push(1)
	q.Push(2)
	q.Push(3)

	if q.queue[0] != 1 ||
		q.queue[1] != 2 ||
		q.queue[2] != 3 {
		fmt.Println(q.queue)
		t.Error()
	}

	if q.Len() != 3 {
		t.Error()
	}

	a := q.Pop()

	if a != 1 || q.Len() != 2 {
		t.Error()
	}

	b := q.Peek()

	if b != 2 {
		t.Error()
	}

	q2 := New()
	q2.Push(0)
	q2.Push(1)
	q2.Push(2)
	q2.Push(3)

	f := func(o interface{}) bool {
		return o.(int) == 3
	}
	q2.PopRule(f)
	fmt.Println(q2.queue)
}
