package gen

import (
	"fmt"
	"github.com/tinylib/msgp/cfg"
	"io"
	"strconv"
	"strings"
)

func decode(w io.Writer, cfg *cfg.MsgpConfig) *decodeGen {
	return &decodeGen{
		p:        printer{w: w},
		hasfield: false,
		cfg:      cfg,
	}
}

type decodeGen struct {
	passes
	p        printer
	hasfield bool
	depth    int
	cfg      *cfg.MsgpConfig
	lifo     []bool
}

func (d *decodeGen) Method() Method { return Decode }

func (d *decodeGen) needsField() {
	if d.hasfield {
		return
	}
	d.p.print("\nvar field []byte; _ = field")
	d.hasfield = true
}

func (d *decodeGen) Execute(p Elem) error {
	p = d.applyall(p)
	if p == nil {
		return nil
	}
	d.hasfield = false
	if !d.p.ok() {
		return d.p.err
	}

	if !IsPrintable(p) {
		return nil
	}

	d.p.comment("DecodeMsg implements msgp.Decodable")

	d.p.printf("\nfunc (%s %s) DecodeMsg(dc *msgp.Reader) (err error) {", p.Varname(), methodReceiver(p))

	next(d, p)
	d.p.printf("\n dc.AlwaysNil = false\n")
	d.p.nakedReturn()
	unsetReceiver(p)
	return d.p.err
}

func (d *decodeGen) gStruct(s *Struct) {
	d.depth++
	defer func() {
		d.depth--
	}()

	if !d.p.ok() {
		return
	}
	if s.AsTuple {
		d.structAsTuple(s)
	} else {
		d.structAsMap(s)
	}
	return
}

func (d *decodeGen) assignAndCheck(name string, typ string) {
	if !d.p.ok() {
		return
	}
	d.p.printf("\n%s, err = dc.Read%s()", name, typ)
	d.p.print(errcheck)
}

func (d *decodeGen) structAsTuple(s *Struct) {
	nfields := len(s.Fields)

	sz := randIdent()
	d.p.declare(sz, u32)
	d.assignAndCheck(sz, arrayHeader)
	d.p.arrayCheck(strconv.Itoa(nfields), sz)
	for i := range s.Fields {
		if !d.p.ok() {
			return
		}
		next(d, s.Fields[i].FieldElem)
	}
}

