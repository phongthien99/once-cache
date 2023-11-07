package once_cache

import (
	"time"

	"golang.org/x/sync/singleflight"
)

type SingleFunc func() (any, error)

type CatchErrorFunc func(cacheStore ICache, key string, err error) any

// IOnceCache is an interface that extends the ICache interface with a method for getting values with a single function.
type IOnceCache interface {
	ICache
	GetWithSingleFunc(key string, f SingleFunc, d time.Duration, catchError *CatchErrorFunc) (any, bool)
}

// OnceCache is a struct that implements the IOnceCache interface.
type OnceCache struct {
	group *singleflight.Group
	ICache
}

// GetWithSingleFunc retrieves a value associated with a key using a single function to generate the value.
// It ensures that the function is called only once for the same key within the specified time duration.
func (o *OnceCache) GetWithSingleFunc(key string, f SingleFunc, d time.Duration, catchError *CatchErrorFunc) (any, bool) {
	// Attempt to get the value from the cache
	value, ok := o.Get(key)
	if !ok {
		// If not found in the cache, use the singleflight.Group to ensure the function is called only once
		// for the same key, even if multiple goroutines request the same key simultaneously.
		defer o.group.Forget(key)
		value, err, _ := o.group.Do(key, f)

		if err != nil {
			// If an error occurred while executing the function, handle the error and return false.
			if catchError != nil {
				catchErrorFunc := *catchError
				catchErrorFunc(o, key, err)
			}
			// Even in case of an error, return the result from the cache if available.
			return o.Get(key)
		} else {
			// If the function was successful, set the value in the cache and return true.
			o.Set(key, value, d)
			return value, true
		}
	}
	// Return the value from the cache.
	return value, ok
}

// NewOnceCache creates a new instance of OnceCache with the specified singleflight.Group and ICache.
func NewOnceCache(group *singleflight.Group, cacheStore ICache) IOnceCache {
	return &OnceCache{
		group:  group,
		ICache: cacheStore,
	}
}
