package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	from, to      string
	limit, offset int64
)

func printError(e error) {
	if _, err := fmt.Fprintln(os.Stderr, e.Error()); err != nil {
		panic(err.Error())
	}
}

func init() {
	flag.StringVar(&from, "from", "", "file to read from")
	flag.StringVar(&to, "to", "", "file to write to")
	flag.Int64Var(&limit, "limit", 0, "limit of bytes to copy")
	flag.Int64Var(&offset, "offset", 0, "offset in input file")
}

func main() {
	flag.Parse()
	err := Copy(from, to, offset, limit)
	if err != nil {
		printError(err)
	}
}
