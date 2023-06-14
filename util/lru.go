package util

import "sync"

const kMaxNum = 10

type Cache interface {
	Find(key string) interface{}
	Insert(key string, value interface{})
	Remove(key string)
}

type Node struct {
	key   string
	value interface{}
	next  *Node
	prev  *Node
}

func NewLRUCache(max int) Cache {
	c := &lruCache{
		max:     max,
		nodeMap: make(map[string]*Node),
	}

	c.head.next = &c.head
	c.head.prev = &c.head
	return c
}

func insertToFirst(head, node *Node) {
	if head.next != nil {
		head.next = node
		head.prev = node
		node.next = head
		node.prev = head
	} else {
		// insert to first
		head.next.prev = node
		node.prev = head
		node.next = head.next
		head.next = node
	}
}

func moveToFirst(head, node *Node) {
	prev := node.prev
	next := node.next

	prev.next = next
	next.prev = prev

	insertToFirst(head, node)
}

func delFrom(head, node *Node) {
	prev := node.prev
	next := node.next

	prev.next = next
	next.prev = prev
}

func delEnd(head *Node) {
	last := head.prev
	lastPrev := last.prev
	head.prev = lastPrev
	lastPrev.next = head
}

type lruCache struct {
	head    Node
	num     int
	max     int
	nodeMap map[string]*Node
}

func (c *lruCache) Find(key string) interface{} {
	if node, ok := c.nodeMap[key]; ok {
		moveToFirst(&c.head, node)
		return node.value
	}
	return nil
}

func (c *lruCache) Insert(key string, value interface{}) {
	if node, ok := c.nodeMap[key]; ok {
		node.value = value
		moveToFirst(&c.head, node)
		return
	}

	node := &Node{
		key:   key,
		value: value,
	}
	insertToFirst(&c.head, node)
	c.num++
	for c.num > c.max {
		delEnd(&c.head)
		c.num--
	}
	c.nodeMap[key] = node
}

func (c *lruCache) Remove(key string) {
	if node, ok := c.nodeMap[key]; ok {
		delete(c.nodeMap, key)
		delFrom(&c.head, node)
	}
}

type mutexCache struct {
	mu  *sync.Mutex
	lru Cache
}

func NewMutexLRUCache(max int) Cache {
	c := &mutexCache{
		mu:  &sync.Mutex{},
		lru: NewLRUCache(max),
	}

	return c
}

func (c *mutexCache) Find(key string) interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lru.Find(key)
}

func (c *mutexCache) Insert(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lru.Insert(key, value)
}

func (c *mutexCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lru.Remove(key)
}
