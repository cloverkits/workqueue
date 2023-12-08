package workqueue

// 用 map 实现一个 map[any]struct{} 的 set

// set is a set of any type.
type empty struct{}

// set is a set of any type.
type set map[any]empty

// object in the set.
func (s set) has(i any) bool {
	_, exists := s[i]
	return exists
}

// insert object in the set.
func (s set) insert(i any) {
	s[i] = empty{}
}

// delete object in the set.
func (s set) delete(i any) {
	delete(s, i)
}

// len returns the number of objects in the set.
func (s set) len() int {
	return len(s)
}
