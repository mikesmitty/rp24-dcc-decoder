package cv

import "github.com/mikesmitty/rp24-dcc-decoder/pkg/cb"

var _ Handler = (*MockHandler)(nil)

type MockHandler struct {
	returnValue bool
	store       map[uint16]uint8
}

func NewMockHandler(returnValue bool, values map[uint16]uint8) *MockHandler {
	return &MockHandler{
		returnValue: returnValue,
		store:       values,
	}
}

func (m *MockHandler) SetCV(cv uint16, value uint8) bool {
	m.store[cv] = value
	return m.returnValue
}

func (m *MockHandler) CV(cv uint16) uint8 {
	return m.store[cv]
}

func (m *MockHandler) CVOk(cv uint16) (uint8, bool) {
	return m.store[cv], m.returnValue
}

func (m *MockHandler) IndexedCV(index uint16, cv uint16) uint8 {
	return m.store[cv]
}

func (m *MockHandler) IndexedCVOk(index uint16, cv uint16) (uint8, bool) {
	return m.store[cv], m.returnValue
}

func (m *MockHandler) IndexedSet(index uint16, cv uint16, value uint8) bool {
	return m.returnValue
}

func (m *MockHandler) IndexedSetSync(index uint16, cv uint16, value uint8) bool {
	return m.returnValue
}

func (m *MockHandler) Reset(cv uint16) bool {
	return m.returnValue
}

func (m *MockHandler) ResetAll() {
}

func (m *MockHandler) ProcessChanges() {
}

func (m *MockHandler) Set(cv uint16, value uint8) bool {
	return m.returnValue
}

func (m *MockHandler) SetSync(cv uint16, value uint8) bool {
	return m.returnValue
}

func (m *MockHandler) RegisterCallback(cv uint16, fn cb.CVCallbackFunc) {
}

func (m *MockHandler) IndexPage(indexCVs ...uint8) uint16 {
	return 0
}

func (m *MockHandler) LoadIndex(page1, page2 uint8) error {
	return nil
}
