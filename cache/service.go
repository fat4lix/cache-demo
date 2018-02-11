package cache


type CacheService struct {
	caches map[string]*Cache
}

type CacheInfo struct {
	Name string
	ItemCount int
}

// NewCacheService create new service wich store caches
func NewCacheService() *CacheService {
	s := &CacheService{}
	s.caches = map[string]*Cache{}
	return s
}

// Add new cache with key and config
func (s *CacheService) Add(key string, c Config) {
	if _, ok := s.caches[key]; !ok {
		s.caches[key] = NewCache(c)
	}
}

// Has check is exsist cache by given key
func (s *CacheService) Has(key string) bool {
	if _, ok := s.caches[key]; ok {
		return true
	}
	return false
}

// Get cache by given key
func (s *CacheService) Get(key string) *Cache {
	val, _ := s.caches[key] 
	return val
}

// List return list of caches
func (s *CacheService) List() []*CacheInfo{
	list := []*CacheInfo{}
	for key, val := range s.caches {
		cinf := &CacheInfo{key,len(val.items)}
		list = append(list, cinf)
	}
	return list
}

// Del is delete cache by given key
func (s *CacheService) Del(key string) bool {
	if _, ok := s.caches[key]; ok {
		delete(s.caches, key)
		return true
	}
	return false
}