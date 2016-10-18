package gen

// While this looks like a regular go file,
// it is not. It is formatted as a Go file
// so that we can use the compiler to help
// catch bugs. But it is actually a
// manually instantiated template, where
// trailing_ underscores will be replaced
// by uniquifying numbers.

import (
	"github.com/tinylib/msgp/msgp"
	"time"
)

type DemoType struct {
	Name     string
	BirthDay time.Time
}

func (z *DemoType) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field

	// We treat empty fields as if they were Nil on the wire.
	var fieldOrder_ = []string{"Name", "BirthDay"}

	const maxFields_ = 2

	// *start template here*
	totalEncodedFields_, err := dc.ReadMapHeader()
	if err != nil {
		return
	}
	encodedFieldsLeft_ := totalEncodedFields_
	missingFieldsLeft_ := maxFields_ - totalEncodedFields_

	var nextMiss_ int32 = -1
	var found_ [maxFields_]bool
	var curField_ string

doneWithStruct_:
	// First fill all the encoded fields, then
	// treat the remaining, missing fields, as Nil.
	for encodedFieldsLeft_ > 0 || missingFieldsLeft_ > 0 {
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
		// *finish template here*
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
		} // end switch curField_
	} // end for
	if nextMiss_ != -1 {
		dc.PopAlwaysNil()
	}
	return
}
