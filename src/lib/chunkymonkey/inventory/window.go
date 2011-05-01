package inventory

import (
	"bytes"
	"io"
	"os"

	"chunkymonkey/proto"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

// IWindow is the interface on to types that represent a view on to multiple
// inventories.
// TODO remove this type if not used in the end.
type IWindow interface {
	Click(slotId SlotId, cursor *slot.Slot, rightClick bool, shiftClick bool) (accepted bool)
}

// IWindowViewer is the required interface of types that wish to receive packet
// updates from changes to inventories viewed inside a window. Typically
// *player.Player implements this.
type IWindowViewer interface {
	TransmitPacket(packet []byte)
}

// inventoryView provides a single mapping between a window view onto an
// inventory at a particular slot range inside the window.
type inventoryView struct {
	window    *Window
	inventory *Inventory
	startSlot SlotId
}

func (iv *inventoryView) Init(window *Window, inventory *Inventory, startSlot SlotId) {
	iv.window = window
	iv.inventory = inventory
	iv.startSlot = startSlot
	iv.inventory.AddSubscriber(iv)
}

func (iv *inventoryView) Finalize() {
	iv.inventory.RemoveSubscriber(iv)
}

// Implementing IInventorySubscriber - relays inventory changes to the viewer
// of the window.
func (iv *inventoryView) SlotUpdate(slot *slot.Slot, slotId SlotId) {
	buf := &bytes.Buffer{}
	slot.SendUpdate(buf, iv.window.windowId, iv.startSlot+slotId)
	iv.window.viewer.TransmitPacket(buf.Bytes())
}

// Window represents the common base behaviour of an inventory window. It acts
// as a view onto multiple Inventories.
type Window struct {
	windowId  WindowId
	invTypeId InvTypeId
	viewer    IWindowViewer
	views     []inventoryView
	title     string
	numSlots  int
}

func (w *Window) Init(windowId WindowId, invTypeId InvTypeId, viewer IWindowViewer, title string, inventories ...*Inventory) {
	w.windowId = windowId
	w.invTypeId = invTypeId
	w.viewer = viewer
	w.title = title

	w.views = make([]inventoryView, len(inventories))
	startSlot := 0
	for index, inv := range inventories {
		w.views[index].Init(w, inv, SlotId(startSlot))
		startSlot += len(inv.slots)
	}
	w.numSlots = startSlot

	return
}

// Finalize cleans up, subscriber information so that the window can be
// properly garbage collected. This should be called when the window is thrown
// away.
func (w *Window) Finalize() {
	for index := range w.views {
		w.views[index].Finalize()
	}
}

// WriteWindowOpen writes a packet describing the window to the writer.
func (w *Window) WriteWindowOpen(writer io.Writer) (err os.Error) {
	err = proto.WriteWindowOpen(writer, w.windowId, w.invTypeId, w.title, byte(w.numSlots))
	return
}

// WriteWindowItems writes a packet describing the window contents to the
// writer. It assumes that any required locks on the inventories are held.
func (w *Window) WriteWindowItems(writer io.Writer) (err os.Error) {
	items := make([]proto.IWindowSlot, w.numSlots)

	slotIndex := 0
	for i := range w.views {
		inv := w.views[i].inventory
		// TODO Acquiring multiple simultaneous locks is somewhat dodgy - as it
		// can lead to deadlock. Find a better solution.
		inv.lock.Lock()
		defer inv.lock.Unlock()
		for i := range inv.slots {
			items[slotIndex] = &inv.slots[i]
			slotIndex++
		}
	}

	err = proto.WriteWindowItems(writer, w.windowId, items)
	return
}
