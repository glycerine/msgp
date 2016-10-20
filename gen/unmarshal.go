package gen

import (
	"fmt"
	"io"
	"strconv"

	"github.com/tinylib/msgp/cfg"
)

func unmarshal(w io.Writer, cfg *cfg.MsgpConfig) *unmarshalGen {
	return &unmarshalGen{
		p:   printer{w: w},
		cfg: cfg,
	}
}

type unmarshalGen struct {
	passes
	p        printer
	hasfield bool
	depth    int
	cfg      *cfg.MsgpConfig
}

func (u *unmarshalGen) Method() Method { return Unmarshal }

func (u *unmarshalGen) needsField() {
	if u.hasfield {
		return
	}
	u.p.print("\nvar field []byte; _ = field")
	u.hasfield = true
}

func (u *unmarshalGen) Execute(p Elem) error {
	u.hasfield = false
	if !u.p.ok() {
		return u.p.err
	}
	if !IsPrintable(p) {
		return nil
	}

	u.p.comment("UnmarshalMsg implements msgp.Unmarshaler")

	u.p.printf("\nfunc (%s %s) UnmarshalMsg(bts []byte) (o []byte, err error) {", p.Varname(), methodReceiver(p))
	u.p.printf("\nvar nbs msgp.NilBitsStack; if msgp.IsNil(bts) { bts = nbs.PushAlwaysNil(bts[1:]) }\n")
	next(u, p)
	u.p.print("\no = bts")
	u.p.nakedReturn()
	unsetReceiver(p)
	return u.p.err
}

// does assignment to the variable "name" with the type "base"
func (u *unmarshalGen) assignAndCheck(name string, base string) {
	if !u.p.ok() {
		return
	}
	u.p.printf("\n%s, bts, err = msgp.Read%sBytes(bts)", name, base)
	u.p.print(errcheck)
}

func (u *unmarshalGen) gStruct(s *Struct) {
	u.depth++
	defer func() {
		u.depth--
	}()

	if !u.p.ok() {
		return
	}
	if s.AsTuple {
		u.tuple(s)
	} else {
		u.mapstruct(s)
	}
	return
}

func (u *unmarshalGen) tuple(s *Struct) {

	// open block
	sz := randIdent()
	u.p.declare(sz, u32)
	u.assignAndCheck(sz, arrayHeader)
	u.p.arrayCheck(strconv.Itoa(len(s.Fields)), sz)
	for i := range s.Fields {
		if !u.p.ok() {
			return
		}
		next(u, s.Fields[i].FieldElem)
	}
}

func (u *unmarshalGen) mapstruct(s *Struct) {
	n := len(s.Fields)
	if n == 0 {
		return
	}
	u.needsField()
	k := genSerial()
	tmpl, nStr := genUnmarshalMsgTemplate(k)

	fieldOrder := fmt.Sprintf("\n var unmarshalMsgFieldOrder%s = []string{", nStr)
	for i := range s.Fields {
		fieldOrder += fmt.Sprintf("%q,", s.Fields[i].FieldTag)
	}
	fieldOrder += "}\n"
	u.p.printf("%s", fieldOrder)

	u.p.printf("const maxFields%s = %d\n", nStr, n)

	found := "found" + nStr
	u.p.printf(tmpl)

	for i := range s.Fields {
		u.p.printf("\ncase \"%s\":", s.Fields[i].FieldTag)
		u.p.printf("\n%s[%d]=true;", found, i)
		u.depth++
		next(u, s.Fields[i].FieldElem)
		u.depth--
		if !u.p.ok() {
			return
		}
	}
	u.p.print("\ndefault:\nbts, err = msgp.Skip(bts)")
	u.p.print(errcheck)
	u.p.print("\n}\n}") // close switch and for loop

	u.p.printf("\n if nextMiss%s != -1 { bts = nbs.PopAlwaysNil(); }\n", nStr)
}

