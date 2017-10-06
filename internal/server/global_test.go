package server

import (
	"testing"
)

func TestLeaseTimes(t *testing.T) {
	g := newGlobal()
	g.settings.defaultLeaseTime = 360
	g.unregisteredSettings.defaultLeaseTime = 380
	g.registeredSettings.defaultLeaseTime = 400

	g.settings.maxLeaseTime = 400
	g.unregisteredSettings.maxLeaseTime = 410
	g.registeredSettings.maxLeaseTime = 450

	// Test default lease time
	if d := g.getLeaseTime(0, true); d != 400 {
		t.Errorf("Expected 400 got %d", d)
	}
	if d := g.getLeaseTime(0, false); d != 380 {
		t.Errorf("Expected 380 got %d", d)
	}

	g.registeredSettings.defaultLeaseTime = 0
	g.unregisteredSettings.defaultLeaseTime = 0
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

	g.registeredSettings.maxLeaseTime = 0
	g.unregisteredSettings.maxLeaseTime = 0
	if d := g.getLeaseTime(500, true); d != 400 {
		t.Errorf("Expected 400 got %d", d)
	}
	if d := g.getLeaseTime(500, false); d != 400 {
		t.Errorf("Expected 400 got %d", d)
	}

	// Test max lease time where client asks for less
	g.unregisteredSettings.maxLeaseTime = 410
	g.registeredSettings.maxLeaseTime = 450
	if d := g.getLeaseTime(350, true); d != 350 {
		t.Errorf("Expected 350 got %d", d)
	}
	if d := g.getLeaseTime(350, false); d != 350 {
		t.Errorf("Expected 350 got %d", d)
	}

	g.registeredSettings.maxLeaseTime = 0
	g.unregisteredSettings.maxLeaseTime = 0
	if d := g.getLeaseTime(350, true); d != 350 {
		t.Errorf("Expected 350 got %d", d)
	}
	if d := g.getLeaseTime(350, false); d != 350 {
		t.Errorf("Expected 350 got %d", d)
	}
}
