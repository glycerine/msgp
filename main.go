// msgp is a code generation tool for
// creating methods to serialize and de-serialize
// Go data structures to and from MessagePack.
//
// This package is targeted at the `go generate` tool.
// To use it, include the following directive in a
// go source file with types requiring source generation:
//
//     //go:generate msgp
//
// The go generate tool should set the proper environment variables for
// the generator to execute without any command-line flags. However, the
// following options are supported, if you need them:
//
//  -o = output file name (default is {input}_gen.go)
//  -file = input file name (or directory; default is $GOFILE, which is set by the `go generate` command)
//  -io = satisfy the `msgp.Decodable` and `msgp.Encodable` interfaces (default is true)
//  -marshal = satisfy the `msgp.Marshaler` and `msgp.Unmarshaler` interfaces (default is true)
//  -tests = generate tests and benchmarks (default is true)
//
// For more information, please read README.md, and the wiki at github.com/glycerine/msgp
//
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/glycerine/msgp/cfg"
	"github.com/glycerine/msgp/gen"
	"github.com/glycerine/msgp/parse"
	"github.com/glycerine/msgp/printer"
	"github.com/ttacon/chalk"
)

// demonstrate the sequence of calls to DefineFlags() and ValidateConfig()
func main() {

	myflags := flag.NewFlagSet("msgp", flag.ExitOnError)
	c := &cfg.MsgpConfig{}
	c.DefineFlags(myflags)

	err := myflags.Parse(os.Args[1:])
	err = c.ValidateConfig()
	if err != nil {
		log.Fatalf("msgp command line flag error: '%s'", err)
	}

	// GOFILE is set by go generate
	if c.GoFile == "" {
		c.GoFile = os.Getenv("GOFILE")
		if c.GoFile == "" {
			fmt.Println(chalk.Red.Color("No file to parse."))
			os.Exit(1)
		}
	}

	var mode gen.Method
	if c.Encode {
		mode |= (gen.Encode | gen.Decode | gen.Size)
	}
	if c.Marshal {
		mode |= (gen.Marshal | gen.Unmarshal | gen.Size)
	}
	if c.Tests {
		mode |= gen.Test
	}

	if mode&^gen.Test == 0 {
		fmt.Println(chalk.Red.Color("No methods to generate; -io=false && -marshal=false"))
		os.Exit(1)
	}

	if err := Run(mode, c); err != nil {
		fmt.Println(chalk.Red.Color(err.Error()))
		os.Exit(1)
	}
}

// Run writes all methods using the associated file or path, e.g.
//
//	err := msgp.Run("path/to/myfile.go", gen.Size|gen.Marshal|gen.Unmarshal|gen.Test, false)
//
func Run(mode gen.Method, c *cfg.MsgpConfig) error {
	if mode&^gen.Test == 0 {
		return nil
	}
	fmt.Println(chalk.Magenta.Color("======== MessagePack Code Generator (with ignoreempty) ======="))
	fmt.Printf(chalk.Magenta.Color(">>> Input: \"%s\"\n"), c.GoFile)
	fs, err := parse.File(c)
	if err != nil {
		return err
	}

	if len(fs.Identities) == 0 {
		fmt.Println(chalk.Magenta.Color("No types requiring code generation were found!"))
		return nil
	}

	return printer.PrintFile(newFilename(c.Out, c.GoFile, fs.Package), fs, mode)
}

// picks a new file name based on input flags and input filename(s).
func newFilename(out, old, pkg string) string {
	if out != "" {
		if pre := strings.TrimPrefix(out, old); len(pre) > 0 &&
			!strings.HasSuffix(out, ".go") {
			return filepath.Join(old, out)
		}
		return out
	}

	if fi, err := os.Stat(old); err == nil && fi.IsDir() {
		old = filepath.Join(old, pkg)
	}
	// new file name is old file name + _gen.go
	return strings.TrimSuffix(old, ".go") + "_gen.go"
}
