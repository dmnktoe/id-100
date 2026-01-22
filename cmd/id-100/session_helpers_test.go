package main

import (
	"math"
	"strconv"
	"testing"
	"time"
)

func TestGetSessionNumber(t *testing.T) {
	// simple int
	if n, ok := getSessionNumber(5); !ok || n != 5 {
		t.Fatalf("expected 5, got %v (ok=%v)", n, ok)
	}

	// int64 within platform bounds
	if n, ok := getSessionNumber(int64(7)); !ok || n != 7 {
		t.Fatalf("expected 7, got %v (ok=%v)", n, ok)
	}

	// string numeric
	if n, ok := getSessionNumber("9"); !ok || n != 9 {
		t.Fatalf("expected 9, got %v (ok=%v)", n, ok)
	}

	// non-numeric string
	if _, ok := getSessionNumber("notint"); ok {
		t.Fatalf("expected parse failure")
	}

	// float64 integral
	if n, ok := getSessionNumber(9876.0); !ok || n != 9876 {
		t.Fatalf("float64 integral failed: got %v (ok=%v)", n, ok)
	}

	// float64 fractional should fail
	if _, ok := getSessionNumber(3.14); ok {
		t.Fatalf("float64 fractional should fail")
	}

	// float64 NaN/Inf should fail
	if _, ok := getSessionNumber(math.NaN()); ok {
		t.Fatalf("float64 NaN should fail")
	}
	if _, ok := getSessionNumber(math.Inf(1)); ok {
		t.Fatalf("float64 Inf should fail")
	}

	// sanity: very large float that won't fit in platform int should fail
	if _, ok := getSessionNumber(1e20); ok {
		t.Fatalf("too-large float64 should fail")
	}
}

func TestGetSessionTime(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	if tm, ok := getSessionTime(now); !ok || !tm.Equal(now) {
		t.Fatalf("expected equal time, got %v (ok=%v)", tm, ok)
	}

	// RFC3339 string
	rfc := now.UTC().Format(time.RFC3339)
	if tm, ok := getSessionTime(rfc); !ok || !tm.Equal(now.UTC()) {
		t.Fatalf("RFC3339 parse failed: got %v (ok=%v)", tm, ok)
	}

	// invalid string
	if _, ok := getSessionTime("not-a-time"); ok {
		t.Fatalf("invalid time string should fail")
	}

	// int seconds
	if tm, ok := getSessionTime(int(1600000000)); !ok || tm.Unix() != 1600000000 {
		t.Fatalf("int seconds failed: got %v (ok=%v)", tm, ok)
	}

	// int64 seconds
	if tm, ok := getSessionTime(int64(1600000000)); !ok || tm.Unix() != 1600000000 {
		t.Fatalf("int64 seconds failed: got %v (ok=%v)", tm, ok)
	}

	// float64 integral seconds
	if tm, ok := getSessionTime(1600000000.0); !ok || tm.Unix() != 1600000000 {
		t.Fatalf("float64 integral seconds failed: got %v (ok=%v)", tm, ok)
	}

	// float64 fractional should fail
	if _, ok := getSessionTime(1600000000.5); ok {
		t.Fatalf("float64 fractional should fail")
	}

	// float64 NaN/Inf should fail
	if _, ok := getSessionTime(math.NaN()); ok {
		t.Fatalf("float64 NaN should fail")
	}
	if _, ok := getSessionTime(math.Inf(-1)); ok {
		t.Fatalf("float64 Inf should fail")
	}

	// out-of-range seconds (one more than allowed max)
	maxInt64 := int64(^uint64(0) >> 1)
	maxSec := maxInt64 / 1e9
	out := maxSec + 1
	if _, ok := getSessionTime(out); ok {
		t.Fatalf("out-of-range seconds should fail: %s", strconv.FormatInt(out, 10))
	}
}
