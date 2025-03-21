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
	RegisterCallback(uint16, func(uint16, uint8) bool)
	IndexPage(...uint8) uint16
}

var _ Handler = (*CVHandler)(nil)

type CVHandler struct {
	cvStore     *store.Store
	cvCallbacks map[uint16][]func(uint16, uint8) bool
}

type CVCallbackFunc func(uint16, uint8) bool

var fwVersion []uint8

func NewCVHandler(version []uint8) *CVHandler {
	s := store.NewStore()
	c := &CVHandler{
		cvCallbacks: make(map[uint16][]func(uint16, uint8) bool),
		cvStore:     s,
	}

	if len(version) != 3 {
		panic("invalid version length")
	}

	// Check if the version has changed
	/* TODO: Check major/minor/patch version
	lastVersion, _ := s.IndexedCV(0, 7)
	if lastVersion != version {
		// TODO: Do stuff if the version has changed
	}
	*/

	// TODO: Handle CVs > 256 properly. Treating CVs as if the index is separate
	// makes handling callback functions a big problem. Also, defaults get kinda
	// wonky. Need to refactor for that change if/when CVs > 256 are implemented

	// Load the last-used index from flash
	cv31, _ := s.IndexedCV(0, 31)
	cv32, _ := s.IndexedCV(0, 32)
	err := c.LoadIndex(cv31, cv32)
	if err != nil {
		println("could not load index: " + err.Error())
	}

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
	// TODO: Set limits around max index pages when indexes are implemented

	var cv31, cv32 uint8
	if len(indexCVs) < 2 {
		cv31, _ = c.cvStore.IndexedCV(0, 31)
		cv32, _ = c.cvStore.IndexedCV(0, 32)
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
	return c.IndexedCV(c.IndexPage(), cvNumber)
}

func (c *CVHandler) CVOk(cvNumber uint16) (uint8, bool) {
	return c.IndexedCVOk(c.IndexPage(), cvNumber)
}

// IndexedCV returns the value of a CV given an index page and CV number
func (c *CVHandler) IndexedCV(index, cvNumber uint16) uint8 {
	v, _ := c.IndexedCVOk(index, cvNumber)
	return v
}

// IndexedCVOk returns the value and pre-existence of a CV given an index page and CV number
func (c *CVHandler) IndexedCVOk(index, cvNumber uint16) (uint8, bool) {
	// Keep CV31/32 reads/writes constrained to index 0
	if cvNumber == 31 || cvNumber == 32 {
		index = 0
	}
	return c.cvStore.IndexedCV(index, cvNumber)
}

// Set sets a CV value and allows it to be written to flash in batches
func (c *CVHandler) Set(cvNumber uint16, value uint8) bool {
	if cvNumber == 31 || cvNumber == 32 {
		// Don't allow setting of CV31/32 directly, only allow through Config Variable Access commands
		return false
	}
	return c.cvStore.Set(cvNumber, value)
}

// SetSync sets a CV and does not return until it is persisted to flash
func (c *CVHandler) SetSync(cvNumber uint16, value uint8) bool {
	ok := c.Set(cvNumber, value)
	if !ok {
		return false
	}
	return c.cvStore.Persist(cvNumber, value)
}

// IndexedSet sets a CV value given a paging index and allows it to be written to flash in batches
func (c *CVHandler) IndexedSet(index, cvNumber uint16, value uint8) bool {
	// Ignore indexes beyond those we support
	if index > maxCVIndexPage {
		return false
	}

	// Check if the CV exists. Unset CVs are not allowed to be set
	// TODO: Need some way of checking if a CV in another index is valid before using higher level CVs
	prev, ok := c.IndexedCVOk(index, cvNumber)
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
			return false
		}
	}
	return c.IndexedSet(index, cvNumber, value)
}

// IndexedSetSync sets a CV given a paging index and does not return until it is persisted to flash
func (c *CVHandler) IndexedSetSync(index, cvNumber uint16, value uint8) bool {
	if !c.IndexedSet(index, cvNumber, value) {
		return false
	}
	return c.cvStore.IndexedPersist(index, cvNumber, value)
}

func (c *CVHandler) Reset(cvNumber uint16) bool {
	if cvNumber == 31 || cvNumber == 32 {
		// Don't allow resetting of CV31/32 directly, only allow through Config Variable Access commands
		return true
	}
	return c.cvStore.Reset(cvNumber)
}

func (c *CVHandler) ResetAll() {
	c.cvStore.ResetAll()
}

func (c *CVHandler) ProcessChanges() {
	c.cvStore.ProcessChanges()
}
