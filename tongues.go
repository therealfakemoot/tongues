package main

import (
	"flag"
	"fmt"
	m "github.com/therealfakemoot/gomarkov"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Normalize takes an input string and returns a slice of whitespace spearated words in all lowercase with all punctuation removed.
func Normalize(s string) []string {
	// This regex is neat. \p{L} means "any letter in any language". \p{Z} means "any whitespace character in any unicode language". I'm using these so the markov engine can be 100% unicode friendly and language agnostic.
	reg, _ := regexp.Compile(`[^\p{L}\p{Z}]+`)
	words := strings.Split(strings.ToLower(reg.ReplaceAllString(s, "")), " ")
	return words

}

// W builds a closure that fits the WalkFunc signature so you can recursively load corpus files.
func W(c *m.Chain) filepath.WalkFunc {
	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if err != nil {
			fmt.Printf("Unable to walk %s.\n", path)
			return err
		}

		// fmt.Printf("Loading: %s\n", path)
		raw, err := LoadFile(path)
		if err != nil {
			fmt.Printf("Unable to load %s.\n", path)
			return err
		}
		c.Add(Normalize(raw))
		return nil
	}
	return wf
}

// LoadFile returns the contents of a file as a raw string.
func LoadFile(fn string) (string, error) {
	b, err := ioutil.ReadFile(fn)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func wiggle(low, high int) int {
	return rand.Intn(high) + low
}

func Text(c *m.Chain, x, y int) string {

	tokens := m.NGram{m.StartToken}

	for i := 0; i < wiggle(x, y); i++ {
		lastIndex := len(tokens) - 1
		var seed m.NGram
		seed = append(seed, tokens[lastIndex:lastIndex+c.Order]...)
		/*
			for _, t := range tokens[lastIndex : lastIndex+c.Order] {
				seed = append(seed, t)
			}
		*/
		n, err := c.Generate(seed)
		if err != nil {
			fmt.Printf("%s", err)
		}
		if n == " " {
			i--
		}
		tokens = append(tokens, n)
	}
	return strings.Join(tokens[1:len(tokens)-1], " ")
}

func main() {
	var order int

	var min, max int
	var dir string

	flag.IntVar(&order, "order", 1, "Ordinality of Markov chains.")
	flag.IntVar(&min, "min", 10, "Minimum output word count.")
	flag.IntVar(&max, "max", 30, "Maximum output word count.")
	flag.StringVar(&dir, "dir", "corpus/", "Directory containing corpus texts to ingest.")

	flag.Parse()

	c := m.NewChain(1)
	w := W(c)
	p, err := filepath.Abs(dir)
	if err != nil {
		fmt.Printf("%s", err)
	}
	err = filepath.Walk(p, w)
	if err != nil {
		fmt.Printf("%s", err)
	}

	fmt.Printf("%s\n", Text(c, min, max))

}
