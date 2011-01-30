package physics

import (
    "io"
    "math"
    "os"

    "chunkymonkey/proto"
    .   "chunkymonkey/types"
)

const (
    // Guestimated gravity value. Unknown how accurate this is.
    gravityBlocksPerSecond2 = 3.0

    gravityBlocksPerTick2 = gravityBlocksPerSecond2 / TicksPerSecond

    // Air resistance, as a denominator of a velocity component
    airResistance = 5

    // Min velocity component before we clamp to zero
    minVel = 0.01

    objBlockDistance = 4.25 / PixelsPerBlock
)

type blockAxisMove byte

const (
    blockAxisMoveX = blockAxisMove(iota)
    blockAxisMoveY = blockAxisMove(iota)
    blockAxisMoveZ = blockAxisMove(iota)
)

type PointObject struct {
    // Used in knowing what to send as client updates
    LastSentPosition AbsIntXYZ
    LastSentVelocity Velocity

    // Used in physical modelling
    Position AbsXYZ
    Velocity AbsVelocity
    onGround bool
}

type BlockSolidFn func(*BlockXYZ) bool

func (obj *PointObject) Init(position *AbsXYZ, velocity *AbsVelocity) {
    obj.LastSentPosition = *position.ToAbsIntXYZ()
    obj.LastSentVelocity = *velocity.ToVelocity()
    obj.Position = *position
    obj.Velocity = *velocity
    obj.onGround = false
}

// Generates any packets needed to update clients as to the position and
// velocity of the object.
// It assumes that the clients have either been sent packets via this method
// before, or that the previous position/velocity sent was generated from the
// LastSentPosition and LastSentVelocity attributes.
func (obj *PointObject) SendUpdate(writer io.Writer, entityID EntityID, look *LookBytes) (err os.Error) {
    curPosition := obj.Position.ToAbsIntXYZ()

    dx := curPosition.X - obj.LastSentPosition.X
    dy := curPosition.Y - obj.LastSentPosition.Y
    dz := curPosition.Z - obj.LastSentPosition.Z

    if dx != 0 || dy != 0 || dz != 0 {
        if dx >= -128 && dx <= 127 && dy >= -128 && dy <= 127 && dz >= -128 && dz <= 127 {
            err = proto.WriteEntityRelMove(
                writer, entityID,
                &RelMove{
                    RelMoveCoord(dx),
                    RelMoveCoord(dy),
                    RelMoveCoord(dz),
                },
            )
        } else {
            err = proto.WriteEntityTeleport(
                writer, entityID,
                curPosition, look)
        }
        if err != nil {
            return
        }
        obj.LastSentPosition = *curPosition
    }

    curVelocity := obj.Velocity.ToVelocity()
    if curVelocity.X != obj.LastSentVelocity.X || curVelocity.Y != obj.LastSentVelocity.Y || curVelocity.Z != obj.LastSentVelocity.Z {
        if err = proto.WriteEntityVelocity(writer, entityID, curVelocity); err != nil {
            return
        }
        obj.LastSentVelocity = *curVelocity
    }

    return
}

