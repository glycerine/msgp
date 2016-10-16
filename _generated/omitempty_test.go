package _generated

import (
	"bytes"
	//	"fmt"
	//	"github.com/shurcooL/go-goon"
	"github.com/tinylib/msgp/msgp"
	"testing"
)

func TestMissingNilledOutWhenUnmarshallingNilIntoNestedStructs(t *testing.T) {

	// UnmarshalMsg
	//
	// Given a tree of structs with element points
	// that is three levels deep, when omitempty fields
	// are omitted from the wire (msgpack data), we
	// should re-use the existing structs, maps, and slices,
	// and shrink them without re-allocation.

	s := TopNester{
		TopId:     43,
		Greetings: "greetings",
		Bullwinkle: &Rocky{
			Road: "Long and winding",
			Bugs: &Bunny{
				Carrots: []int{41, 8},
				Sayings: map[string]string{"whatsup": "doc"},
				BunnyId: 88,
			},
			Moose: &Moose{
				Trees:   []int{0, 1, 2, 3},
				Sayings: map[string]string{"one": "singular sensation"},
				Id:      2,
			},
		},
	}

	// so pointers should not change upon decoding from nil
	pGreet := &s.Greetings
	pBull := &s.Bullwinkle
	pBugs := &s.Bullwinkle.Bugs
	pCarrots := &s.Bullwinkle.Bugs.Carrots
	pSay1 := &s.Bullwinkle.Bugs.Sayings
	pMoose := &s.Bullwinkle.Moose
	pTree := &s.Bullwinkle.Moose.Trees
	pSay2 := &s.Bullwinkle.Moose.Sayings

	//	fmt.Printf("\n ======== BEGIN goon.Dump of TopNester *BEFORE* Unmarshal:\n")
	//	goon.Dump(s)
	//	fmt.Printf("\n ======== END goon.Dump of TopNester *BEFORE* Unmarshal\n")

	nilMsg := []byte{0xc0}
	o, err := s.UnmarshalMsg(nilMsg)
	if err != nil {
		panic(err)
	}

	//	fmt.Printf("\n ======== BEGIN goon.Dump of TopNester AFTER Unmarshal:\n")
	//	goon.Dump(s)
	//	fmt.Printf("\n ======== END goon.Dump of TopNester AFTER Unmarshal\n")

	if len(o) != 0 {
		t.Fatal("nilMsg should have been consumed")
	}
	if pGreet != &s.Greetings {
		t.Fatal("pGreet differed from original")
	}
	if pBull != &s.Bullwinkle {
		t.Fatal("pBull differed from original")
	}
	if pBugs != &s.Bullwinkle.Bugs {
		t.Fatal("pBugs differed from original")
	}
	if pCarrots != &s.Bullwinkle.Bugs.Carrots {
		t.Fatal("pCarrots differed from original")
	}
	if pSay1 != &s.Bullwinkle.Bugs.Sayings {
		t.Fatal("pSay1 differed from original")
	}
	if pMoose != &s.Bullwinkle.Moose {
		t.Fatal("pMoose differed from original")
	}
	if pTree != &s.Bullwinkle.Moose.Trees {
		t.Fatal("pTree differed from original")
	}
	if pSay2 != &s.Bullwinkle.Moose.Sayings {
		t.Fatal("pSay2 differed from original")
	}

	// and yet, the maps and slices should be size 0,
	// the strings empty, the integers zeroed out.

	// TopNester
	if s.TopId != 0 {
		t.Fatal("s.TopId should be 0")
	}
	if len(s.Greetings) != 0 {
		t.Fatal("s.Grettings should be len 0")
	}
	if s.Bullwinkle == nil {
		t.Fatal("s.Bullwinkle should not be nil")
	}

	// TopNester.Bullwinkle
	if s.Bullwinkle.Bugs == nil {
		t.Fatal("s.Bullwinkle.Bugs should not be nil")
	}
	if len(s.Bullwinkle.Road) != 0 {
		t.Fatal("s.Bullwinkle.Road should be len 0")
	}
	if s.Bullwinkle.Moose == nil {
		t.Fatal("s.Bullwinkle.Moose should not be nil")
	}

	// TopNester.Bullwinkle.Bugs
	if len(s.Bullwinkle.Bugs.Carrots) != 0 {
		panic("this is wrong")
	}
	if len(s.Bullwinkle.Bugs.Sayings) != 0 {
		panic("this is wrong")
	}
	if s.Bullwinkle.Bugs.BunnyId != 0 {
		panic("this is wrong")
	}

	// TopNester.Bullwinkle.Moose
	if len(s.Bullwinkle.Moose.Trees) != 0 {
		panic("this is wrong")
	}
	if len(s.Bullwinkle.Moose.Sayings) != 0 {
		panic("this is wrong")
	}
	if s.Bullwinkle.Moose.Id != 0 {
		panic("this is wrong")
	}
}

