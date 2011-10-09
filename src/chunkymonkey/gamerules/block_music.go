package gamerules

import (
	"os"

	"chunkymonkey/types"
	"nbt"
)

func makeMusicAspect() (aspect IBlockAspect) {
	return &MusicAspect{}
}

type musicTileEntity struct {
	tileEntity
	note types.NotePitch
}

func NewMusicTileEntity() ITileEntity {
	return &musicTileEntity{}
}

func (music *musicTileEntity) UnmarshalNbt(tag *nbt.Compound) (err os.Error) {
	if err = music.tileEntity.UnmarshalNbt(tag); err != nil {
		return
	}

	if noteTag, ok := tag.Lookup("note").(*nbt.Byte); !ok {
		return os.NewError("missing or incorrect type for Music note")
	} else {
		music.note = types.NotePitch(noteTag.Value)
	}

	return nil
}

func (music *musicTileEntity) MarshalNbt(tag *nbt.Compound) (err os.Error) {
	if err = music.tileEntity.MarshalNbt(tag); err != nil {
		return
	}

	tag.Set("id", &nbt.String{"Music"})
	tag.Set("note", &nbt.Byte{int8(music.note)})

	return nil
}

type MusicAspect struct {
	StandardAspect
}

func (aspect MusicAspect) Name() string {
	return "Music"
}

// TODO behaviours for music.
