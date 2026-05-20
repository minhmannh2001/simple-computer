package components

import "testing"

func TestIOBus_SetUnset(t *testing.T) {
	b := NewIOBus()
	b.Set()
	if !b.IsSet() {
		t.Error("after Set(): IsSet should be true")
	}
	b.Unset()
	if b.IsSet() {
		t.Error("after Unset(): IsSet should be false")
	}
}

func TestIOBus_Enable(t *testing.T) {
	b := NewIOBus()
	b.Enable()
	if !b.IsEnable() {
		t.Error("after Enable(): IsEnable should be true")
	}
}

func TestIOBus_InputDataMode(t *testing.T) {
	b := NewIOBus()
	b.Update(false, false)
	if !b.IsInputMode() {
		t.Error("mode=false: IsInputMode should be true")
	}
	if !b.IsDataMode() {
		t.Error("dataOrAddress=false: IsDataMode should be true")
	}
}

func TestIOBus_OutputAddressMode(t *testing.T) {
	b := NewIOBus()
	b.Update(true, true)
	if !b.IsOutputMode() {
		t.Error("mode=true: IsOutputMode should be true")
	}
	if !b.IsAddressMode() {
		t.Error("dataOrAddress=true: IsAddressMode should be true")
	}
}
