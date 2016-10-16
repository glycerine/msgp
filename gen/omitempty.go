package gen

import (
	"fmt"
)

func emptyOmitter(p *printer, varname string) *omitEmpty {
	return &omitEmpty{
		p:       p,
		varname: varname,
	}
}

type omitEmpty struct {
	p       *printer
	varname string
}

func (s *omitEmpty) gStruct(st *Struct) {
	//fmt.Printf("omitEmpty.gStruct(st=%#v) called.\n", st)
	s.p.printf("false\n")
}

func (s *omitEmpty) gPtr(p *Ptr) {
	//fmt.Printf("omitEmpty.gPtr(st=%#v) called.\n", p)
	s.p.printf("false\n")
}

func (s *omitEmpty) gSlice(sl *Slice) {
	//fmt.Printf("omitEmpty.gSlice(sl=%#v) called.\n", sl)
	s.p.printf("%s", IsEmptySlice(sl.vname))
}

func (s *omitEmpty) gArray(a *Array) {
	//fmt.Printf("omitEmpty.gArray(a=%#v) called.\n", a)
	s.p.printf("%s", IsEmptySlice(a.vname))
}

func (s *omitEmpty) gMap(m *Map) {
	//fmt.Printf("omitEmpty.gMap(m=%#v) called.\n", m)
	s.p.printf("%s", IsEmptyMap(m.vname))
}

func (s *omitEmpty) gBase(b *BaseElem) {
	//fmt.Printf("omitEmpty.gBase(a=%#v) called.\n", b)
	switch b.Value {
	case Bytes:
		s.p.printf("%s", IsEmptySlice(b.Varname()))
	case String:
		s.p.printf("%s", IsEmptyString(b.Varname()))
	case Float32, Float64, Complex64, Complex128, Uint, Uint8, Uint16, Uint32, Uint64, Byte, Int, Int8, Int16, Int32, Int64:
		s.p.printf("%s", IsEmptyNumber(b.Varname()))
	case Bool:
		s.p.printf("%s", IsEmptyBool(b.Varname()))
	case Time: // time.Time
		s.p.printf("%s", IsEmptyTime(b.Varname()))
	case Intf: // interface{}
		// assume, for now, never empty. rarely makes sense to serialize these.
		fallthrough
	default:
		s.p.print("false\n")
	}
}

func IsEmptyNumber(f string) string {
	return fmt.Sprintf("(%s == 0) // number, omitempty\n",
		f)
}

func IsEmptyString(f string) string {
	return fmt.Sprintf("(len(%s) == 0) // string, omitempty\n",
		f)
}

func IsEmptyMap(f string) string {
	return fmt.Sprintf("(len(%s) == 0) // map, omitempty\n",
		f)
}

func IsEmptyBool(f string) string {
	return fmt.Sprintf("(!%s) // bool, omitempty\n",
		f)
}

func IsEmptySlice(f string) string {
	return fmt.Sprintf("(len(%s) == 0) // slice/array, omitempty\n",
		f)
}

func IsEmptyTime(f string) string {
	return fmt.Sprintf("(%s.IsZero()) // time.Time, omitempty\n",
		f)
}

func (s *Struct) NumOmitEmptyFields() int {
	c := 0
	for i := range s.Fields {
		if s.Fields[i].OmitEmpty {
			c++
		}
	}
	return c
}
