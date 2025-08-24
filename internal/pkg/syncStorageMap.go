package pkg

import (
	"HelaList/internal/driver"
	"sync"
)

/*
Storage需要一个异步Map,该Map实现挂载点路径，与对应Driver类型的映射。
需要异步，所以在这里设计，后续可能打包成pkg吧。
*/

type SyncStorageMap struct {
	mu   sync.RWMutex
	data map[string]driver.Driver
}

func NewSyncStorageMap() *SyncStorageMap {
	return &SyncStorageMap{
		data: make(map[string]driver.Driver),
	}
}

func (m *SyncStorageMap) Store(key string, value driver.Driver) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

func (m *SyncStorageMap) Load(key string) (driver.Driver, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.data[key]
	return value, ok
}

func (m *SyncStorageMap) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

func (m *SyncStorageMap) Has(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.data[key]
	return ok
}

func (m *SyncStorageMap) Values() []driver.Driver {
	m.mu.Lock()
	defer m.mu.Unlock()
	values := make([]driver.Driver, 0, len(m.data))
	for _, v := range m.data {
		values = append(values, v)
	}
	return values
}

// 遍历用，之所以起名叫Range是因为for ... range
func (m *SyncStorageMap) Range(fn func(string, driver.Driver) bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, v := range m.data {
		if !fn(k, v) {
			break
		}
	}
}
