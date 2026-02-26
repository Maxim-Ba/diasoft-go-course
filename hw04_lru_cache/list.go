package hw04lrucache

import "sync"

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	front  *ListItem
	back   *ListItem
	length int
	mu     *sync.Mutex
}

func NewList() List {
	var mu sync.Mutex
	return &list{
		mu: &mu,
	}
}

func (l *list) Len() int {
	return l.length
}

func (l *list) Front() *ListItem {
	return l.front
}

func (l *list) Back() *ListItem {
	return l.back
}

func (l *list) PushFront(v interface{}) *ListItem {
	l.mu.Lock()
	defer l.mu.Unlock()

	newfront := ListItem{
		Value: v,
		Next:  l.front,
	}
	if l.front == nil {
		l.front = &newfront
		if l.back == nil {
			l.back = &newfront
		}
	} else {
		l.front.Prev = &newfront
		l.front = &newfront
	}
	l.length++
	return l.front
}

func (l *list) PushBack(v interface{}) *ListItem {
	l.mu.Lock()
	defer l.mu.Unlock()

	newback := ListItem{
		Value: v,
		Prev:  l.back,
	}

	if l.back == nil {
		l.back = &newback
		if l.front == nil {
			l.front = &newback
		}
	} else {
		l.back.Next = &newback
		l.back = &newback
	}
	l.length++

	return l.back
}

func (l *list) Remove(i *ListItem) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.collapse(i)
	l.length--
}

func (l *list) MoveToFront(i *ListItem) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.collapse(i)
	i.Prev = nil
	i.Next = l.front
	l.front.Prev = i
	l.front = i
}

func (l *list) collapse(i *ListItem) {
	prev := i.Prev
	next := i.Next
	if prev != nil {
		prev.Next = next
	}
	if next != nil {
		next.Prev = prev
	}
	if i == l.front {
		l.front = i.Next
	}
	if i == l.back {
		l.back = i.Prev
	}
}
