package fontscan

import (
	"encoding/binary"
	"math/rand"
	"reflect"
	"testing"
)

func TestSerializeDeserialize(t *testing.T) {
	for _, fp := range []Footprint{
		{
			Family: "a strange one",
			Runes:  NewRuneSet(1, 0, 2, 0x789, 0xfffee),
			Aspect: Aspect{1, 200, 0.45},
			Format: OpenType,
		},
		{
			Runes: RuneSet{},
		},
	} {
		b := fp.serializeTo(nil)

		var got Footprint
		n, err := got.deserializeFrom(b)
		if err != nil {
			t.Fatal(err)
		}
		if n != len(b) {
			t.Fatalf("unexpected number of bytes read: %d", n)
		}

		if !reflect.DeepEqual(got, fp) {
			t.Fatalf("unexepected Footprint: %v, expected %v", got, fp)
		}
	}
}

func randomBytes() []byte {
	out := make([]byte, 1000)
	rand.Read(out)
	return out
}

func TestDeserializeInvalid(t *testing.T) {
	for range [50]int{} {
		src := randomBytes()
		if rand.Intn(2) == 0 { // indicate a small string
			binary.BigEndian.PutUint16(src, 10)
		}
		if rand.Intn(2) == 0 { // indicate no string and no rune set
			binary.BigEndian.PutUint16(src, 0)
			binary.BigEndian.PutUint32(src[2:], 0)
			src = src[:8] // truncate to simulate a broken input
		}
		var fp Footprint
		_, err := fp.deserializeFrom(src)
		if err == nil {
			t.Fatal("expected error on random input")
		}
	}
}

func TestFormat_Loader(t *testing.T) {
	tests := []Format{
		OpenType, Type1, PCF,
	}
	for _, ft := range tests {
		if ft.Loader() == nil {
			t.Fatalf("missing loader for %d", ft)
		}
	}

	if Format(0).Loader() != nil {
		t.Fatal("unexpected loader")
	}
	if Format(0).String() != "<format 0>" {
		t.Fatal("unexpected representation")
	}
}
