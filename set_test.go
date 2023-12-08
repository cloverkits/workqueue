package workqueue

import "testing"

func TestSet(t *testing.T) {
	s := make(set)
	if s.has("foo") {
		t.Fatal("has() returned true for non-existent item")
	}
	s.insert("foo")
	if !s.has("foo") {
		t.Fatal("has() returned false for existent item")
	}
	s.delete("foo")
	if s.has("foo") {
		t.Fatal("has() returned true for deleted item")
	}
	if s.len() != 0 {
		t.Fatal("len() returned non-zero for empty set")
	}
	s.insert("foo")
	s.insert("bar")
	if s.len() != 2 {
		t.Fatal("len() returned wrong value")
	}
}
