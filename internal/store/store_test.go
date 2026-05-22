package store

import (
	"os"
	"testing"
)

func TestStore_BasicOperations(t *testing.T) {
	f, err := os.CreateTemp("", "wal-*.log")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(f.Name())
	defer f.Close()

	s, err := NewStore(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	err = s.Put("username", "brian")
	if err != nil {
		t.Fatal(err)
	}

	val, ok := s.Get("username")
	if !ok || val != "brian" {
		t.Errorf("expected brian, got %s", val)
	}

	ok = s.Delete("username")
	if !ok {
		t.Errorf("expected delete to return true")
	}

	_, ok = s.Get("username")
	if ok {
		t.Errorf("expected key to be gone after delete")
	}
}

func TestStore_WALReplay(t *testing.T) {
	f, err := os.CreateTemp("", "wal-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	s, err := NewStore(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	s.Put("username", "brian")
	s.Put("othername", "joe")
	s.Delete("username")

	s2, err := NewStore(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	_, ok := s2.Get("username")
	if ok {
		t.Errorf("expected username to be deleted after replay")
	}

	val, ok := s2.Get("othername")
	if !ok || val != "joe" {
		t.Errorf("expected othername=joe after replay, got %s", val)
	}
}

func BenchmarkStore_Put(b *testing.B) {
	f, err := os.CreateTemp("", "wal-*.log")
	if err != nil {
		b.Fatal(err)
	}

	defer os.Remove(f.Name())
	defer f.Close()

	s, err := NewStore(f.Name())
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Put("key", "value")
	}
}

func BenchmarkStore_Get(b *testing.B) {
	f, err := os.CreateTemp("", "wal-*.log")
	if err != nil {
		b.Fatal(err)
	}

	defer os.Remove(f.Name())
	defer f.Close()

	s, err := NewStore(f.Name())
	if err != nil {
		b.Fatal(err)
	}

	s.Put("key", "value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Get("key")
	}
}

func BenchmarkStore_ConcurrentGet(b *testing.B) {
	f, err := os.CreateTemp("", "wal-*.log")
	if err != nil {
		b.Fatal(err)
	}

	defer os.Remove(f.Name())
	defer f.Close()

	s, err := NewStore(f.Name())
	if err != nil {
		b.Fatal(err)
	}

	s.Put("key", "value")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.Get("key")
		}
	})
}
