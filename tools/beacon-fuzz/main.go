package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/prysmaticlabs/prysm/shared/fileutil"
)

var (
	output = flag.String("output", "", "Output filepath for generated states file.")
)

const tpl = `// Code generated by //tools/beacon-fuzz:beacon-fuzz. DO NOT EDIT.
package {{.Package}}

// generateStates is a map of generated states from ssz.
var generatedStates = map[uint16]string{
	{{.MapStr}}
}
`

// This program generates a map of ID(uint16) -> hex encoded strings of SSZ binary data. This tool
// exists to facilitate running beacon-fuzz targets within the constraints of fuzzit. I.e. fuzz
// only loads the corpus to the file system without the beacon states. An alternative approach would
// be to create a docker image where the state files are available or changing the corpus seed data
// to contain the beacon state itself.
func main() {
	flag.Parse()
	if *output == "" {
		panic("Missing output. Usage: beacon-fuzz --output=out.go path/to/state/0 path/to/state/1 ...")
	}
	statePaths := os.Args[2:]

	if len(statePaths) > 15 {
		statePaths = statePaths[:15]
	}

	m := make(map[int][]byte, len(statePaths))
	for _, p := range statePaths {
		ID, err := strconv.Atoi(filepath.Base(p))
		if err != nil {
			panic(fmt.Sprintf("%s does not end in an integer for the filename.", p))
		}
		b, err := ioutil.ReadFile(p)
		if err != nil {
			panic(err)
		}
		m[ID] = b
	}

	res := execTmpl(tpl, input{Package: "testing", MapStr: sszBytesToMapStr(m)})
	if err := fileutil.WriteFile(*output, res.Bytes()); err != nil {
		panic(err)
	}
}

func sszBytesToMapStr(ss map[int][]byte) string {
	dst := ""
	for i, s := range ss {
		dst += fmt.Sprintf("%d: \"%x\",", i, s)
	}
	return dst
}

type input struct {
	Package string
	States  map[int][]byte
	MapStr  string
}

func execTmpl(tpl string, input interface{}) *bytes.Buffer {
	tmpl, err := template.New("template").Parse(tpl)
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	if err = tmpl.Execute(buf, input); err != nil {
		panic(err)
	}
	return buf
}
