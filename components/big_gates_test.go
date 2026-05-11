package components

import "testing"

func TestANDGate3(t *testing.T) {
	cases := []struct {
		a, b, c bool
		want    bool
	}{
		{false, false, false, false},
		{true, false, false, false},
		{true, true, false, false},
		{true, true, true, true},
	}
	for _, c := range cases {
		g := NewANDGate3()
		g.Update(c.a, c.b, c.c)
		if got := g.Output(); got != c.want {
			t.Errorf("ANDGate3(%v,%v,%v) = %v, want %v", c.a, c.b, c.c, got, c.want)
		}
	}
}

func TestANDGate4(t *testing.T) {
	cases := []struct {
		a, b, c, d bool
		want       bool
	}{
		{false, false, false, false, false},
		{true, true, true, false, false},
		{true, true, true, true, true},
	}
	for _, c := range cases {
		g := NewANDGate4()
		g.Update(c.a, c.b, c.c, c.d)
		if got := g.Output(); got != c.want {
			t.Errorf("ANDGate4(%v,%v,%v,%v) = %v, want %v", c.a, c.b, c.c, c.d, got, c.want)
		}
	}
}

func TestANDGate5(t *testing.T) {
	cases := []struct {
		a, b, c, d, e bool
		want          bool
	}{
		{false, false, false, false, false, false},
		{true, true, true, true, false, false},
		{true, true, true, true, true, true},
	}
	for _, c := range cases {
		g := NewANDGate5()
		g.Update(c.a, c.b, c.c, c.d, c.e)
		if got := g.Output(); got != c.want {
			t.Errorf("ANDGate5 = %v, want %v", got, c.want)
		}
	}
}

func TestANDGate8(t *testing.T) {
	allTrue := [8]bool{true, true, true, true, true, true, true, true}

	g := NewANDGate8()
	g.Update(allTrue[0], allTrue[1], allTrue[2], allTrue[3], allTrue[4], allTrue[5], allTrue[6], allTrue[7])
	if !g.Output() {
		t.Error("ANDGate8(all true) = false, want true")
	}

	// each position false must produce false
	for i := 0; i < 8; i++ {
		inputs := allTrue
		inputs[i] = false
		g.Update(inputs[0], inputs[1], inputs[2], inputs[3], inputs[4], inputs[5], inputs[6], inputs[7])
		if g.Output() {
			t.Errorf("ANDGate8 with input[%d]=false = true, want false", i)
		}
	}

	g.Update(false, false, false, false, false, false, false, false)
	if g.Output() {
		t.Error("ANDGate8(all false) = true, want false")
	}
}

func TestORGate3(t *testing.T) {
	cases := []struct {
		a, b, c bool
		want    bool
	}{
		{false, false, false, false},
		{true, false, false, true},
		{false, true, false, true},
		{false, false, true, true},
		{true, true, true, true},
	}
	for _, c := range cases {
		g := NewORGate3()
		g.Update(c.a, c.b, c.c)
		if got := g.Output(); got != c.want {
			t.Errorf("ORGate3(%v,%v,%v) = %v, want %v", c.a, c.b, c.c, got, c.want)
		}
	}
}

func TestORGate4(t *testing.T) {
	cases := []struct {
		a, b, c, d bool
		want       bool
	}{
		{false, false, false, false, false},
		{false, false, true, false, true},
		{true, true, true, true, true},
	}
	for _, c := range cases {
		g := NewORGate4()
		g.Update(c.a, c.b, c.c, c.d)
		if got := g.Output(); got != c.want {
			t.Errorf("ORGate4(%v,%v,%v,%v) = %v, want %v", c.a, c.b, c.c, c.d, got, c.want)
		}
	}
}

func TestORGate5(t *testing.T) {
	cases := []struct {
		a, b, c, d, e bool
		want          bool
	}{
		{false, false, false, false, false, false},
		{false, false, false, false, true, true},
		{true, true, true, true, true, true},
	}
	for _, c := range cases {
		g := NewORGate5()
		g.Update(c.a, c.b, c.c, c.d, c.e)
		if got := g.Output(); got != c.want {
			t.Errorf("ORGate5 = %v, want %v", got, c.want)
		}
	}
}

func TestORGate6(t *testing.T) {
	g := NewORGate6()

	g.Update(false, false, false, false, false, false)
	if g.Output() {
		t.Error("ORGate6(all false) = true, want false")
	}

	// each single input true must produce true
	for i := 0; i < 6; i++ {
		inputs := [6]bool{}
		inputs[i] = true
		g.Update(inputs[0], inputs[1], inputs[2], inputs[3], inputs[4], inputs[5])
		if !g.Output() {
			t.Errorf("ORGate6 with only input[%d]=true = false, want true", i)
		}
	}

	g.Update(true, true, true, true, true, true)
	if !g.Output() {
		t.Error("ORGate6(all true) = false, want true")
	}
}