func (obj *PointObject) Tick(blockSolid BlockSolidFn) (destroyed bool) {
    // TODO this algorithm can probably be sped up a bit, but initially trying
    // to keep things simple and more or less correct

    p := &obj.Position
    v := &obj.Velocity

    // FIXME note that if the block under the item should become non-solid,
    // then we need to turn off onGround to re-enable physics
    // TODO if the object has stopped moving (i.e is at rest on top of a solid
    // block and not inside a flowing block), take the object out of a
    // "physically active" list. Note that the object will have to be re-added
    // if any blocks it is adjacent to change in solidity or "flow"
    stopped := obj.updateVelocity()

    if stopped {
        // The object isn't moving, we're done
        return
    }

    // Enforce max absolute velocity per dimension
    v.X.Constrain()
    v.Y.Constrain()
    v.Z.Constrain()

    // t0 = time at start of tick,
    // t1 = time at end of tick,
    // t = current time in tick (t0 <= t <= t1)
    var t0, t1, t TickTime

    // `Dt` and `dt` means delta-time, that is, time relative to `t`
    var nextBlockXDt, nextBlockYDt, nextBlockZDt TickTime

    var dt TickTime

    t0 = 0.0
    t1 = 1.0
    dt = 0

    var move blockAxisMove

    // Project the velocity in block space to see if we hit anything solid, and
    // stop the object's velocity component if so

    for t = t0; t < t1; t += dt {
        // How long after t0 is it that the object hits a block boundary on
        // each axis?
        nextBlockXDt = calcNextBlockDt(p.X, v.X)
        nextBlockYDt = calcNextBlockDt(p.Y, v.Y)
        nextBlockZDt = calcNextBlockDt(p.Z, v.Z)

        // In the axis of which block are we moving? In X, Y or Z axis?
        move, dt = getBlockAxisMove(nextBlockXDt, nextBlockYDt, nextBlockZDt)

        // Don't calculate beyond 1 tick of time
        if t+dt >= t1 {
            // It will be after the end of this tick when the object crosses a
            // block boundary
            dt = t1 - t
            p.ApplyVelocity(dt, v)

            // We're all done
            break
        } else {

            // Examine the block being entered
            blockLoc := obj.nextBlockToEnter(move)

            // Is it solid?
            if blockSolid(blockLoc) {
                // Collision - cancel axis movement
                switch move {
                case blockAxisMoveX:
                    applyCollision(&p.X, &v.X)
                case blockAxisMoveY:
                    applyCollision(&p.Y, &v.Y)
                    obj.onGround = true
                case blockAxisMoveZ:
                    applyCollision(&p.Z, &v.Z)
                }

                // Move the object up to the block boundary
                p.ApplyVelocity(dt, v)
            } else {
                // No collision, continue as normal
                // HACK: We add 1e-4 to dt to "break past" the block boundary,
                // otherwise we end up at rest on it in an infinite loop. dt
                // would otherwise be *approximately* sufficient to reach the
                // block boundary.
                p.ApplyVelocity(dt+1e-4, v)
            }
        }
    }

    // Destroy object if fallen out the bottom of the world
    if p.Y < 0 {
        destroyed = true
    }
    return
}

func (obj *PointObject) updateVelocity() (stopped bool) {
    v := &obj.Velocity

    if !obj.onGround {
        v.Y -= gravityBlocksPerTick2
    }

    stopped = true

    if v.X > -minVel && v.X < minVel {
        v.X = 0
    } else {
        v.X -= v.X / airResistance
        stopped = false
    }
    if v.Y > -minVel && v.Y < minVel {
        v.Y = 0
    } else {
        v.Y -= v.Y / airResistance
        stopped = false
    }
    if v.Z > -minVel && v.Z < minVel {
        v.Z = 0
    } else {
        v.Z -= v.Z / airResistance
        stopped = false
    }

    return
}

func (obj *PointObject) nextBlockToEnter(move blockAxisMove) *BlockXYZ {
    p := &obj.Position
    v := &obj.Velocity

    block := p.ToBlockXYZ()

    switch move {
    case blockAxisMoveX:
        if v.X > 0 {
            block.X += 1
        } else {
            block.X -= 1
        }
    case blockAxisMoveY:
        if v.Y > 0 {
            block.Y += 1
        } else {
            block.Y -= 1
        }
    case blockAxisMoveZ:
        if v.Z > 0 {
            block.Z += 1
        } else {
            block.Z -= 1
        }
    }

    return block
}

// In one dimension, calculates time taken for movement from position `p` with
// velocity `v` until intersection with a block boundary. Note that if v is
// small enough or zero then math.MaxFloat64 is returned.
func calcNextBlockDt(p AbsCoord, v AbsVelocityCoord) TickTime {
    if v > -1e-20 && v < 1e-20 {
        return math.MaxFloat64
    }

    if p < 0 {
        p = -p
        v = -v
    }

    var p_prime AbsCoord
    if v > 0 {
        p_prime = AbsCoord(math.Floor(float64(p + 1.0)))
    } else {
        p_prime = AbsCoord(math.Floor(float64(p)))
    }

    return TickTime(float64(p_prime-p) / float64(v))
}

// Given 3 time deltas, it returns the axis that the smallest was on, and the
// value of the smallest time delta. This is used to know on which axis and how
// long until the next block transition is. Only +ve numbers should be passed
// in for it to be sensible.
func getBlockAxisMove(xDt, yDt, zDt TickTime) (move blockAxisMove, dt TickTime) {
    if xDt <= yDt {
        if xDt <= zDt {
            return blockAxisMoveX, xDt
        } else {
            return blockAxisMoveZ, zDt
        }
    } else {
        if yDt <= zDt {
            return blockAxisMoveY, yDt
        } else {
            return blockAxisMoveZ, zDt
        }
    }
    return
}

func applyCollision(p *AbsCoord, v *AbsVelocityCoord) {
    if *v > 0 {
        *p = AbsCoord(math.Ceil(float64(*p)) - objBlockDistance)
    } else {
        *p = AbsCoord(math.Floor(float64(*p)) + objBlockDistance)
    }
    *v = 0
}
