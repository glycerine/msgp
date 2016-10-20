package msgp

// NilBitsStack is a helper for Unmarshal
// methods to track where we are when
// deserializing from empty/nil/missing
// fields.
type NilBitsStack struct {
	// simulate getting nils on the wire
	AlwaysNil     bool
	LifoAlwaysNil []bool
	LifoBts       [][]byte
}

// OnlyNilSlice is a slice that contains
// only the msgpack nil (0xc0) bytes.
var OnlyNilSlice = []byte{mnil}

// AlwaysNilString returns a string representation
// of the internal state of the NilBitsStack for
// debugging purposes.
func (r *NilBitsStack) AlwaysNilString() string {
	s := "bottom: "
	for _, v := range r.LifoAlwaysNil {
		if v {
			s += "T"
		} else {
			s += "f"
		}
	}
	return s
}

// PushAlwaysNil will set r.AlwaysNil to true
// and store bts on the internal stack.
func (r *NilBitsStack) PushAlwaysNil(bts []byte) []byte {
	// save current state
	r.LifoBts = append(r.LifoBts, bts)
	r.LifoAlwaysNil = append(r.LifoAlwaysNil, r.AlwaysNil)
	// set reader r to always return nils
	r.AlwaysNil = true

	return OnlyNilSlice
}

// PopAlwaysNil pops the last []byte off the internal
// stack and returns it. If the stack is empty
// we panic.
func (r *NilBitsStack) PopAlwaysNil() []byte {
	n := len(r.LifoAlwaysNil)
	//fmt.Printf("\n Reader.PopAlwaysNil() called! qlen = %d, '%s'\n",
	//	n, r.AlwaysNilString())
	if n == 0 {
		panic("PopAlwaysNil called on empty lifo")
	}
	a := r.LifoAlwaysNil[n-1]
	r.LifoAlwaysNil = r.LifoAlwaysNil[:n-1]
	r.AlwaysNil = a

	bts := r.LifoBts[n-1]
	r.LifoBts = r.LifoBts[:n-1]
	return bts
}
