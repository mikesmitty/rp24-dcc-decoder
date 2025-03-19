package cv

import "github.com/mikesmitty/rp24-dcc-decoder/pkg/store"

type Handler interface {
	CV(uint16) uint8
	CVOk(uint16) (uint8, bool)
	IndexedCV(uint16, uint16) uint8
	IndexedCVOk(uint16, uint16) (uint8, bool)
	Set(uint16, uint8) bool
	SetSync(uint16, uint8) bool
	IndexedSet(uint16, uint16, uint8) bool
	IndexedSetSync(uint16, uint16, uint8) bool
	Reset(uint16) bool
	ResetAll()
	ProcessChanges()
	SetDefault(uint16, uint8, store.CVFlags)
	RegisterCallback(uint16, func(uint16, uint8) bool)
	IndexPage(...uint8) uint16
}

var _ Handler = (*CVHandler)(nil)

type CVHandler struct {
	cvStore     *store.Store
	cvCallbacks map[uint16][]func(uint16, uint8) bool
}

type CVCallbackFunc func(uint16, uint8) bool

func NewCVHandler(version uint8) *CVHandler {
	s := store.NewStore()
	c := &CVHandler{
		cvCallbacks: make(map[uint16][]func(uint16, uint8) bool),
		cvStore:     s,
	}

	lastVersion, _ := s.CV(7)
	if lastVersion != version {
		/* FIXME: Do stuff if the version has changed
		s.ResetAll()
		*/
	}

	c.setDefaults(s, version)

	return c
}

func (c *CVHandler) RegisterCallback(cvNumber uint16, fn func(cvNumber uint16, value uint8) bool) {
	if _, ok := c.CVOk(cvNumber); !ok {
		// It's not likely this will ever happen, but just to be sure
		panic("CV not found")
	}
	c.cvCallbacks[cvNumber] = append(c.cvCallbacks[cvNumber], fn)
}

// Return the current index page indicated by CV31/32 or a provided equivalent
func (c *CVHandler) IndexPage(indexCVs ...uint8) uint16 {
	// FIXME: Set limits around max index pages

	var cv31, cv32 uint8
	if len(indexCVs) < 2 {
		cv31, _ = c.cvStore.CV(31)
		cv32, _ = c.cvStore.CV(32)
	} else {
		cv31 = indexCVs[0]
		cv32 = indexCVs[1]
	}
	// 00010000
	if cv31 < 16 {
		return 0
	}
	// 00000000 00000000 is page 0
	// 00010000 00000000 is page 1 (257-512)
	return (uint16(cv31-16)<<8 | uint16(cv32)) + 1
}

func (c *CVHandler) CV(cvNumber uint16) uint8 {
	v, _ := c.CVOk(cvNumber)
	return v
}

func (c *CVHandler) CVOk(cvNumber uint16) (uint8, bool) {
	return c.CVOk(cvNumber)
}

// IndexedCV returns the value of a CV given an index page and CV number
func (c *CVHandler) IndexedCV(index, cvNumber uint16) uint8 {
	v, _ := c.IndexedCVOk(index, cvNumber)
	return v
}

// IndexedCVOk returns the value and pre-existence of a CV given an index page and CV number
func (c *CVHandler) IndexedCVOk(index, cvNumber uint16) (uint8, bool) {
	return c.cvStore.IndexedCV(index, cvNumber)
}

// Set sets a CV value and allows it to be written to flash in batches
func (c *CVHandler) Set(cvNumber uint16, value uint8) bool {
	return c.Set(cvNumber, value)
}

// SetSync sets a CV and does not return until it is persisted to flash
func (c *CVHandler) SetSync(cvNumber uint16, value uint8) bool {
	return c.SetSync(cvNumber, value)
}

// IndexedSet sets a CV value given a paging index and allows it to be written to flash in batches
// FIXME: Implement the batching
func (c *CVHandler) IndexedSet(index, cvNumber uint16, value uint8) bool {
	// Check if the CV exists first. Unset CVs are not allowed to be set
	// FIXME: We are the arbiter of what is allowed to be set, check against our cv bitmaps
	prev, ok := c.cvStore.CV(cvNumber)
	if !ok {
		return false
	}
	rejected := false

	// Run the callbacks first to make sure the value isn't rejected
	if callbacks, ok := c.cvCallbacks[cvNumber]; ok {
		for _, fn := range callbacks {
			if !fn(cvNumber, value) {
				rejected = true
			}
		}
		if rejected {
			// Run it back now ya'll
			// Prod the callbacks to roll back their caches
			for _, fn := range callbacks {
				fn(cvNumber, prev)
			}
		}
	}
	return c.cvStore.Set(cvNumber, value)
}

// IndexedSetSync sets a CV given a paging index and does not return until it is persisted to flash
func (c *CVHandler) IndexedSetSync(index, cvNumber uint16, value uint8) bool {
	if !c.IndexedSet(index, cvNumber, value) {
		return false
	}
	return c.cvStore.IndexedPersist(index, cvNumber, value)
}

func (c *CVHandler) indexedCVNumber(index, cvNumber uint16) uint16 {
	switch cvNumber {
	case 31, 32:
	// These CVs ignore the index
	default:
		cvNumber += index * 256
	}
	return cvNumber
}

func (c *CVHandler) Reset(cvNumber uint16) bool {
	return c.cvStore.Reset(cvNumber)
}

func (c *CVHandler) ResetAll() {
	c.cvStore.ResetAll()
}

func (c *CVHandler) ProcessChanges() {
	c.cvStore.ProcessChanges()
}

func (c *CVHandler) SetDefault(cvNumber uint16, value uint8, flags store.CVFlags) {
	c.cvStore.SetDefault(cvNumber, value, flags)
}
