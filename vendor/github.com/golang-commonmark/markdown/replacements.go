// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown

import (
	"bytes"

	"github.com/golang-commonmark/markdown/byteutil"
)

var (
	trg       [256]bool
	sym       [256]bool
	exclquest [256]bool
)

func init() {
	for _, b := range "(!+,-.?" {
		trg[b] = true
	}
	for _, b := range "cprtCPRT" {
		sym[b] = true
	}
	for _, b := range "?!" {
		exclquest[b] = true
	}
}

func performReplacements(s string) string {
	var buf bytes.Buffer

	for i := 0; i < len(s); i++ {
		b := s[i]

		if trg[b] {

		outer:
			switch b {
			case '(':
				if i+2 >= len(s) {
					break
				}

				b2 := s[i+1]
				if !sym[b2] {
					break
				}

				b2 = byteutil.ByteToLower(b2)
				switch b2 {
				case 'c', 'r', 'p':
					if s[i+2] != ')' {
						break outer
					}
					switch b2 {
					case 'c':
						buf.WriteString("©")
					case 'r':
						buf.WriteString("®")
					case 'p':
						buf.WriteString("§")
					}
					i += 2
					continue

				case 't':
					if i+3 >= len(s) {
						break outer
					}
					if s[i+3] != ')' || byteutil.ByteToLower(s[i+2]) != 'm' {
						break outer
					}
					buf.WriteString("™")
					i += 3
					continue
				}

			case '+':
				if i+1 >= len(s) || s[i+1] != '-' {
					break
				}
				buf.WriteString("±")
				i++
				continue

			case '.':
				if i+1 >= len(s) || s[i+1] != '.' {
					break
				}

				j := i + 2
				for j < len(s) && s[j] == '.' {
					j++
				}
				if i == 0 || !(s[i-1] == '?' || s[i-1] == '!') {
					buf.WriteString("…")
				} else {
					buf.WriteString("..")
				}
				i = j - 1
				continue

			case '?', '!':
				if i+3 >= len(s) {
					break
				}
				if !(exclquest[s[i+1]] && exclquest[s[i+2]] && exclquest[s[i+3]]) {
					break
				}
				buf.WriteString(s[i : i+3])
				j := i + 3
				for j < len(s) && exclquest[s[j]] {
					j++
				}
				i = j - 1
				continue

			case ',':
				if i+1 >= len(s) || s[i+1] != ',' {
					break
				}
				buf.WriteByte(',')
				j := i + 2
				for j < len(s) && s[j] == ',' {
					j++
				}
				i = j - 1
				continue

			case '-':
				if i+1 >= len(s) || s[i+1] != '-' {
					break
				}
				if i+2 >= len(s) || s[i+2] != '-' {
					buf.WriteString("–")
					i++
					continue
				}
				if i+3 >= len(s) || s[i+3] != '-' {
					buf.WriteString("—")
					i += 2
					continue
				}

				j := i + 3
				for j < len(s) && s[j] == '-' {
					j++
				}
				buf.WriteString(s[i:j])
				i = j - 1
				continue
			}
		}

		buf.WriteByte(b)
	}
	return buf.String()
}

func ruleReplacements(s *StateCore) {
	if !s.Md.Typographer {
		return
	}

	for _, tok := range s.Tokens {
		if tok, ok := tok.(*Inline); ok {
			for _, itok := range tok.Children {
				switch itok := itok.(type) {
				case *Text:
					itok.Content = performReplacements(itok.Content)
				}
			}
		}
	}
}
