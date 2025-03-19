package store

import (
	"testing"
)

func TestSetAndGetCV(t *testing.T) {
	cvTable := NewStore()

	if !cvTable.Set(1, 10) {
		t.Errorf("Failed to set CV 1")
	}
	if !cvTable.Set(7, 1) {
		t.Errorf("Failed to set CV 2")
	}
	cvTable.SetReadOnly(7)

	value, ok := cvTable.CV(1)
	if !ok {
		t.Errorf("CV 1 not found")
	}
	if value != 10 {
		t.Errorf("CV 1: expected value 10, got %d", value)
	}

	if cvTable.Set(7, 20) {
		t.Errorf("Set CV 7 should have failed (read-only)")
	}
}

func TestSetReadOnly(t *testing.T) {
	cvTable := NewStore()
	cvTable.SetDefault(1, 3, Volatile)
	cvTable.SetReadOnly(1)

	if cvTable.Set(1, 10) {
		t.Errorf("Set CV 1 should have failed (read-only)")
	}
}

func TestResetCV(t *testing.T) {
	cvTable := NewStore()

	cvTable.SetDefault(1, 3, Volatile)
	cvTable.Set(1, 10)
	cvTable.Reset(1)

	value, ok := cvTable.CV(1)
	if !ok {
		t.Errorf("CV 1 not found")
	}
	if value != 3 {
		t.Errorf("CV 1: expected value 3 after reset, got %d", value)
	}
}

func TestResetAllCVs(t *testing.T) {
	cvTable := NewStore()

	cvTable.SetDefault(1, 3, Volatile)
	cvTable.SetDefault(2, 0, Volatile)
	cvTable.Set(1, 10)
	cvTable.Set(2, 20)
	cvTable.ResetAll()

	tests := []struct {
		cvNumber      uint16
		expectedValue uint8
	}{
		{1, 3},
		{2, 0},
	}

	for _, tt := range tests {
		value, ok := cvTable.CV(tt.cvNumber)
		if !ok {
			t.Errorf("CV %d not found", tt.cvNumber)
		}
		if value != tt.expectedValue {
			t.Errorf("CV %d: expected value %d after reset, got %d", tt.cvNumber, tt.expectedValue, value)
		}
	}
}

func TestProcessChanges(t *testing.T) {
	cvTable := NewStore()

	cvTable.Set(1, 10)
	cvTable.Set(2, 20)
	cvTable.ProcessChanges()

	// Add assertions based on what ProcessChanges is supposed to do
	// For example, if it should clear the Dirty flag:
	data := cvTable.data[1]
	if (data.Flags & Dirty) != 0 {
		t.Errorf("CV 1: Dirty flag should be cleared")
	}
	data = cvTable.data[2]
	if (data.Flags & Dirty) != 0 {
		t.Errorf("CV 2: Dirty flag should be cleared")
	}
}
