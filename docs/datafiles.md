blocks.json
===========

The structure of blocks.json is at the top-level a mapping from block type ID
to the definition. The definition itself consists of two main parts: the basic
block attributes and the aspect.

  {
    "0": {
      "BlockAttrs": {
        ... basic block attributes ...
      },
      "Aspect": "aspect name",
      "AspectArgs": {
        ... aspect parameters ...
      }
    }
  }


`BlockAttrs`
------------

This defines very basic physical properties of a block type. The fields are:

*  `Name` (string) a very short name for the block type. This isn't typically
   used for much (yet), but can be useful in human-readability of both
   blocks.json, and in any output that the server might output relating to a block
   type.
*  `Opacity` (integer) the amount of light attenuation upon light potentially
   passing through the block. 15 is completely opaque (like stone), 0 is
   completely transparent (like glass or air).
*  `Destructable` (bool) `true` means that the block is destroyable via normal
   means. This includes players digging and (potentially) explosions. Typically
   bedrock is not destructable, but anything else is.
*  `Solid` (bool) `true` means that players, mobs, etc. cannot pass through the
   block. This is used in physical modelling of entities in the world. Air and
   water are non-solid, stone and earth are solid. There are many blocks that
   do not impede movement as well, such as torches, light snow, etc.
*  `Replaceable` (bool) `true` means that blocks can be placed by players to
   replace blocks of this type. Air, water and lava are some of the (few)
   examples of blocks that are replaceable.
*  `Attachable` (bool) `true` means that players can place blocks *against*
   this block type. Stone is attachable, chests, water, torches etc. are not.

`Aspect` and `AspectArgs`
-------------------------

This defines more advanced behaviour of the block. `Aspect` refers to a block
behaviour type in code (these must be present and registered in
`src/chunkymonkey/gamerules/block_loader.go` in order for blocks.json to load
without error. `AspectArgs` supplies aspect-specific parameters, fine tuning
the behaviour. The parameters for each aspect type is varied, and as a general
rule, looking at the contents of `src/chunkymonkey/gamerules/block_*.go` will
provide some useful information.
