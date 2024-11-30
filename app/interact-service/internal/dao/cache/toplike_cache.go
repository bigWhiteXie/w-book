package cache

type TopLikeCache interface {
	UpdateResourceCache(resourceType string, resources []int64) error
	GetTopResources(resourceType string) ([]int64, error)
	TryLock(resourceType string) bool
	Unlock(resourceType string)
}
