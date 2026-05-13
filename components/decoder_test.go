package components

import "testing"

// --- Decoder2x4 ---

func TestDecoder2x4AllCombinations(t *testing.T) {
	cases := []struct {
		a, b    bool
		wantOut int
	}{
		{false, false, 0},
		{false, true, 1},
		{true, false, 2},
		{true, true, 3},
	}
	for _, c := range cases {
		d := NewDecoder2x4()
		d.Update(c.a, c.b)
		for i := range 4 {
			got := d.GetOutputWire(i)
			want := i == c.wantOut
			if got != want {
				t.Errorf("Decoder2x4(%v,%v) out[%d]=%v, want %v", c.a, c.b, i, got, want)
			}
		}
	}
}

// --- Decoder3x8 ---

func TestDecoder3x8AllCombinations(t *testing.T) {
	cases := []struct {
		a, b, c bool
		wantIdx int
	}{
		{false, false, false, 0},
		{false, false, true, 1},
		{false, true, false, 2},
		{false, true, true, 3},
		{true, false, false, 4},
		{true, false, true, 5},
		{true, true, false, 6},
		{true, true, true, 7},
	}
	for _, c := range cases {
		d := NewDecoder3x8()
		d.Update(c.a, c.b, c.c)
		for i := range 8 {
			got := d.GetOutputWire(i)
			want := i == c.wantIdx
			if got != want {
				t.Errorf("Decoder3x8(%v,%v,%v) out[%d]=%v, want %v", c.a, c.b, c.c, i, got, want)
			}
		}
		if d.Index() != c.wantIdx {
			t.Errorf("Decoder3x8(%v,%v,%v) Index()=%d, want %d", c.a, c.b, c.c, d.Index(), c.wantIdx)
		}
	}
}

// --- Decoder4x16 ---

func TestDecoder4x16Zero(t *testing.T) {
	d := NewDecoder4x16()
	d.Update(false, false, false, false)
	if !d.GetOutputWire(0) {
		t.Error("Decoder4x16(0000): out[0] should be true")
	}
	for i := 1; i < 16; i++ {
		if d.GetOutputWire(i) {
			t.Errorf("Decoder4x16(0000): out[%d] should be false", i)
		}
	}
	if d.Index() != 0 {
		t.Errorf("Decoder4x16(0000): Index()=%d, want 0", d.Index())
	}
}

func TestDecoder4x16Max(t *testing.T) {
	d := NewDecoder4x16()
	d.Update(true, true, true, true)
	if !d.GetOutputWire(15) {
		t.Error("Decoder4x16(1111): out[15] should be true")
	}
	for i := range 15 {
		if d.GetOutputWire(i) {
			t.Errorf("Decoder4x16(1111): out[%d] should be false", i)
		}
	}
	if d.Index() != 15 {
		t.Errorf("Decoder4x16(1111): Index()=%d, want 15", d.Index())
	}
}

func TestDecoder4x16AllCombinations(t *testing.T) {
	inputs := [16][4]bool{}
	for i := range 16 {
		inputs[i] = [4]bool{i&8 != 0, i&4 != 0, i&2 != 0, i&1 != 0}
	}
	for wantIdx, bits := range inputs {
		d := NewDecoder4x16()
		d.Update(bits[0], bits[1], bits[2], bits[3])
		if d.Index() != wantIdx {
			t.Errorf("Decoder4x16 input %04b: Index()=%d, want %d", wantIdx, d.Index(), wantIdx)
		}
		if !d.GetOutputWire(wantIdx) {
			t.Errorf("Decoder4x16 input %04b: out[%d] should be true", wantIdx, wantIdx)
		}
	}
}

// --- Decoder8x256 ---

// bitsFromByte unpacks a byte into 8 bools, MSB first (a=bit7 ... h=bit0).
func bitsFromByte(b byte) (a, bb, c, d, e, f, g, h bool) {
	return b&0x80 != 0, b&0x40 != 0, b&0x20 != 0, b&0x10 != 0,
		b&0x08 != 0, b&0x04 != 0, b&0x02 != 0, b&0x01 != 0
}

func TestDecoder8x256Zero(t *testing.T) {
	d := NewDecoder8x256()
	d.Update(bitsFromByte(0x00))
	if d.Index() != 0 {
		t.Errorf("Decoder8x256(0x00): Index()=%d, want 0", d.Index())
	}
}

func TestDecoder8x256One(t *testing.T) {
	d := NewDecoder8x256()
	d.Update(bitsFromByte(0x01))
	if d.Index() != 1 {
		t.Errorf("Decoder8x256(0x01): Index()=%d, want 1", d.Index())
	}
}

func TestDecoder8x256Bit7(t *testing.T) {
	d := NewDecoder8x256()
	d.Update(bitsFromByte(0x80))
	if d.Index() != 128 {
		t.Errorf("Decoder8x256(0x80): Index()=%d, want 128", d.Index())
	}
}

func TestDecoder8x256Max(t *testing.T) {
	d := NewDecoder8x256()
	d.Update(bitsFromByte(0xFF))
	if d.Index() != 255 {
		t.Errorf("Decoder8x256(0xFF): Index()=%d, want 255", d.Index())
	}
}

func TestDecoder8x256MidValue(t *testing.T) {
	d := NewDecoder8x256()
	d.Update(bitsFromByte(0x7F)) // 0111 1111 → 127
	if d.Index() != 127 {
		t.Errorf("Decoder8x256(0x7F): Index()=%d, want 127", d.Index())
	}
}

func TestDecoder8x256InRange(t *testing.T) {
	d := NewDecoder8x256()
	for v := range 256 {
		d.Update(bitsFromByte(byte(v)))
		if d.Index() != v {
			t.Errorf("Decoder8x256(0x%02X): Index()=%d, want %d", v, d.Index(), v)
		}
	}
}
