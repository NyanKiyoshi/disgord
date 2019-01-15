package disgord

import (
	"github.com/andersfylling/disgord/cache/interfaces"
	"github.com/andersfylling/disgord/httd"
	"github.com/andersfylling/snowflake/v3"
	"github.com/pkg/errors"
)

type gatewayCacher interface {
	handleGatewayEvent(data []byte) error
}
type restCacher interface {
	handleRESTResponse(obj interface{}) error
}

func extractRootID(data []byte) (id snowflake.ID, err error) {
	id, err = extractAttribute([]byte(`"id":"`), 0, data)
	if err != nil {
		err := httd.Unmarshal(data, &struct {
			ID *snowflake.ID `json:"id"`
		}{
			ID: &id,
		})
		if err != nil {
			return id, err
		}
		if id.Empty() {
			return id, errors.New("snowflake is 0")
		}
	}
	return id, nil
}

type /**/ usersCache struct {
	internal interfaces.CacheAlger
}

func (c *usersCache) handleGatewayEvent(data []byte) (err error) {
	var id snowflake.ID
	id, err = extractRootID(data)
	if err != nil {
		return err
	}

	c.internal.Lock()
	defer c.internal.Unlock()
	if item, exists := c.internal.Get(id); exists {
		err = httd.Unmarshal(data, item.Object().(*User))
	} else {
		var user *User
		err = httd.Unmarshal(data, &user)
		if err == nil {
			c.internal.Set(id, c.internal.CreateCacheableItem(user))
		}
	}

	return err
}
func (c *usersCache) handleRESTResponse(obj interface{}) (err error) {
	// don't checking if it's actually a user.
	// panics here will only help us improve the data flow if this method was incorrectly used.
	user := obj.(*User)
	if user == nil {
		return
	}

	c.internal.Lock()
	defer c.internal.Unlock()
	if item, exists := c.internal.Get(user.ID); exists {
		err = user.copyOverToCache(item.Object().(*User))
	} else {
		c.internal.Set(user.ID, c.internal.CreateCacheableItem(user))
	}

	return err
}
func (c *usersCache) Delete(userID snowflake.ID) {
	c.internal.Lock()
	defer c.internal.Unlock()
	c.internal.Delete(userID)
}
func (c *usersCache) Get(userID snowflake.ID) (user *User) {
	c.internal.RLock()
	defer c.internal.RUnlock()
	if item, exists := c.internal.Get(userID); exists {
		user = item.Object().(*User)
	}

	return user
}
func (c *usersCache) Size() uint {
	c.internal.RLock()
	defer c.internal.RUnlock()
	return c.internal.Size()
}
func (c *usersCache) Cap() uint {
	c.internal.RLock()
	defer c.internal.RUnlock()
	return c.internal.Cap()
}
func (c *usersCache) ListIDs() []snowflake.ID {
	c.internal.RLock()
	defer c.internal.RUnlock()
	return c.internal.ListIDs()
}

// Foreach allows you iterate over the users. This is not blocking for the rest of the system
// as it blocks only when it copies or extract data from one user at the time.
// This is faster when you make the cache mutable, but then again that introduces higher
// risk are then involved (race conditions, incorrect cache, etc).
func (c *usersCache) Foreach(cb func(*User)) {
	ids := c.ListIDs()

	for i := range ids {
		user := c.Get(ids[i])
		if user != nil {
			cb(user)
		}
	}
}

var _ gatewayCacher = (*usersCache)(nil)
var _ restCacher = (*usersCache)(nil)
