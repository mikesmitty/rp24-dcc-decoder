package iir

type IIRFilter struct {
	Alpha float32
	empty bool
	last  float32
}

func NewIIRFilter(alpha float32) *IIRFilter {
	return &IIRFilter{
		Alpha: alpha,
		empty: true,
		last:  0,
	}
}

func (f *IIRFilter) Filter(values ...float32) float32 {
	if f.empty && len(values) > 0 {
		f.last = values[0]
		f.empty = false
	}
	for _, value := range values {
		f.last = f.Alpha*f.last + (1-f.Alpha)*value
	}
	return f.last
}

func (f *IIRFilter) Output() float32 {
	return f.last
}

func (f *IIRFilter) Reset() {
	f.empty = true
	f.last = 0
}
