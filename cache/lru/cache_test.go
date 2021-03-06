package lru

import (
	"testing"
)

type randomStruct struct {
	ID Snowflake
}

func TestCacheList(t *testing.T) {
	t.Run("size limit", func(t *testing.T) {
		limit := uint(10)
		list := NewCacheList(limit)
		if list.size() != 0 {
			t.Error("size if not 0")
		}
		for i := 1; i < 256; i++ {
			usr := &randomStruct{}
			usr.ID = Snowflake(i)

			item := NewCacheItem(usr)
			list.Set(usr.ID, item)
		}

		if list.size() > limit {
			t.Errorf("list has a greater size than expected limit. Got %d, wants %d", list.size(), limit)
		}
	})
	t.Run("replaces only LRU", func(t *testing.T) {
		ids := []Snowflake{4, 7, 12, 46, 74, 89}
		list := NewCacheList(uint(len(ids)))
		for i := 1; i < 256; i++ {
			usr := &randomStruct{}
			usr.ID = Snowflake(i)
			item := NewCacheItem(usr)

			for _, id := range ids {
				if usr.ID == id {
					item.lastUsed += 4
				}
			}

			list.Set(usr.ID, item)
		}

		for _, item := range list.items {
			if item.lastUsed < 4 {
				if item.item.(*randomStruct).ID != Snowflake(255) {
					t.Errorf("expected lru counter to be higher. Got %d, wants above %d", item.lastUsed, 4)
				}
			}
		}
	})
}
