package gamerules

import (
	"expvar"
	"io"
	"os"

	"chunkymonkey/nbtutil"
	"chunkymonkey/physics"
	"chunkymonkey/proto"
	. "chunkymonkey/types"
	"nbt"
)

var (
	expVarMobSpawnCount *expvar.Int
)

func init() {
	expVarMobSpawnCount = expvar.NewInt("mob-spawn-count")
}

// When using an object of type Mob or a sub-type, the caller must set an
// EntityId, most likely obtained from the EntityManager.
type Mob struct {
	EntityId
	physics.PointObject
	mobType EntityMobType
	look    LookDegrees
	// TODO(nictuku): Move to a more structured form.
	metadata map[byte]byte
	// TODO: Change to an AABB object when we have that.
}

func (mob *Mob) Init(id EntityMobType) {
	mob.mobType = id
	mob.metadata = map[byte]byte{
		0:  byte(0),
		16: byte(0),
	}

	expVarMobSpawnCount.Add(1)
}

func (mob *Mob) ReadNbt(tag nbt.ITag) (err os.Error) {
	if err = mob.PointObject.ReadNbt(tag); err != nil {
		return
	}

	if mob.look, err = nbtutil.ReadLookDegrees(tag, "Rotation"); err != nil {
		return
	}

	// TODO
	_ = tag.Lookup("FallDistance").(*nbt.Float).Value
	_ = tag.Lookup("Air").(*nbt.Short).Value
	_ = tag.Lookup("Fire").(*nbt.Short).Value

	return nil
}

func (mob *Mob) SetLook(look LookDegrees) {
	mob.look = look
}

func (mob *Mob) SetBurning(burn bool) {
	if burn {
		mob.metadata[0] |= 0x01
	} else {
		mob.metadata[0] ^= 0x01
	}
}

func (mob *Mob) Tick(blockQuerier physics.IBlockQuerier) (leftBlock bool) {
	// TODO: Spontaneous mob movement.
	return mob.PointObject.Tick(blockQuerier)
}

func (mob *Mob) FormatMetadata() []proto.EntityMetadata {
	x := make([]proto.EntityMetadata, len(mob.metadata))
	i := 0
	for k, v := range mob.metadata {
		x[i] = proto.EntityMetadata{0, k, v}
		i++
	}
	return x
}

func (mob *Mob) SendUpdate(writer io.Writer) (err os.Error) {
	if err = proto.WriteEntity(writer, mob.EntityId); err != nil {
		return
	}

	err = mob.PointObject.SendUpdate(writer, mob.EntityId, mob.look.ToLookBytes())

	return
}

func (mob *Mob) SendSpawn(writer io.Writer) (err os.Error) {
	err = proto.WriteEntitySpawn(
		writer,
		mob.EntityId,
		mob.mobType,
		&mob.PointObject.LastSentPosition,
		mob.look.ToLookBytes(),
		mob.FormatMetadata())
	if err != nil {
		return
	}
	err = proto.WriteEntityVelocity(
		writer,
		mob.EntityId,
		&mob.PointObject.LastSentVelocity)
	if err != nil {
		return
	}

	return
}

// Evil mobs.

type Creeper struct {
	Mob
}

var (
	creeperNormal   = byte(0)
	creeperBlueAura = byte(1)
)

func NewCreeper() (c *Creeper) {
	c = new(Creeper)
	c.Mob.Init(CreeperType.Id)
	c.Mob.metadata[17] = creeperNormal
	c.Mob.metadata[16] = byte(255)
	return c
}

func (c *Creeper) SetNormalStatus() {
	c.Mob.metadata[17] = creeperNormal
}

func (c *Creeper) CreeperSetBlueAura() {
	c.Mob.metadata[17] = creeperBlueAura
}

type Skeleton struct {
	Mob
}

func NewSkeleton() (s *Skeleton) {
	s = new(Skeleton)
	s.Mob.Init(SkeletonType.Id)
	return
}

type Spider struct {
	Mob
}

func NewSpider() (s *Spider) {
	s = new(Spider)
	s.Mob.Init(SpiderType.Id)
	return
}

type Zombie struct {
	Mob
}

func NewZombie() (s *Zombie) {
	s = new(Zombie)
	s.Mob.Init(ZombieType.Id)
	return
}

// Passive mobs.

type Pig struct {
	Mob
}

func NewPig() (p *Pig) {
	p = new(Pig)
	p.Mob.Init(PigType.Id)
	return
}

type Sheep struct {
	Mob
}

func NewSheep() (s *Sheep) {
	s = new(Sheep)
	s.Mob.Init(SheepType.Id)
	return
}

type Cow struct {
	Mob
}

func NewCow() (c *Cow) {
	c = new(Cow)
	c.Mob.Init(CowType.Id)
	return
}

type Hen struct {
	Mob
}

func NewHen() (h *Hen) {
	h = new(Hen)
	h.Mob.Init(HenType.Id)
	return
}

type Squid struct {
	Mob
}

func NewSquid() (s *Squid) {
	s = new(Squid)
	s.Mob.Init(SquidType.Id)
	return
}

type Wolf struct {
	Mob
}

func NewWolf() (w *Wolf) {
	w = new(Wolf)
	w.Mob.Init(WolfType.Id)
	// TODO(nictuku): String with an optional owner's username.
	w.Mob.metadata[17] = 0
	w.Mob.metadata[16] = 0
	w.Mob.metadata[18] = 0
	return w
}
