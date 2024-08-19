package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"unicode/utf8"

	"github.com/charmbracelet/log"

	"github.com/kmatt/csvlint"
)

func printHelpAndExit(code int) {
	flag.PrintDefaults()
	os.Exit(code)
}

func main() {
	//TODO Support wildcards in file argument

	delimiter := flag.String("delimiter", ",", "Field delimiter in the file, ex: '\\t' or '|'")
	comment := flag.String("comment", "", "Lines beginning with the comment character without preceding whitespace are ignored")
	lazyquotes := flag.Bool("lazyquotes", false, "A quote may appear in an unquoted field and a non-doubled quote may appear in a quoted field")
	debug := flag.Bool("debug", false, "Print debug information")
	help := flag.Bool("help", false, "Print help and exit")
	flag.Parse()

	if *help {
		printHelpAndExit(0)
	}

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	if flag.NFlag() > 0 {
		log.Warn("Not using defaults, may not validate CSV to RFC 4180")
	}

	convertedDelimiter, err := strconv.Unquote(`'` + *delimiter + `'`)
	if err != nil {
		log.Errorf("Error unquoting delimiter '%s', note that only one-character delimiters are supported\n\n", *delimiter)
		printHelpAndExit(1)
	}
	comma, _ := utf8.DecodeRuneInString(convertedDelimiter) // don't need to check size since Unquote returns one-character string

	commentChar, err := strconv.Unquote(`'` + *comment + `'`)
	if commentChar == "" {
		// https://pkg.go.dev/encoding/csv#Reader
		commentChar = "0"
	}
	if err != nil {
		log.Errorf("Error unquoting comment rune '%s', note that only one-character is supported\n\n", *comment)
		printHelpAndExit(1)
	}
	commentRune, _ := utf8.DecodeRuneInString(commentChar)

	if len(flag.Args()) != 1 {
		log.Error("csvlint accepts a single filepath as an argument\n\n")
		printHelpAndExit(1)
	}

	f, err := os.Open(flag.Args()[0])
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatal("File '%s' does not exist\n", flag.Args()[0])
		} else {
			panic(err)
		}
	}
	defer f.Close()

	log.Debugf("Reading file %s", f.Name())
	invalids, halted, rc, err := csvlint.Validate(f, comma, commentRune, *lazyquotes)
	if err != nil {
		panic(err)
	}

	if len(invalids) == 0 {
		log.Infof("File is valid - %d records", rc)
		os.Exit(0)
	}

	for _, invalid := range invalids {
		fmt.Printf(" %s\n", invalid.Error())
	}

	if len(invalids) > 0 {
		log.Warnf("%d malformed records", len(invalids))
	}

	if halted {
		log.Fatal("Halted")
	}
	os.Exit(2)
}