/* func (d *decodeGen) structAsMap(s *Struct):

    // Missing (empty) field handling logic:
    //
    // The approach to missing field handling is to
    // keep the logic the same whether the field is
    // missing or nil on the wire. To do so we use
    // the Reader.PushAlwaysNil() method to tell
    // the Reader to pretend to supply
    // only nils until further notice. The further
    // notice comes from the terminating dc.PopAlwaysNil()
    // calls emptying the LIFO. The stack is
    // needed because multiple struct decodes may
    // be nested due to inlining.

	var fieldOrder_ = []string{"Name", "BirthDay"}
	const maxFields_ = 2

    totalEncodedFields_, err := dc.ReadMapHeader()
	if err != nil {
		return
	}
	encodedFieldsLeft_ := totalEncodedFields_
    missingFieldsLeft_ := maxFields_ - totalEncodedFields_

	var nextMiss_ int32 = -1
	var found_ [maxFields_]bool
	var field []byte
    _ = field
	var curField_ string

 doneWithStruct_:
	for encodedFieldsLeft_ > 0 || missingFieldsLeft_ > 0 {

         // First phase: do all the available fields.
         // Only after all available have been handled
         // do we embark on the second phase: missing
         // field handling.

		if encodedFieldsLeft_ > 0 {
			encodedFieldsLeft_--
			field, err = dc.ReadMapKeyPtr()
			if err != nil {
				return
			}
			curField_ = msgp.UnsafeString(field)
		} else {
			//missing fields need handling
			if nextMiss_ < 0 {
                // tell the reader to only give us Nils
                // until further notice.
				dc.PushAlwaysNil()
				nextMiss_ = 0
			}
			for nextMiss_ < maxFields_ && found_[nextMiss_] {
				nextMiss_++
			}
			if nextMiss_ == maxFields_ {
				// filled all the empty fields!
				break doneWithStruct_
			}
			missingFieldsLeft_--
			curField_ = fieldOrder_[nextMiss_]
		}

		switch curField_ {
		case "Name":
			found_[0] = true
			z.Name, err = dc.ReadString()
			if err != nil {
				return
			}
		case "BirthDay":
			found_[1] = true
			z.BirthDay, err = dc.ReadTime()
			if err != nil {
				return
			}

        } // end switch curField

    } // end for
	if nextMiss_ != -1 {
		dc.PopAlwaysNil()
	}

*/
func (d *decodeGen) structAsMap(s *Struct) {

	// Label to break to
	lab := "lab_" + randIdent()

	// fieldsInOrder
	tn := s.TypeName()
	if strings.HasPrefix(tn, "struct") {
		tn = "anon"
	}
	fieldsInOrder := "fieldOrder_" + tn + "_" + randIdent()
	var map0 string
	map0 = fmt.Sprintf(`var %s = []string{`, fieldsInOrder)
	for i := range s.Fields {
		map0 += fmt.Sprintf("%q,", s.Fields[i].FieldTag)
	}
	map0 += `}`
	d.p.printf("\n%s\n", map0)

	n := len(s.Fields)
	needNull := "needNull_" + randIdent()
	nextNull := "nextNull_" + randIdent()

	ns := "maxFields_" + randIdent()
	d.p.printf("const %s = %d\n", ns, n)

	f := "curField_" + randIdent()
	d.p.printf("var %s uint32 = %s\n", needNull, ns)
	d.p.printf("var %s int32 = -1\n", nextNull)
	found := "found_" + randIdent()
	d.p.printf("\nvar %s [%s]bool\n", found, ns)

	d.needsField()
	sz := randIdent()
	d.p.declare(sz, u32)
	d.p.printf("\nvar %s string\n", f)
	d.assignAndCheck(sz, mapHeader)

	d.p.printf("\n %s -= %s\n", needNull, sz)
	d.p.printf("%s:\nfor %s > 0 || %s > 0 {\n", lab, sz, needNull)
	//d.p.printf("fmt.Printf(\"\\n ...top of '%s' loop, sz=%%d, needNull=%%d... nextNull=%%d   depth=%d   ...lifo:'%%s'\\n\", %s, %s, %s, dc.AlwaysNilString())\n", lab, d.depth, sz, needNull, nextNull)
	d.p.printf("if %s > 0 {%s--\n", sz, sz)
	d.assignAndCheck("field", mapKey)
	d.p.printf("\n %s = msgp.UnsafeString(field)\n", f)
	d.p.printf("}else{\n //missing field handling\n if %s < 0 { dc.PushAlwaysNil(); %s = 0 }\n for %s < %s && %s[%s] { %s++ }\n", nextNull, nextNull, nextNull, ns, found, nextNull, nextNull)
	d.p.printf("if %s == %s {\n break %s }\n", nextNull, ns, lab)
	d.p.printf(" %s--\n", needNull)
	d.p.printf(" %s = %s[%s]\n}\n", f, fieldsInOrder, nextNull)
	d.p.printf("\nswitch %s {", f)
	for i := range s.Fields {
		d.p.printf("\ncase \"%s\":", s.Fields[i].FieldTag)
		d.p.printf("\n%s[%d]=true;", found, i)
		//d.p.printf("\n fmt.Printf(\"I found field '%s' at depth=%d. dc.AlwaysNil = %%v\", dc.AlwaysNil);\n", s.Fields[i].FieldTag, d.depth)
		d.depth++
		next(d, s.Fields[i].FieldElem)
		d.depth--
		if !d.p.ok() {
			return
		}
	}
	d.p.print("\ndefault:\nerr = dc.Skip()")
	d.p.print(errcheck)
	d.p.closeblock() // close switch
	d.p.closeblock() // close for loop

	d.p.printf("\n if %s != -1 {dc.PopAlwaysNil(); }\n", nextNull)
}

