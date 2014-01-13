package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

type editOperation int
const (
	ins = iota
	del
	sub
	match
)

type editScript []editOperation

type characterTag int
const (
	tagO = iota
	tagI
	tagT
	tagS
)

var levOptions levenshtein.Options = levenshtein.Options {
	InsCost: 1,
	DelCost: 1,
	SubCost: 2,
	Matches: func (source rune, target rune) bool {
		if ((source == ' ' || source == '\n') && unicode.IsSpace(target)) {
			return true
		}
		return source == target
	},
}

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("usage: %s <rawfile> <tokfile>\n",
				filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	rawName := os.Args[1]
	tokName := os.Args[2]
	var rawFile *os.File
	var tokFile *os.File
	var err error
	if rawFile, err = os.Open(rawName); err != nil {
		log.Println("failed to open raw file: ", err)
		os.Exit(1)
	}
	defer rawFile.Close()
	if tokFile, err = os.Open(tokName); err != nil {
		log.Println("failed to open tok file: ", err)
		os.Exit(1)
	}
	defer tokFile.Close()
	rawReader := bufio.NewReader(rawFile)
	tokReader := bufio.NewReader(tokFile)
	for { // while read line from rawReader
		rawLine, err := rawReader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println("error reading raw file: ", err)
			os.Exit(1)
		} else if !utf8.ValidString(rawLine) {
			log.Println("invalid UTF-8 in raw file")
			os.Exit(1)
		} else {
			tokLine, err := tokReader.ReadString('\n')
			if err == io.EOF {
				log.Println("unexpected end of tok file")
				os.Exit(1)
			} else if err != nil {
				log.Println("error reading raw file: ", err)
				os.Exit(1)
			} else if !utf8.ValidString(tokLine) {
				log.Println("invalid UTF-8 in tok file")
				os.Exit(1)
			} else {
				rawLine = strings.TrimRight(rawLine, "\n")
				tokLine = strings.TrimRight(tokLine, "\n")
				processLines(rawLine, tokLine)
			}
		}
	}
	_, err = tokReader.ReadBytes('\n')
	if err != io.EOF {
		if err == nil {
			log.Println("unexpected end of raw file")
		} else {
			log.Println("error reading tok file: ", err)
		}
		os.Exit(1)
	}
}

func processLines(rawLine string, tokLine string) {
	rawText := []rune(strings.Replace(rawLine, "<NEWLINE>", "\n", -1))
	tokText := []rune(strings.Replace(tokLine, "<SENT>", "\n", -1))
	characterTags := iobify(rawText, tokText)
	for i, tag := range characterTags {
		fmt.Printf("%d %s\n", rawText[i], tagToString(tag))
	}
	fmt.Println() // separate articles by newlines
}

func iobify(raw []rune, tok []rune) []characterTag {
	tokTags := tagTokenizedText(tok)
	alignment := align(raw, tok)
	result := make([]characterTag, 0, len(raw))
	oldI := -1 // contains previous aligned tok index
	var tag characterTag
	for _, i := range alignment { // i is char index in tok
		if i == -1 {
			// unaligned
			tag = tagO
		} else if oldI == i {
			// continuation of a stretch of raw characters
			// aligned to the same tok character
			tag = tagI
		} else {
			// start of a new stretch
			tag = tokTags[i]
		}
		result = append(result, tag)
		oldI = i
	}
	return result
}

func align(raw []rune, tok []rune) []int {
	script := levenshtein.EditScriptForStrings(tok, raw, levOptions)
	// For each raw char, alignment slice indicates index of aligned tok
	// tok char, or -1 for unaligned.
	alignment := make([]int, 0, len(raw))
	i := 0 // tok offset
	j := 0 // raw offset
	for _, op := range script {
		if op == levenshtein.Match || op == levenshtein.Sub {
			alignment = append(alignment, i)
			i++
			j++
		} else if op == levenshtein.Ins {
			if !unicode.IsSpace(raw[j]) && j > 0 {
				// align to left (or mark as unaligned if left
				// is unaligned)
				alignment = append(alignment, alignment[j - 1])
			} else {
				// mark as unaligned
				alignment = append(alignment, -1)
			}
			j++
		} else if op == levenshtein.Del {
			i++
		}
	}
	// add alignments to right
	for j = len(raw) - 2; j >= 0; j-- {
		if alignment[j] == -1 && !unicode.IsSpace(raw[j]) {
			alignment[j] = alignment[j + 1]
		}
	}
	return alignment
}

func tagTokenizedText(tok []rune) []characterTag {
	result := make([]characterTag, 0, len(tok))
	inSentence := false
	inToken := false
	var tag characterTag
	for _, char := range tok {
		if char == ' ' {
			tag = tagO
			inToken = false
		} else if char == '\n' {
			tag = tagO
			inToken = false
			inSentence = false
		} else {
			if inToken {
				tag = tagI
			} else if inSentence {
				tag = tagT
				inToken = true
			} else {
				tag = tagS
				inToken = true
				inSentence = true
			}
		}
		result = append(result, tag)
	}
	return result
}

func tagToString(tag characterTag) string {
	switch(tag) {
	case tagO:
		return "O"
	case tagI:
		return "I"
	case tagT:
		return "T"
	case tagS:
		return "S"
	}
	return ""
}

