package widgets

import (
	"strings"
	"unicode/utf8"

	"github.com/awesome-gocui/gocui"
	"github.com/mattn/go-runewidth"
)

func truncateANSIString(input string, maxLen int) (string, int) {
	return newANSI(input).Truncate(maxLen)
}

type ansiBlob []ansiSegment

func (a ansiBlob) Truncate(maxLen int) (string, int) {
	var output string
	var total int
	for _, segment := range a {
		value := segment.value
		if runewidth.StringWidth(segment.value) > maxLen {
			value = value[:maxLen]
		}
		output += segment.style + value
		total += runewidth.StringWidth(value)
		maxLen -= runewidth.StringWidth(value)
		if maxLen <= 0 {
			break
		}
	}
	return output, total
}

func (a ansiBlob) Strip() string {
	var output string
	for _, segment := range a {
		output += segment.value
	}
	return output
}

func (a ansiBlob) TrimSpace() ansiBlob {
	return newANSI(strings.TrimSpace(a.String()))
}

func (a ansiBlob) Len() int {
	var c int
	for _, segment := range a {
		c += runewidth.StringWidth(segment.value)
	}
	return c
}

func (a ansiBlob) String() string {
	var output string
	for _, segment := range a {
		output += segment.style + segment.value
	}
	return output
}

func (a ansiBlob) ANSI() string {
	var output string
	for _, segment := range a {
		output += segment.style
	}
	return simplifyANSI(output)
}

func simplifyANSI(input string) string {
	parts := strings.Split(input, "\x1b[0m")
	return parts[len(parts)-1]
}

func (a ansiBlob) Cut(index int) (ansiBlob, ansiBlob) {
	var current int
	var outputBefore string
	var outputAfter string
	var found bool
	for _, segment := range a {
		if found {
			outputAfter += segment.style + segment.value
			continue
		}
		if index < current+utf8.RuneCountInString(segment.value) {
			localIndex := index - current
			outputBefore += string([]rune(segment.value)[:localIndex])
			outputAfter = segment.style + string([]rune(segment.value)[localIndex:])
			found = true
			continue
		}
		outputBefore += segment.style + segment.value
		current += utf8.RuneCountInString(segment.value)
	}
	return newANSI(outputBefore), newANSI(outputAfter)
}

func (a ansiBlob) Words() []ansiBlob {
	var output []ansiBlob // nolint:prealloc
	words := strings.Split(a.String(), " ")
	var ansi string
	for _, word := range words {
		w := newANSI(word).TrimSpace()
		ansi = simplifyANSI(ansi + w.ANSI())
		if w.Len() == 0 {
			continue
		}
		output = append(output, w)
	}
	return output
}

type ansiSegment struct {
	value string
	style string
}

func newANSI(input string) ansiBlob {
	var output []ansiSegment
	var current ansiSegment
	inCSI := false
	prev := rune(0)
	for _, r := range input {
		if inCSI { //nolint:gocritic
			current.style += string(r)
			if r >= 0x40 && r <= 0x7E {
				inCSI = false
			}
		} else if r == '[' && prev == 0x1b {
			current.value = current.value[:utf8.RuneCountInString(current.value)-1]
			if current.value != "" {
				output = append(output, current)
				current = ansiSegment{}
			}
			inCSI = true
			current.style += "\x1b["
		} else {
			current.value += string(r)
		}
		prev = r
	}
	if current.value != "" || current.style != "" {
		output = append(output, current)
	}
	return output
}

func exitModal(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView(v.Name()); err != nil {
		return err
	}
	_, err := g.SetCurrentView(Results)
	return err
}
