package cache

import "codexie.com/w-book-interact/internal/dao/db"

type TopLikeCache interface {
	UpdateFromRedis(resourceType string) error
	GetTopResources(resourceType string) ([]db.Interaction, error)
}