func TestMissingNilledOutWhenDecodingNilIntoNestedStructs(t *testing.T) {

	// DecodeMsg
	//
	// Given a tree of structs with element points
	// that is three levels deep, when omitempty fields
	// are omitted from the wire (msgpack data), we
	// should re-use the existing structs, maps, and slices,
	// and shrink them without re-allocation.

	s := TopNester{
		TopId:     43,
		Greetings: "greetings",
		Bullwinkle: &Rocky{
			Road: "Long and winding",
			Bugs: &Bunny{
				Carrots: []int{41, 8},
				Sayings: map[string]string{"whatsup": "doc"},
				BunnyId: 88,
			},
			Moose: &Moose{
				Trees:   []int{0, 1, 2, 3},
				Sayings: map[string]string{"one": "singular sensation"},
				Id:      2,
			},
		},
	}

	// so pointers should not change upon decoding from nil
	pGreet := &s.Greetings
	pBull := &s.Bullwinkle
	pBugs := &s.Bullwinkle.Bugs
	pCarrots := &s.Bullwinkle.Bugs.Carrots
	pSay1 := &s.Bullwinkle.Bugs.Sayings
	pMoose := &s.Bullwinkle.Moose
	pTree := &s.Bullwinkle.Moose.Trees
	pSay2 := &s.Bullwinkle.Moose.Sayings

	//	fmt.Printf("\n ======== BEGIN goon.Dump of TopNester *BEFORE* Unmarshal:\n")
	//	goon.Dump(s)
	//	fmt.Printf("\n ======== END goon.Dump of TopNester *BEFORE* Unmarshal\n")

	nilMsg := []byte{0xc0}
	dc := msgp.NewReader(bytes.NewBuffer(nilMsg))
	err := s.DecodeMsg(dc) // msgp: attempted to decode type "nil" with method for "map"
	if err != nil {
		panic(err)
	}

	//	fmt.Printf("\n ======== BEGIN goon.Dump of TopNester AFTER Unmarshal:\n")
	//	goon.Dump(s)
	//	fmt.Printf("\n ======== END goon.Dump of TopNester AFTER Unmarshal\n")

	if pGreet != &s.Greetings {
		t.Fatal("pGreet differed from original")
	}
	if pBull != &s.Bullwinkle {
		t.Fatal("pBull differed from original")
	}
	if pBugs != &s.Bullwinkle.Bugs {
		t.Fatal("pBugs differed from original")
	}
	if pCarrots != &s.Bullwinkle.Bugs.Carrots {
		t.Fatal("pCarrots differed from original")
	}
	if pSay1 != &s.Bullwinkle.Bugs.Sayings {
		t.Fatal("pSay1 differed from original")
	}
	if pMoose != &s.Bullwinkle.Moose {
		t.Fatal("pMoose differed from original")
	}
	if pTree != &s.Bullwinkle.Moose.Trees {
		t.Fatal("pTree differed from original")
	}
	if pSay2 != &s.Bullwinkle.Moose.Sayings {
		t.Fatal("pSay2 differed from original")
	}

	// and yet, the maps and slices should be size 0,
	// the strings empty, the integers zeroed out.

	// TopNester
	if s.TopId != 0 {
		t.Fatal("s.TopId should be 0")
	}
	if len(s.Greetings) != 0 {
		t.Fatal("s.Grettings should be len 0")
	}
	if s.Bullwinkle == nil {
		t.Fatal("s.Bullwinkle should not be nil")
	}

	// TopNester.Bullwinkle
	if s.Bullwinkle.Bugs == nil {
		t.Fatal("s.Bullwinkle.Bugs should not be nil")
	}
	if len(s.Bullwinkle.Road) != 0 {
		t.Fatal("s.Bullwinkle.Road should be len 0")
	}
	if s.Bullwinkle.Moose == nil {
		t.Fatal("s.Bullwinkle.Moose should not be nil")
	}

	// TopNester.Bullwinkle.Bugs
	if len(s.Bullwinkle.Bugs.Carrots) != 0 {
		panic("this is wrong")
	}
	if len(s.Bullwinkle.Bugs.Sayings) != 0 {
		panic("this is wrong")
	}
	if s.Bullwinkle.Bugs.BunnyId != 0 {
		panic("this is wrong")
	}

	// TopNester.Bullwinkle.Moose
	if len(s.Bullwinkle.Moose.Trees) != 0 {
		panic("this is wrong")
	}
	if len(s.Bullwinkle.Moose.Sayings) != 0 {
		panic("this is wrong")
	}
	if s.Bullwinkle.Moose.Id != 0 {
		panic("this is wrong")
	}
}
