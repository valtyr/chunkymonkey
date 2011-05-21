package entity

import (
	"io"
	"os"
	"sync"

	"chunkymonkey/physics"
	. "chunkymonkey/types"
)

// TODO Should EntityManager hold interfaces to Entity-like objects instead?

type Entity struct {
	EntityId EntityId
}

// TODO EntityManager should be a service in its own right, able to hand out
// blocks of IDs and running its own goroutine (potentially shardable by
// entityId if necessary). Right now taking the easy option of using a simple
// lock.
type EntityManager struct {
	nextEntityId EntityId
	entities     map[EntityId]*Entity
	lock         sync.Mutex
}

func (mgr *EntityManager) Init() {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	mgr.nextEntityId = 0
	mgr.entities = make(map[EntityId]*Entity)
}

func (mgr *EntityManager) createEntityId() EntityId {
	// Search for next free ID
	entityId := mgr.nextEntityId
	_, exists := mgr.entities[entityId]
	for exists {
		entityId++
		if entityId == mgr.nextEntityId {
			// TODO Better handling of this? It shouldn't happen, realistically - but
			// neither should it explode.
			panic("EntityId space exhausted")
		}
		_, exists = mgr.entities[entityId]
	}
	mgr.nextEntityId = entityId + 1

	return entityId
}

// AddEntity adds an entity to the manager, and assigns it a world-unique
// EntityId.
func (mgr *EntityManager) AddEntity(entity *Entity) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	entity.EntityId = mgr.createEntityId()
	mgr.entities[entity.EntityId] = entity
}

// RemoveEntity removes an entity from the manager.
// TODO deprecate
func (mgr *EntityManager) RemoveEntity(entity *Entity) {
	mgr.RemoveEntityById(entity.EntityId)
}

// RemoveEntity removes an entity from the manager.
func (mgr *EntityManager) RemoveEntityById(entityId EntityId) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	mgr.entities[entityId] = nil, false
}

// ISpawn has the ability to spawn an item or mob. It's used for example by the
// chunks.
// TODO move into shard server
type ISpawn interface {
	SendSpawn(io.Writer) os.Error
	GetEntity() *Entity
	GetPosition() *AbsXyz
	SendUpdate(io.Writer) os.Error
	Tick(physics.BlockQueryFn) (leftBlock bool)
}
