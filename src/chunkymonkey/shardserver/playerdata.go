package shardserver

import (
	"io"
	"os"

	"chunkymonkey/proto"
	. "chunkymonkey/types"
)

// playerData represents a Chunk's knowledge about a player. Only one Chunk has
// this data at a time. This data is occasionally updated from the frontend
// server.
type playerData struct {
	entityId   EntityId
	name       string
	position   AbsXyz
	look       LookBytes
	heldItemId ItemTypeId
	// TODO Armor data.
}

func (player *playerData) SendSpawn(writer io.Writer) (err os.Error) {
	return proto.WriteNamedEntitySpawn(
		writer,
		player.entityId, player.name,
		player.position.ToAbsIntXyz(),
		&player.look,
		player.heldItemId,
	)
}
