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
	fmt.Printf("omitEmpty.gStruct(st=%#v) called.\n", st)
	s.p.printf("false\n")
}

func (s *omitEmpty) gPtr(p *Ptr) {
	fmt.Printf("omitEmpty.gPtr(st=%#v) called.\n", p)
	s.p.printf("false\n")
}

func (s *omitEmpty) gSlice(sl *Slice) {
	fmt.Printf("omitEmpty.gSlice(sl=%#v) called.\n", sl)
	s.p.printf("%s", IsEmptySlice(s.varname, sl.vname))
}

func (s *omitEmpty) gArray(a *Array) {
	fmt.Printf("omitEmpty.gSlice(a=%#v) called.\n", a)
	s.p.printf("%s", IsEmptySlice(s.varname, a.vname))
}

func (s *omitEmpty) gMap(m *Map) {
	fmt.Printf("omitEmpty.gSlice(m=%#v) called.\n", m)
	s.p.printf("%s", IsEmptyMap(s.varname, m.vname))
}

func (s *omitEmpty) gBase(b *BaseElem) {
	fmt.Printf("omitEmpty.gBase(a=%#v) called.\n", b)
	switch b.Value {
	case Bytes:
		s.p.printf("%s", IsEmptySlice(s.varname, b.Varname()))
	case String:
		s.p.printf("%s", IsEmptyString(s.varname, b.Varname()))
	case Float32, Float64, Complex64, Complex128, Uint, Uint8, Uint16, Uint32, Uint64, Byte, Int, Int8, Int16, Int32, Int64:
		s.p.printf("%s", IsEmptyNumber(s.varname, b.Varname()))
	case Bool:
		s.p.printf("%s", IsEmptyBool(s.varname, b.Varname()))
	case Time: // time.Time
		s.p.printf("%s", IsEmptyTime(s.varname, b.Varname()))
	case Intf: // interface{}
		// assume, for now, never empty. rarely makes sense to serialize these.
		fallthrough
	default:
		s.p.print("false\n")
	}
}

func IsEmptyNumber(v, f string) string {
	return fmt.Sprintf("(%s.%s == 0) // number, omitempty\n",
		v, f)
}

func IsEmptyString(v, f string) string {
	return fmt.Sprintf("(len(%s.%s) == 0) // string, omitempty\n",
		v, f)
}

func IsEmptyMap(v, f string) string {
	return fmt.Sprintf("(len(%s.%s) == 0) // map, omitempty\n",
		v, f)
}

func IsEmptyBool(v, f string) string {
	return fmt.Sprintf("(!%s.%s) // bool, omitempty\n",
		v, f)
}

func IsEmptySlice(v, f string) string {
	return fmt.Sprintf("(len(%s.%s) == 0) // slice/array, omitempty\n",
		v, f)
}

func IsEmptyTime(v, f string) string {
	return fmt.Sprintf("(%s.%s.IsZero()) // time.Time, omitempty\n",
		v, f)
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
