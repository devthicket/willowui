package theme

import (
	"bytes"
	"testing"
)

// ── Roundtrip ───────────────────────────────────────────────────────────────

func TestBinaryRoundtrip(t *testing.T) {
	themeJSON := []byte(`{"name":"test","colors":{}}`)
	atlasJSON := []byte(`{"frames":{}}`)
	atlasPNG := []byte{0x89, 0x50, 0x4E, 0x47} // fake PNG header

	encoded, err := EncodeThemeBinary(themeJSON, atlasJSON, atlasPNG)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := DecodeThemeBinary(encoded)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(decoded.ThemeJSON, themeJSON) {
		t.Errorf("ThemeJSON mismatch: got %q, want %q", decoded.ThemeJSON, themeJSON)
	}
	if !bytes.Equal(decoded.AtlasJSON, atlasJSON) {
		t.Errorf("AtlasJSON mismatch: got %q, want %q", decoded.AtlasJSON, atlasJSON)
	}
	if !bytes.Equal(decoded.AtlasPNG, atlasPNG) {
		t.Errorf("AtlasPNG mismatch: got %x, want %x", decoded.AtlasPNG, atlasPNG)
	}
}

func TestBinaryRoundtrip_NoAtlas(t *testing.T) {
	themeJSON := []byte(`{"name":"minimal"}`)

	encoded, err := EncodeThemeBinary(themeJSON, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := DecodeThemeBinary(encoded)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(decoded.ThemeJSON, themeJSON) {
		t.Errorf("ThemeJSON mismatch")
	}
	if decoded.AtlasJSON != nil {
		t.Errorf("expected nil AtlasJSON, got %v", decoded.AtlasJSON)
	}
	if decoded.AtlasPNG != nil {
		t.Errorf("expected nil AtlasPNG, got %v", decoded.AtlasPNG)
	}
}

// ── Encode errors ───────────────────────────────────────────────────────────

func TestEncode_NilThemeJSON(t *testing.T) {
	_, err := EncodeThemeBinary(nil, nil, nil)
	if err == nil {
		t.Error("expected error for nil themeJSON")
	}
}

// ── Decode errors ───────────────────────────────────────────────────────────

func TestDecode_TooShort(t *testing.T) {
	_, err := DecodeThemeBinary([]byte{0, 1, 2})
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestDecode_BadMagic(t *testing.T) {
	data := make([]byte, wuitHeaderSize)
	copy(data[0:4], []byte("NOPE"))
	_, err := DecodeThemeBinary(data)
	if err == nil {
		t.Error("expected error for bad magic")
	}
}

func TestDecode_BadVersion(t *testing.T) {
	data := make([]byte, wuitHeaderSize)
	copy(data[0:4], wuitMagic[:])
	data[4] = 99 // bad version
	data[5] = 0
	_, err := DecodeThemeBinary(data)
	if err == nil {
		t.Error("expected error for bad version")
	}
}

func TestDecode_Truncated(t *testing.T) {
	themeJSON := []byte(`{"test":true}`)
	encoded, err := EncodeThemeBinary(themeJSON, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Chop off last byte
	_, err = DecodeThemeBinary(encoded[:len(encoded)-1])
	if err == nil {
		t.Error("expected error for truncated data")
	}
}

// ── Header structure ────────────────────────────────────────────────────────

func TestEncode_HeaderMagic(t *testing.T) {
	encoded, err := EncodeThemeBinary([]byte(`{}`), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded[0:4]) != "WUIT" {
		t.Errorf("magic = %q, want WUIT", string(encoded[0:4]))
	}
}

func TestEncode_HeaderSize(t *testing.T) {
	themeJSON := []byte(`{"a":"b"}`)
	atlas := []byte(`{"x":1}`)
	png := []byte{1, 2, 3, 4, 5}

	encoded, err := EncodeThemeBinary(themeJSON, atlas, png)
	if err != nil {
		t.Fatal(err)
	}

	expectedLen := wuitHeaderSize + len(themeJSON) + len(atlas) + len(png)
	if len(encoded) != expectedLen {
		t.Errorf("encoded length = %d, want %d", len(encoded), expectedLen)
	}
}
