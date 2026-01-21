package main

import (
	"testing"
	"time"
)

func TestGetSessionNumber(t *testing.T) {
	if n, ok := getSessionNumber(5); !ok || n != 5 {
		t.Fatalf("expected 5, got %v (ok=%v)", n, ok)
	}
	if n, ok := getSessionNumber(int64(7)); !ok || n != 7 {
		t.Fatalf("expected 7, got %v (ok=%v)", n, ok)
	}
	if n, ok := getSessionNumber("9"); !ok || n != 9 {
		t.Fatalf("expected 9, got %v (ok=%v)", n, ok)
	}
	if _, ok := getSessionNumber("notint"); ok {
		t.Fatalf("expected parse failure")
	}
}

func TestGetSessionTime(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	if tm, ok := getSessionTime(now); !ok || !tm.Equal(now) {
		t.Fatalf("expected equal time, got %v (ok=%v)", tm, ok)
	}
	rfc := now.Format(time.RFC3339)
	if tm, ok := getSessionTime(rfc); !ok || !tm.Equal(now) {
		t.Fatalf("expected parsed time equal to now, got %v (ok=%v)", tm, ok)
	}
	if _, ok := getSessionTime("not-a-time"); ok {
		t.Fatalf("expected parse failure")
	}
}