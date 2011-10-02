package proto

// This file is concerned with reading/writing packets, dispatching between
// packet IDs and their type.

import (
	"io"
	"os"
	"reflect"
)

var (
	// Maps from a packet ID to information about that packet type. This is used
	// when receiving packets.
	pktIdInfo [256]pktInfo

	// Maps from an interface type ptr to the ID for that packet. This is used
	// when sending packets.
	pktTypeId map[reflect.Type]byte = make(map[reflect.Type]byte, 256)

	ErrorUnknownPacketType = os.NewError("unknown packet type")
	ErrorUnexpectedPacket  = os.NewError("unexpected packet id")
)

type pktInfo struct {
	validPacket    bool
	clientToServer bool
	serverToClient bool
	pktType        reflect.Type
}

// TODO Consider moving all the reading/writing into serialize.go and just make
// this functions to work with packet IDs and types.

func WritePacket(writer io.Writer, ps *PacketSerializer, pkt interface{}) (err os.Error) {
	var idArray [1]byte
	var ok bool

	pktType := reflect.TypeOf(pkt)

	idArray[0], ok = pktTypeId[pktType]
	if !ok {
		return ErrorUnknownPacketType
	}

	if _, err = writer.Write(idArray[0:1]); err != nil {
		return
	}

	return ps.WritePacket(writer, pkt)
}

func ReadPacket(reader io.Reader, ps *PacketSerializer, isServer bool) (pkt interface{}, err os.Error) {
	var idArray [1]byte

	if _, err = io.ReadFull(reader, idArray[0:1]); err != nil {
		return
	}

	id := idArray[0]
	pktInfo := pktIdInfo[id]
	if !pktInfo.validPacket {
		return nil, ErrorUnknownPacketType
	}

	if isServer && !pktInfo.clientToServer {
		return nil, ErrorUnexpectedPacket
	} else if !pktInfo.serverToClient {
		return nil, ErrorUnexpectedPacket
	}

	pkt = reflect.New(pktInfo.pktType).Interface()
	if err = ps.ReadPacket(reader, pkt); err != nil {
		return nil, err
	}

	return
}

// Packet defintions.
var pktDefns = []struct {
	id             byte
	clientToServer bool
	serverToClient bool
	pkt            interface{}
}{
	// id, c->s, s->c, packet
	{0x00, true, true, &PacketKeepAlive{}},
	{0x01, true, true, &PacketLogin{}},
	{0x02, true, true, &PacketHandshake{}},
	{0x03, true, true, &PacketChatMessage{}},
	{0x04, false, true, &PacketTimeUpdate{}},
	{0x05, false, true, &PacketEntityEquipment{}},
	{0x06, false, true, &PacketSpawnPosition{}},
	{0x07, true, true, &PacketUseEntity{}},
	{0x08, false, true, &PacketUpdateHealth{}},
	{0x09, true, true, &PacketRespawn{}},
	{0x0a, true, false, &PacketPlayer{}},
	{0x0b, true, true, &PacketPlayerPosition{}},
	{0x0c, true, true, &PacketPlayerLook{}},
	{0x0d, true, true, &PacketPlayerPositionLook{}},
	{0x0e, true, true, &PacketPlayerBlockHit{}},
	{0x0f, true, true, &PacketPlayerBlockInteract{}},
	{0x10, true, false, &PacketPlayerHoldingChange{}},
	{0x11, false, true, &PacketPlayerUseBed{}},
	{0x12, true, true, &PacketEntityAnimation{}},
	{0x13, true, true, &PacketEntityAction{}},
	{0x14, false, true, &PacketNamedEntitySpawn{}},
	{0x15, false, true, &PacketItemSpawn{}},
	{0x16, false, true, &PacketItemCollect{}},
	{0x17, false, true, &PacketObjectSpawn{}},
	{0x18, false, true, &PacketMobSpawn{}},
	{0x19, false, true, &PacketPaintingSpawn{}},
	{0x1a, false, true, &PacketExperienceOrb{}},
	{0x1c, false, true, &PacketEntityVelocity{}},
	{0x1d, false, true, &PacketEntityDestroy{}},
	{0x1e, false, true, &PacketEntity{}},
	{0x1f, false, true, &PacketEntityRelMove{}},
	{0x20, false, true, &PacketEntityLook{}},
	{0x21, false, true, &PacketEntityLookAndRelMove{}},
	{0x22, false, true, &PacketEntityTeleport{}},
	{0x26, false, true, &PacketEntityStatus{}},
	{0x27, false, true, &PacketEntityAttach{}},
	{0x28, false, true, &PacketEntityMetadata{}},
	{0x29, false, true, &PacketEntityEffect{}},
	{0x2a, false, true, &PacketEntityRemoveEffect{}},
	{0x2b, false, true, &PacketPlayerExperience{}},
	{0x32, false, true, &PacketPreChunk{}},
	{0x33, false, true, &PacketMapChunk{}},
	{0x34, false, true, &PacketMultiBlockChange{}},
	{0x35, false, true, &PacketBlockChange{}},
	{0x36, false, true, &PacketSoundEffect{}},
	{0x3c, false, true, &PacketExplosion{}},
	{0x3d, false, true, &PacketSoundEffect{}},
	{0x46, false, true, &PacketState{}},
	{0x47, false, true, &PacketThunderbolt{}},
	{0x64, true, true, &PacketWindowOpen{}},
	{0x65, true, true, &PacketWindowClose{}},
	{0x66, true, false, &PacketWindowClick{}},
	{0x67, false, true, &PacketWindowSetSlot{}},
	{0x68, false, true, &PacketWindowItems{}},
	{0x69, false, true, &PacketWindowProgressBar{}},
	{0x6a, true, true, &PacketWindowTransaction{}},
	{0x6b, false, true, &PacketQuickbarSlotUpdate{}},
	{0x82, true, true, &PacketSignUpdate{}},
	{0x83, false, true, &PacketItemData{}},
	{0xc8, false, true, &PacketIncrementStatistic{}},
	{0xc9, false, true, &PacketPlayerListItem{}},
	{0xfe, true, false, &PacketServerListPing{}},
	{0xff, true, true, &PacketDisconnect{}},
}

func init() {
	for i := range pktDefns {
		defn := &pktDefns[i]
		pktType := reflect.TypeOf(defn.pkt)

		// Map from ID to info.
		pktIdInfo[defn.id] = pktInfo{
			validPacket:    true,
			clientToServer: defn.clientToServer,
			serverToClient: defn.serverToClient,
			pktType:        pktType,
		}

		// Map from type to ID.
		pktTypeId[pktType] = defn.id
	}
}
