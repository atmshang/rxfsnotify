package concurrent

import (
	"sync"
)

type SafeMap struct {
	mu sync.RWMutex
	m  map[string]bool
}

func NewSafeMap() *SafeMap {
	return &SafeMap{
		m: make(map[string]bool),
	}
}

func (sm *SafeMap) Set(key string, value bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.m[key] = value
}

func (sm *SafeMap) Get(key string) (bool, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	value, ok := sm.m[key]
	return value, ok
}

func (sm *SafeMap) ToList() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var result []string
	for key, value := range sm.m {
		if value {
			result = append(result, key)
		}
	}
	return result
}
