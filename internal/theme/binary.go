package theme

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// WUIT binary format:
//
//	[4 bytes]  magic: "WUIT"
//	[2 bytes]  version: uint16 (little-endian), currently 1
//	[4 bytes]  theme JSON length: uint32 (little-endian)
//	[4 bytes]  atlas JSON length: uint32 (little-endian)
//	[4 bytes]  atlas image length: uint32 (little-endian)
//	[...]      theme JSON bytes
//	[...]      atlas JSON bytes (region map)
//	[...]      atlas image bytes (PNG-encoded RGBA)

var wuitMagic = [4]byte{'W', 'U', 'I', 'T'}

const (
	wuitVersion    uint16 = 1
	wuitHeaderSize        = 4 + 2 + 4 + 4 + 4 // 18 bytes
)

// EncodeThemeBinary encodes a compiled theme into the WUIT binary format.
// themeJSON is the original theme JSON (colors, components, etc.).
// atlasJSON is the atlas region map. atlasPNG is the packed atlas image
// encoded as PNG. Any of atlasJSON/atlasPNG may be nil (no atlas).
func EncodeThemeBinary(themeJSON, atlasJSON, atlasPNG []byte) ([]byte, error) {
	if themeJSON == nil {
		return nil, errors.New("theme binary: themeJSON is required")
	}

	atlasJSONLen := uint32(len(atlasJSON))
	atlasPNGLen := uint32(len(atlasPNG))

	total := wuitHeaderSize + len(themeJSON) + len(atlasJSON) + len(atlasPNG)
	buf := make([]byte, total)

	// Header
	copy(buf[0:4], wuitMagic[:])
	binary.LittleEndian.PutUint16(buf[4:6], wuitVersion)
	binary.LittleEndian.PutUint32(buf[6:10], uint32(len(themeJSON)))
	binary.LittleEndian.PutUint32(buf[10:14], atlasJSONLen)
	binary.LittleEndian.PutUint32(buf[14:18], atlasPNGLen)

	// Blobs
	off := wuitHeaderSize
	copy(buf[off:], themeJSON)
	off += len(themeJSON)
	copy(buf[off:], atlasJSON)
	off += len(atlasJSON)
	copy(buf[off:], atlasPNG)

	return buf, nil
}

// DecodedThemeBinary holds the three sections extracted from a WUIT file.
type DecodedThemeBinary struct {
	ThemeJSON []byte
	AtlasJSON []byte
	AtlasPNG  []byte
}

// DecodeThemeBinary parses a WUIT binary blob and returns the three sections.
// Returns an error on invalid magic, unsupported version, or truncated data.
func DecodeThemeBinary(data []byte) (*DecodedThemeBinary, error) {
	if len(data) < wuitHeaderSize {
		return nil, fmt.Errorf("theme binary: data too short (%d bytes, need at least %d)", len(data), wuitHeaderSize)
	}

	// Magic
	if data[0] != wuitMagic[0] || data[1] != wuitMagic[1] ||
		data[2] != wuitMagic[2] || data[3] != wuitMagic[3] {
		return nil, fmt.Errorf("theme binary: invalid magic %q (expected \"WUIT\")", string(data[0:4]))
	}

	// Version
	ver := binary.LittleEndian.Uint16(data[4:6])
	if ver != wuitVersion {
		return nil, fmt.Errorf("theme binary: unsupported version %d (expected %d)", ver, wuitVersion)
	}

	// Lengths
	themeLen := binary.LittleEndian.Uint32(data[6:10])
	atlasJSONLen := binary.LittleEndian.Uint32(data[10:14])
	atlasPNGLen := binary.LittleEndian.Uint32(data[14:18])

	needed := uint64(wuitHeaderSize) + uint64(themeLen) + uint64(atlasJSONLen) + uint64(atlasPNGLen)
	if uint64(len(data)) < needed {
		return nil, fmt.Errorf("theme binary: data truncated (have %d bytes, need %d)", len(data), needed)
	}

	off := uint32(wuitHeaderSize)
	result := &DecodedThemeBinary{
		ThemeJSON: data[off : off+themeLen],
	}
	off += themeLen

	if atlasJSONLen > 0 {
		result.AtlasJSON = data[off : off+atlasJSONLen]
	}
	off += atlasJSONLen

	if atlasPNGLen > 0 {
		result.AtlasPNG = data[off : off+atlasPNGLen]
	}

	return result, nil
}
