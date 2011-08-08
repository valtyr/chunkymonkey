package chunkstore

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path"

	. "chunkymonkey/types"
	"nbt"
)

// Handle on a chunk file - used to read chunk data from the file.
type regionFile struct {
	offsets   regionFileHeader
	endSector uint32 // The index of the sector just beyond the last used sector.
	file      *os.File
}

func newRegionFile(filePath string) (rf *regionFile, err os.Error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|O_CREATE, 0666)
	if err != nil {
		if sysErr, ok := err.(*os.SyscallError); ok && sysErr.Errno == os.ENOENT {
			err = NoSuchChunkError(false)
		}
		return
	}

	fi, err := file.Stat()
	if err != nil {
		return
	}

	rf = &regionFile{
		file: file,
	}

	if fi.Size == 0 {
		// Newly created region file. Create new header index if so.
		if err = binary.Write(file, binary.BigEndian, &rf.offsets); err != nil {
			return
		}

		rf.endSector = 1
	} else {
		// Existing region file, read header index.
		err = binary.Read(file, binary.BigEndian, &rf.offsets)
		if err != nil {
			rf = nil
			return
		}

		// Find the index of the sector at the end of the file.
		for i := range rf.offsets {
			sectorCount, sectorIndex := rf.offsets[i].Get()
			lastUsedSector := sectorIndex + sectorCount
			if lastUsedSector >= rf.endSector {
				rf.endSector = lastUsedSector + 1
			}
		}
	}

	return
}

func (rf *regionFile) Close() {
	rf.file.Close()
}

func (rf *regionFile) ReadChunkData(chunkLoc ChunkXz) (r *nbtChunkReader, err os.Error) {
	offset := rf.offsets.Offset(chunkLoc)

	if !offset.IsPresent() {
		// Chunk doesn't exist in file
		err = NoSuchChunkError(false)
		return
	}

	sectorCount, sectorIndex := offset.Get()

	if sectorIndex == 0 || sectorCount == 0 {
		err = os.NewError("Header gave bad chunk offset.")
		return
	}

	rf.file.Seek(int64(sectorIndex)*regionFileSectorSize, 0)

	maxChunkDataSize := (sectorCount * regionFileSectorSize) - chunkDataHeaderSize

	var header chunkDataHeader
	binary.Read(rf.file, binary.BigEndian, &header)
	if header.DataSize > maxChunkDataSize {
		err = os.NewError("Chunk is too big for the sectors it is within.")
		return
	}

	dataReader, err := header.DataReader(rf.file)
	if err != nil {
		return
	}
	defer dataReader.Close()

	r, err = newNbtChunkReader(dataReader)

	return
}

func (rf *regionFile) WriteChunkData(w *nbtChunkWriter) (err os.Error) {
	chunkData, err := serializeChunkData(w)
	if err != nil {
		return
	}

	header := chunkDataHeader{
		DataSize: uint32(len(chunkData)),
		Version:  chunkCompressionZlib,
	}

	requiredSize := chunkDataHeaderSize + len(chunkData)

	offset := rf.offsets.Offset(w.ChunkLoc())

	if !offset.IsPresent() {
		// Chunk doesn't yet exist in the region file. Write it at the end of the
		// file.
		// TODO
	}

	// TODO Chunk already exists in the region file.

	// TODO Will the chunk fit in the sectors it was in previously?

	return
}

// serializeChunkData produces the compressed chunk NBT data.
func serializeChunkData(w *nbtChunkWriter) (chunkData []byte, err os.Error) {
	// Serialize and compress the NBT data.
	buffer := bytes.NewBuffer(make([]byte, 0, chunkDataGuessSize))
	if zlibWriter, err := zlib.NewWriter(buffer); err != nil {
		return nil, err
	} else {
		if err = nbt.Write(zlibWriter, w.RootTag()); err != nil {
			zlibWriter.Close()
			return nil, err
		}
		if err = zlibWriter.Close(); err != nil {
			return nil, err
		}
	}
	chunkData = buffer.Bytes()
	return chunkData, nil
}

// A chunk file header entry.
type chunkOffset uint32

// Returns true if the offset value states that the chunk is present in the
// file.
func (o chunkOffset) IsPresent() bool {
	return o != 0
}

func (o chunkOffset) Get() (sectorCount, sectorIndex uint32) {
	sectorCount = uint32(o & 0xff)
	sectorIndex = uint32(o >> 8)
	return
}

// Represents a chunk file header containing chunk data offsets.
type regionFileHeader [regionFileEdge * regionFileEdge]chunkOffset

// Returns the chunk offset data for the given chunk. It assumes that chunkLoc
// is within the chunk file - discarding upper bits of the X and Z coords.
func (h regionFileHeader) Offset(chunkLoc ChunkXz) chunkOffset {
	x := chunkLoc.X & (regionFileEdge - 1)
	z := chunkLoc.Z & (regionFileEdge - 1)
	return h[x+(z<<regionFileEdgeShift)]
}

// Represents the header of a single chunk of data within a chunkfile.
type chunkDataHeader struct {
	DataSize uint32
	Version  byte
}

// Returns an io.Reader to correctly decompress data from the chunk data.
// The reader passed in must be just after the chunkDataHeader in the source
// data stream. The caller is responsible for closing the returned ReadCloser.
func (cdh *chunkDataHeader) DataReader(raw io.Reader) (output io.ReadCloser, err os.Error) {
	limitReader := io.LimitReader(raw, int64(cdh.DataSize))
	switch cdh.Version {
	case chunkCompressionGzip:
		output, err = gzip.NewReader(limitReader)
	case chunkCompressionZlib:
		output, err = zlib.NewReader(limitReader)
	default:
		err = os.NewError("Chunk data header contained unknown version number.")
	}
	return
}

type regionCoord int32

type regionLoc struct {
	X, Z regionCoord
}

func regionLocForChunkXz(chunkLoc ChunkXz) regionLoc {
	return regionLoc{
		regionCoord(chunkLoc.X >> regionFileEdgeShift),
		regionCoord(chunkLoc.Z >> regionFileEdgeShift),
	}
}

func (loc *regionLoc) regionKey() uint64 {
	return uint64(loc.X)<<32 | uint64(uint32(loc.Z))
}

func (loc *regionLoc) regionFilePath(regionPath string) string {
	return path.Join(
		regionPath,
		fmt.Sprintf("r.%d.%d.mcr", loc.X, loc.Z),
	)
}
