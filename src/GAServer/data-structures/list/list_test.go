package list

import (
	"fmt"
	"testing"
)

func TestList(t *testing.T) {
	l := New(1, 2, 3, 4, 5)
	l.Add(6)
	l.Append(7, 8)
	l2 := New(9, 10)
	l.Concat(l2)
	fmt.Println(l)
	fmt.Println(l.Find(func(o interface{}) bool {
		return o.(int) == 7
	}))
	l.Add(7)
	fmt.Println(l.FindAll(func(o interface{}) bool {
		return o.(int) == 7
	}))
	fmt.Println(l.RemoveRule(func(o interface{}) bool {
		return o.(int) == 7
	}))
	fmt.Println(l)
	l.Insert(5, 99)
	fmt.Println(l)

	fmt.Println("remove:", l.Remove(5), l)
	fmt.Println(l.RemoveRule(func(o interface{}) bool {
		return o.(int) == 7
	}))
	fmt.Println("rule:", l)
	l.Insert(0, 99)
	l.Insert(5, 99)
	l.Insert(5, 99)
	l.Insert(999, 99)
	fmt.Println(l)
	fmt.Println(l.RemoveAllRule(func(o interface{}) bool {
		return o.(int) == 99
	}))
	fmt.Println("ruleall:", l)
	l.Each(func(o interface{}) {
		fmt.Println(o.(int) * 2)
	})
	fmt.Println(l)
	l3 := New(88, 88, 99, 99)
	l.DeepCopy(l3)
	fmt.Println("dc", l)
	l.Clear()
	fmt.Println(l)
}
