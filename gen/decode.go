package gen

import (
	"fmt"
	"github.com/tinylib/msgp/cfg"
	"io"
	"strconv"
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

	s, isStruct := p.(*Struct)

	d.p.comment("DecodeMsg implements msgp.Decodable")
	d.p.comment("We treat empty fields as if we read a Nil from the wire.")
	d.p.printf("\nfunc (%s %s) DecodeMsg(dc *msgp.Reader) (err error) {", p.Varname(), methodReceiver(p))

	// next will increment k, but we want the first, top level DecodeMsg
	// to refer to this same k ...
	next(d, p)

	if isStruct && s.AsTuple {
		d.p.printf("\n dc.AlwaysNil = false\n")
	}
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
//
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
*/
func (d *decodeGen) structAsMap(s *Struct) {
	n := len(s.Fields)
	if n == 0 {
		return
	}
	d.needsField()

	k := genSerial()
	tmpl, nStr := genDecodeMsgTemplate(k)

	// allocate the fieldOrder tracker once, at program startup.
	// This happens just before the struct-specific DecodeMsg method.
	//
	fieldOrder := fmt.Sprintf("\n var decodeMsgFieldOrder%s = []string{", nStr)
	for i := range s.Fields {
		fieldOrder += fmt.Sprintf("%q,", s.Fields[i].FieldTag)
	}
	fieldOrder += "}\n"
	d.p.printf("%s", fieldOrder)

	d.p.printf("const maxFields%s = %d\n", nStr, n)

	found := "found" + nStr
	d.p.printf(tmpl)
	// after printing tmpl, we are at this point:
	// switch curField_ {
	// -- templateDecodeMsg ends here --

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

	d.p.printf("\n if nextMiss%s != -1 {dc.PopAlwaysNil(); }\n", nStr)
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