func (u *unmarshalGen) gBase(b *BaseElem) {
	if !u.p.ok() {
		return
	}

	refname := b.Varname() // assigned to
	lowered := b.Varname() // passed as argument
	if b.Convert {
		// begin 'tmp' block
		refname = randIdent()
		lowered = b.ToBase() + "(" + lowered + ")"
		u.p.printf("\n{\nvar %s %s", refname, b.BaseType())
	}

	switch b.Value {
	case Bytes:
		u.p.printf("\n if nbs.AlwaysNil { %s = %s[:0]} else { %s, bts, err = msgp.ReadBytesBytes(bts, %s)\n", refname, refname, refname, lowered)
		u.p.print(errcheck)
		u.p.closeblock()
	case Ext:
		u.p.printf("\n if nbs.AlwaysNil { // what here?\n} else {bts, err = msgp.ReadExtensionBytes(bts, %s)\n", lowered)
		u.p.print(errcheck)
		u.p.closeblock()
	case IDENT:
		u.p.printf("\n if nbs.AlwaysNil { %s.UnmarshalMsg(msgp.OnlyNilSlice) } else { bts, err = %s.UnmarshalMsg(bts)\n", lowered, lowered)
		u.p.print(errcheck)
		u.p.closeblock()
	default:
		u.p.printf("\n if nbs.AlwaysNil { %s \n} else {  %s, bts, err = msgp.Read%sBytes(bts)\n", b.ZeroLiteral(refname), refname, b.BaseName())
		u.p.print(errcheck)
		u.p.closeblock()
	}
	if b.Convert {
		// close 'tmp' block
		u.p.printf("\n%s = %s(%s)\n}", b.Varname(), b.FromBase(), refname)
	}
}

func (u *unmarshalGen) gArray(a *Array) {
	if !u.p.ok() {
		return
	}

	// special case for [const]byte objects
	// see decode.go for symmetry
	if be, ok := a.Els.(*BaseElem); ok && be.Value == Byte {
		u.p.printf("\nbts, err = msgp.ReadExactBytes(bts, %s[:])", a.Varname())
		u.p.print(errcheck)
		return
	}

	sz := randIdent()
	u.p.declare(sz, u32)
	u.assignAndCheck(sz, arrayHeader)
	u.p.arrayCheck(a.Size, sz)
	u.p.rangeBlock(a.Index, a.Varname(), u, a.Els)
}

func (u *unmarshalGen) gSlice(s *Slice) {
	if !u.p.ok() {
		return
	}
	u.p.printf("\n if nbs.AlwaysNil { %s \n} else {\n",
		s.ZeroLiteral(`(`+s.Varname()+`)`))
	sz := randIdent()
	u.p.declare(sz, u32)
	u.assignAndCheck(sz, arrayHeader)
	u.p.resizeSlice(sz, s)
	u.p.rangeBlock(s.Index, s.Varname(), u, s.Els)
	u.p.closeblock()
}

func (u *unmarshalGen) gMap(m *Map) {
	u.depth++
	defer func() {
		u.depth--
	}()

	if !u.p.ok() {
		return
	}
	u.p.printf("\n if nbs.AlwaysNil { %s \n} else {\n",
		m.ZeroLiteral(m.Varname()))
	sz := randIdent()
	u.p.declare(sz, u32)
	u.assignAndCheck(sz, mapHeader)

	// allocate or clear map
	u.p.resizeMap(sz, m)

	// loop and get key,value
	u.p.printf("\nfor %s > 0 {", sz)
	u.p.printf("\nvar %s string; var %s %s; %s--", m.Keyidx, m.Validx, m.Value.TypeName(), sz)
	u.assignAndCheck(m.Keyidx, stringTyp)
	next(u, m.Value)
	u.p.mapAssign(m)
	u.p.closeblock()
	u.p.closeblock()
}

func (u *unmarshalGen) gPtr(p *Ptr) {
	u.p.printf("\nif nbs.AlwaysNil { %s = nil } else { if msgp.IsNil(bts) { bts, err = msgp.ReadNilBytes(bts); if err != nil { return }; %s = nil; } else { ", p.Varname(), p.Varname())
	u.p.initPtr(p)
	next(u, p.Value)
	u.p.closeblock()
	u.p.closeblock()
}
