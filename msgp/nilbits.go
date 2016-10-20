package msgp

type NilBitsStack struct {
	// simulate getting nils on the wire
	AlwaysNil     bool
	LifoAlwaysNil []bool
	LifoBts       [][]byte
}

var OnlyNilSlice = []byte{mnil}

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

func (r *NilBitsStack) PushAlwaysNil(bts []byte) []byte {
	// save current state
	r.LifoBts = append(r.LifoBts, bts)
	r.LifoAlwaysNil = append(r.LifoAlwaysNil, r.AlwaysNil)
	// set reader r to always return nils
	r.AlwaysNil = true

	return OnlyNilSlice
}

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
