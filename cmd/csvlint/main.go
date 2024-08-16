package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"unicode/utf8"

	"github.com/kmatt/csvlint"
)

func printHelpAndExit(code int) {
	flag.PrintDefaults()
	os.Exit(code)
}

func main() {
	delimiter := flag.String("delimiter", ",", "Field delimiter in the file, ex: '\\t' or '|'")
	comment := flag.String("comment", "0", "If not 0, lines beginning with the comment character without preceding whitespace are ignored")
	lazyquotes := flag.Bool("lazyquotes", false, "A quote may appear in an unquoted field and a non-doubled quote may appear in a quoted field")
	help := flag.Bool("help", false, "Print help and exit")
	flag.Parse()

	if *help {
		printHelpAndExit(0)
	}

	if flag.NFlag() > 0 {
		fmt.Fprintln(os.Stderr, "Warning: not using defaults, may not validate CSV to RFC 4180")
	}

	convertedDelimiter, err := strconv.Unquote(`'` + *delimiter + `'`)
	if err != nil {
		fmt.Printf("Error unquoting delimiter '%s', note that only one-character delimiters are supported\n\n", *delimiter)
		printHelpAndExit(1)
	}
	comma, _ := utf8.DecodeRuneInString(convertedDelimiter) // don't need to check size since Unquote returns one-character string

	commentChar, err := strconv.Unquote(`'` + *comment + `'`)
	if err != nil {
		fmt.Printf("Error unquoting comment rune '%s', note that only one-character is supported\n\n", *comment)
		printHelpAndExit(1)
	}
	commentRune, _ := utf8.DecodeRuneInString(commentChar)

	if len(flag.Args()) != 1 {
		fmt.Print("csvlint accepts a single filepath as an argument\n\n")
		printHelpAndExit(1)
	}

	f, err := os.Open(flag.Args()[0])
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("File '%s' does not exist\n", flag.Args()[0])
			os.Exit(1)
		} else {
			panic(err)
		}
	}
	defer f.Close()

	invalids, halted, rc, err := csvlint.Validate(f, comma, commentRune, *lazyquotes)
	if err != nil {
		panic(err)
	}

	if len(invalids) == 0 {
		fmt.Printf("File is valid - %d records", rc)
		os.Exit(0)
	}

	for _, invalid := range invalids {
		fmt.Println(invalid.Error())
	}

	if len(invalids) > 0 {
		fmt.Printf("\n%d errors", len(invalids))
	}

	if halted {
		fmt.Println(" - Halted")
		os.Exit(1)
	}
	os.Exit(2)
}