func (d *decodeGen) gBase(b *BaseElem) {
	if !d.p.ok() {
		return
	}

	// open block for 'tmp'
	var tmp string
	if b.Convert {
		tmp = randIdent()
		d.p.printf("\n{ var %s %s", tmp, b.BaseType())
	}

	vname := b.Varname()  // e.g. "z.FieldOne"
	bname := b.BaseName() // e.g. "Float64"

	// handle special cases
	// for object type.
	switch b.Value {
	case Bytes:
		if b.Convert {
			d.p.printf("\n%s, err = dc.ReadBytes([]byte(%s))", tmp, vname)
		} else {
			d.p.printf("\n%s, err = dc.ReadBytes(%s)", vname, vname)
		}
	case IDENT:
		d.p.printf("\nerr = %s.DecodeMsg(dc)", vname)
	case Ext:
		d.p.printf("\nerr = dc.ReadExtension(%s)", vname)
	default:
		if b.Convert {
			d.p.printf("\n%s, err = dc.Read%s()", tmp, bname)
		} else {
			d.p.printf("\n%s, err = dc.Read%s()", vname, bname)
		}
	}

	// close block for 'tmp'
	if b.Convert {
		d.p.printf("\n%s = %s(%s)\n}", vname, b.FromBase(), tmp)
	}

	d.p.print(errcheck)
}

func (d *decodeGen) gMap(m *Map) {
	d.depth++
	defer func() {
		d.depth--
	}()

	if !d.p.ok() {
		return
	}
	sz := randIdent()

	// resize or allocate map
	d.p.declare(sz, u32)
	d.assignAndCheck(sz, mapHeader)
	d.p.resizeMap(sz, m)

	// for element in map, read string/value
	// pair and assign
	d.p.printf("\nfor %s > 0 {\n%s--", sz, sz)
	d.p.declare(m.Keyidx, "string")
	d.p.declare(m.Validx, m.Value.TypeName())
	d.assignAndCheck(m.Keyidx, stringTyp)
	next(d, m.Value)
	d.p.mapAssign(m)
	d.p.closeblock()
}

func (d *decodeGen) gSlice(s *Slice) {
	if !d.p.ok() {
		return
	}
	sz := randIdent()
	d.p.declare(sz, u32)
	d.assignAndCheck(sz, arrayHeader)
	d.p.resizeSlice(sz, s)
	d.p.rangeBlock(s.Index, s.Varname(), d, s.Els)
}

func (d *decodeGen) gArray(a *Array) {
	if !d.p.ok() {
		return
	}

	// special case if we have [const]byte
	if be, ok := a.Els.(*BaseElem); ok && (be.Value == Byte || be.Value == Uint8) {
		d.p.printf("\nerr = dc.ReadExactBytes(%s[:])", a.Varname())
		d.p.print(errcheck)
		return
	}
	sz := randIdent()
	d.p.declare(sz, u32)
	d.assignAndCheck(sz, arrayHeader)
	d.p.arrayCheck(a.Size, sz)

	d.p.rangeBlock(a.Index, a.Varname(), d, a.Els)
}

func (d *decodeGen) gPtr(p *Ptr) {
	if !d.p.ok() {
		return
	}
	d.p.print("\nif dc.IsNil() {")
	d.p.print("\nerr = dc.ReadNil()")
	d.p.print(errcheck)
	d.p.printf("\n%s = nil\n} else {", p.Varname())
	d.p.initPtr(p)
	next(d, p.Value)
	d.p.closeblock()
}
