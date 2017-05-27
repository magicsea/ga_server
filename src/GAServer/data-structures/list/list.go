package list

import (
	"fmt"
)

//not safe list
type List struct {
	data []interface{}
}

func New(data ...interface{}) *List {
	l := &List{}
	l.Append(data...)
	return l
}

//添加一个
func (l *List) Add(v interface{}) {
	l.data = append(l.data, v)
}

//附加一组
func (l *List) Append(v ...interface{}) {
	l.data = append(l.data, v...)
}

//插入
func (l *List) Insert(index int, o interface{}) int {

	if index > len(l.data) {
		index = len(l.data)
	}
	if index < 0 {
		index = 0
	}
	var R []interface{}
	R = append(R, l.data[index:]...)
	l.data = append(l.data[:index], o)
	l.data = append(l.data, R...)
	return index
}

//合并
func (l *List) Concat(k *List) {
	l.data = append(l.data, k.RawList()...)
}

//深拷贝
func (l *List) DeepCopy(k *List) {
	l.data = append(l.data[0:0], k.RawList()...)
}

//按序号移除一个节点
func (l *List) Remove(index int) interface{} {
	if index < 0 || index >= len(l.data) {
		return nil
	}
	v := l.data[index]
	l.data = append(l.data[:index], l.data[index+1:]...)
	return v
}

type RuleFunc func(interface{}) bool

//移除一个节点
func (l *List) RemoveRule(rule RuleFunc) interface{} {
	for index := 0; index < len(l.data); index++ {
		v := l.data[index]
		if rule(v) {
			l.data = append(l.data[:index], l.data[index+1:]...)
			return v
		}
	}
	return nil
}

//移除所有符合条件节点
func (l *List) RemoveAllRule(rule RuleFunc) int {
	var i, c int
	le := len(l.data)
	for {
		if i+c >= le {
			break
		}
		v := l.data[i]
		if rule(v) {
			l.data = append(l.data[:i], l.data[i+1:]...)
			c++
		} else {
			i++
		}
	}

	return c
}

//所有节点执行f函数
func (l *List) Each(f func(o interface{})) {
	for _, v := range l.data {
		f(v)
	}
}

//按规则查找一个
func (l *List) Find(rule RuleFunc) interface{} {
	for _, v := range l.data {
		if rule(v) {
			return v
		}
	}
	return nil
}

//按规则查找所有
func (l *List) FindAll(rule RuleFunc) []interface{} {
	var tempL []interface{}
	for _, v := range l.data {
		if rule(v) {
			tempL = append(tempL, v)
		}
	}
	return tempL
}

//原始列表
func (l *List) RawList() []interface{} {
	return l.data
}

//清理
func (l *List) Clear() {
	l.data = nil
}

func (l *List) String() string {
	return fmt.Sprintf("%v", l.data)
}
