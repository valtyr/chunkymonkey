package gamerules

import (
	"chunkymonkey/types"
)

const (
	MaxStackDefault = types.ItemCount(64)
)

type ToolTypeId byte

type ItemType struct {
	Id       types.ItemTypeId
	Name     string
	MaxStack types.ItemCount
	ToolType ToolTypeId
	ToolUses types.ItemData
}

type ItemTypeMap map[types.ItemTypeId]*ItemType
