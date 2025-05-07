package sconfig

import (
	"testing"
)

func TestLeaseTimes(t *testing.T) {
	g := newGlobal()
	g.settings.DefaultLeaseTime = 360
	g.UnregisteredSettings.DefaultLeaseTime = 380
	g.RegisteredSettings.DefaultLeaseTime = 400

	g.settings.MaxLeaseTime = 400
	g.UnregisteredSettings.MaxLeaseTime = 410
	g.RegisteredSettings.MaxLeaseTime = 450

	// Test default lease time
	if d := g.getLeaseTime(0, true); d != 400 {
		t.Errorf("Expected 400 got %d", d)
	}
	if d := g.getLeaseTime(0, false); d != 380 {
		t.Errorf("Expected 380 got %d", d)
	}

	g.RegisteredSettings.DefaultLeaseTime = 0
	g.UnregisteredSettings.DefaultLeaseTime = 0
	if d := g.getLeaseTime(0, true); d != 360 {
		t.Errorf("Expected 360 got %d", d)
	}
	if d := g.getLeaseTime(0, false); d != 360 {
		t.Errorf("Expected 360 got %d", d)
	}

	// Test max lease time where client asks for too much
	if d := g.getLeaseTime(500, true); d != 450 {
		t.Errorf("Expected 450 got %d", d)
	}
	if d := g.getLeaseTime(500, false); d != 410 {
		t.Errorf("Expected 410 got %d", d)
	}

	g.RegisteredSettings.MaxLeaseTime = 0
	g.UnregisteredSettings.MaxLeaseTime = 0
	if d := g.getLeaseTime(500, true); d != 400 {
		t.Errorf("Expected 400 got %d", d)
	}
	if d := g.getLeaseTime(500, false); d != 400 {
		t.Errorf("Expected 400 got %d", d)
	}

	// Test max lease time where client asks for less
	g.UnregisteredSettings.MaxLeaseTime = 410
	g.RegisteredSettings.MaxLeaseTime = 450
	if d := g.getLeaseTime(350, true); d != 350 {
		t.Errorf("Expected 350 got %d", d)
	}
	if d := g.getLeaseTime(350, false); d != 350 {
		t.Errorf("Expected 350 got %d", d)
	}

	g.RegisteredSettings.MaxLeaseTime = 0
	g.UnregisteredSettings.MaxLeaseTime = 0
	if d := g.getLeaseTime(350, true); d != 350 {
		t.Errorf("Expected 350 got %d", d)
	}
	if d := g.getLeaseTime(350, false); d != 350 {
		t.Errorf("Expected 350 got %d", d)
	}
}
