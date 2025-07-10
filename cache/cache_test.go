package cache_test

import (
	"math"
	"slices"
	"testing"
	"time"

	"github.com/database64128/cubic-go-playground/cache"
)

func TestExpirationCache(t *testing.T) {
	c := cache.NewExpirationCache[int, int](3)
	now := time.Now()
	assertExpirationCacheLenCapacityContent(t, c, nil, 3, now)
	c.SetFromHead(1, -1, now, now.Add(time.Second))
	assertExpirationCacheLenCapacityContent(t, c, []cache.Entry[int, int]{{1, -1}}, 3, now)
	c.SetFromHead(2, -2, now, now.Add(2*time.Second))
	assertExpirationCacheLenCapacityContent(t, c, []cache.Entry[int, int]{{1, -1}, {2, -2}}, 3, now)
	c.SetFromTail(3, -3, now, now.Add(4*time.Second))
	c.SetFromTail(4, -4, now, now.Add(3*time.Second))
	assertExpirationCacheLenCapacityContent(t, c, []cache.Entry[int, int]{{2, -2}, {4, -4}, {3, -3}}, 3, now)
	if !c.Remove(4) {
		t.Error("c.Remove(4) = false, want true")
	}
	assertExpirationCacheLenCapacityContent(t, c, []cache.Entry[int, int]{{2, -2}, {3, -3}}, 3, now)
	now = now.Add(2 * time.Second)
	c.SetFromTail(3, 3, now, now.Add(5*time.Second))
	assertExpirationCacheLenCapacityContent(t, c, []cache.Entry[int, int]{{3, 3}}, 3, now)
	c.Clear()
	assertExpirationCacheLenCapacityContent(t, c, nil, 3, now)
}

func TestExpirationCacheUnboundedCapacity(t *testing.T) {
	c := cache.NewExpirationCache[int, int](0)
	now := time.Now()
	assertExpirationCacheLenCapacityContent(t, c, nil, math.MaxInt, now)
	c.SetFromTail(1, -1, now, now.Add(3*time.Second))
	c.SetFromHead(2, -2, now, now.Add(2*time.Second))
	c.SetFromTail(3, -3, now, now.Add(time.Second))
	for range c.All() {
		for range c.Backward() {
			break
		}
		break
	}
	c.SetFromTail(4, -4, now, now.Add(6*time.Second))
	c.SetFromTail(5, -5, now, now.Add(5*time.Second))
	c.SetFromHead(6, -6, now, now.Add(4*time.Second))
	assertExpirationCacheLenCapacityContent(t, c, []cache.Entry[int, int]{
		{3, -3}, {2, -2}, {1, -1}, {6, -6}, {5, -5}, {4, -4},
	}, math.MaxInt, now)
}

func assertExpirationCacheLenCapacityContent(t *testing.T, c *cache.ExpirationCache[int, int], want []cache.Entry[int, int], expectedCapacity int, now time.Time) {
	t.Helper()

	if got := c.Len(); got != len(want) {
		t.Errorf("c.Len() = %d, want %d", got, len(want))
	}
	if got := c.Capacity(); got != expectedCapacity {
		t.Errorf("c.Capacity() = %d, want %d", got, expectedCapacity)
	}

	got := make([]cache.Entry[int, int], 0, len(want))
	for key, value := range c.All() {
		got = append(got, cache.Entry[int, int]{Key: key, Value: value})
	}
	if !slices.Equal(got, want) {
		t.Errorf("c.All() = %v, want %v", got, want)
	}

	got = got[:0]
	for key, value := range c.Backward() {
		got = append(got, cache.Entry[int, int]{Key: key, Value: value})
	}
	if !slicesReverseEqual(got, want) {
		t.Errorf("c.Backward() = %v, want %v", got, want)
	}

	for key := range 10 {
		if index := slices.IndexFunc(want, func(e cache.Entry[int, int]) bool {
			return e.Key == key
		}); index != -1 {
			expectedEntry := want[index]
			expectedValue := expectedEntry.Value

			if !c.Contains(key) {
				t.Errorf("c.Contains(%d) = false, want true", key)
			}
			if !c.TryContains(key) {
				t.Errorf("c.TryContains(%d) = false, want true", key)
			}

			value, ok := c.Get(key, now)
			if value != expectedValue || !ok {
				t.Errorf("c.Get(%d, %v) = %d, %v, want %d, true", key, now, value, ok, expectedValue)
			}

			entry, ok := c.GetEntry(key, now)
			if entry == nil || *entry != expectedEntry || !ok {
				t.Errorf("c.GetEntry(%d, %v) = %v, %v, want {Key: %d, Value: %d}, true", key, now, entry, ok, expectedEntry.Key, expectedEntry.Value)
			}
		} else {
			if c.Contains(key) {
				t.Errorf("c.Contains(%d) = true, want false", key)
			}
			if c.TryContains(key) {
				t.Errorf("c.TryContains(%d) = true, want false", key)
			}

			value, ok := c.Get(key, now)
			if value != 0 || ok {
				t.Errorf("c.Get(%d, %v) = %d, %v, want 0, false", key, now, value, ok)
			}

			entry, ok := c.GetEntry(key, now)
			if entry != nil || ok {
				t.Errorf("c.GetEntry(%d, %v) = %v, %v, want nil, false", key, now, entry, ok)
			}

			if c.Remove(key) {
				t.Errorf("c.Remove(%d) = true, want false", key)
			}
		}
	}
}

func slicesReverseEqual[S ~[]E, E comparable](s1, s2 S) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[len(s2)-1-i] {
			return false
		}
	}
	return true
}
