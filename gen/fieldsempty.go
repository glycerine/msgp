package gen

import (
	"fmt"
	"io"
)

func fieldsempty(w io.Writer) *fieldsEmpty {
	return &fieldsEmpty{
		p: printer{w: w},
	}
}

type fieldsEmpty struct {
	passes
	p     printer
	recvr string
}

func (e *fieldsEmpty) Method() Method { return FieldsEmpty }

func (e *fieldsEmpty) Execute(p Elem) error {
	if !e.p.ok() {
		return e.p.err
	}
	p = e.applyall(p)
	if p == nil {
		return nil
	}
	if !IsPrintable(p) {
		return nil
	}

	e.recvr = fmt.Sprintf("%s %s", p.Varname(), imutMethodReceiver(p))

	next(e, p)
	return e.p.err
}

func (e *fieldsEmpty) gStruct(s *Struct) {
	fmt.Printf("fieldsEmpty.gStruct() called.\n")

	e.p.printf("\n\n// FieldsNotEmpty must be provided with an isempty slice\n")
	e.p.printf("// which points to a zero valued array whose size matches\n")
	e.p.printf("// the number of fields in our receiver. We will write\n")
	e.p.printf("// true for isemtpy[i] if the i-th field of our receiver\n")
	e.p.printf("// is empty (nil pointer, length zero map/string/slice,\n")
	e.p.printf("// or a 0 number). We support the omitempty tag.\n")
	e.p.printf("// We return the count of non-empty fields.\n")
	e.p.printf("func (%s) FieldsNotEmpty(isempty []bool) uint32 {", e.recvr)

	nfields := len(s.Fields)
	om := emptyOmitter(&e.p, s.vname)

	e.p.printf("if len(isempty) == 0 { return %v }\n", nfields)
	e.p.printf("var fieldsInUse uint32 = %v\n", nfields)
	for i := range s.Fields {
		if s.Fields[i].OmitEmpty {
			e.p.printf("isempty[%v] = ", i)
			next(om, s.Fields[i].FieldElem)
			e.p.printf("if isempty[%v] { fieldsInUse-- }\n", i)
		}
	}
	e.p.printf("\n return fieldsInUse \n}\n")
}

func (s *fieldsEmpty) gPtr(p *Ptr) {}

func (s *fieldsEmpty) gSlice(sl *Slice) {}

func (s *fieldsEmpty) gArray(a *Array) {}

func (s *fieldsEmpty) gMap(m *Map) {}

func (s *fieldsEmpty) gBase(b *BaseElem) {}
