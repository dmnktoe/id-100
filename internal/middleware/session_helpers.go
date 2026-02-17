package middleware

import (
	"math"
	"strconv"
	"time"
)

// Session key constants
const (
	SessionKeyPlayerName   = "player_name"
	SessionKeyPlayerCity   = "player_city"
	SessionKeyToken        = "token"
	SessionKeyTokenID      = "token_id"
	SessionKeyBagName      = "bag_name"
	SessionKeySessionNum   = "session_number"
	SessionKeySessionStart = "session_started_at"
)

// GetSessionNumber safely converts a session value to int
func GetSessionNumber(v interface{}) (int, bool) {
	// Determine platform int bounds using strconv.IntSize
	bits := strconv.IntSize
	maxInt64 := int64(1<<(bits-1) - 1)
	minInt64 := -int64(1 << (bits - 1))

	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		if n >= minInt64 && n <= maxInt64 {
			return int(n), true
		}
		return 0, false
	case float64:
		if math.IsNaN(n) || math.IsInf(n, 0) {
			return 0, false
		}
		if math.Trunc(n) != n {
			return 0, false
		}
		// Ensure the float value fits in the platform int range before converting
		if n < float64(minInt64) || n > float64(maxInt64) {
			return 0, false
		}
		asInt64 := int64(n)
		return int(asInt64), true
	case string:
		if x, err := strconv.Atoi(n); err == nil {
			return x, true
		}
	}
	return 0, false
}

// GetSessionTime safely converts a session value to time.Time
func GetSessionTime(v interface{}) (time.Time, bool) {
	// Compute safe Unix second bounds such that sec*1e9 doesn't overflow int64
	maxInt64 := int64(^uint64(0) >> 1)
	minInt64 := -maxInt64 - 1
	maxSec := maxInt64 / 1e9
	minSec := minInt64 / 1e9

	switch t := v.(type) {
	case time.Time:
		return t, true
	case string:
		if tm, err := time.Parse(time.RFC3339, t); err == nil {
			return tm, true
		}
	case int64:
		if t >= minSec && t <= maxSec {
			return time.Unix(t, 0), true
		}
		return time.Time{}, false
	case int:
		sec := int64(t)
		if sec >= minSec && sec <= maxSec {
			return time.Unix(sec, 0), true
		}
		return time.Time{}, false
	case float64:
		if math.IsNaN(t) || math.IsInf(t, 0) {
			return time.Time{}, false
		}
		if math.Trunc(t) != t {
			return time.Time{}, false
		}
		sec := int64(t)
		if sec >= minSec && sec <= maxSec {
			return time.Unix(sec, 0), true
		}
		return time.Time{}, false
	}
	return time.Time{}, false
}
