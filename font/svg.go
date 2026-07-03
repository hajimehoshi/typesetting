// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package font

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-text/typesetting/font/opentype/tables"
)

type svg []svgDocument

func newSvg(table tables.SVG) (svg, error) {
	rawData := table.SVGDocumentList.SVGRawData
	out := make(svg, len(table.SVGDocumentList.DocumentRecords))
	for i, rec := range table.SVGDocumentList.DocumentRecords {
		start, end := rec.SvgDocOffset, rec.SvgDocOffset+tables.Offset32(rec.SvgDocLength)
		if len(rawData) < int(end) {
			return nil, fmt.Errorf("invalid svg table (EOF: expected %d, got %d)", end, len(rawData))
		}
		out[i] = svgDocument{
			first: rec.StartGlyphID,
			last:  rec.EndGlyphID,
			svg:   rawData[start:end],
		}
	}
	return out, nil
}

type svgDocument struct {
	// svg document
	// each glyph description must be written
	// in an element with id=glyphXXX
	svg   []byte
	first gID // The first glyph ID in the range described by this index entry.
	last  gID // The last glyph ID in the range described by this index entry. Must be >= startGlyphID.
}

// rawGlyphData returns the SVG document for [gid], or false.
func (s svg) rawGlyphData(gid gID) ([]byte, bool) {
	// binary search
	for i, j := 0, len(s); i < j; {
		h := i + (j-i)/2
		entry := s[h]
		if gid < entry.first {
			j = h
		} else if entry.last < gid {
			i = h + 1
		} else {
			return entry.svg, true
		}
	}
	return nil, false
}

// SVGViewBox is a rectangle in the "user" coordinate space
// of an SVG document.
type SVGViewBox struct {
	MinX, MinY, Width, Height float32
}

// svgViewBox returns the initial viewport of the SVG document [doc],
// resolved from the root <svg> element attributes, defaulting to the
// em square given by [upem], as required by the specification.
func svgViewBox(doc []byte, upem uint16) SVGViewBox {
	viewBox, width, height, ok := svgRootAttributes(doc)
	if ok {
		if vb, ok := parseSVGViewBox(viewBox); ok {
			return vb
		}
		if w, okW := parseSVGLength(width); okW {
			if h, okH := parseSVGLength(height); okH {
				return SVGViewBox{0, 0, w, h}
			}
		}
	}
	em := float32(upem)
	return SVGViewBox{0, 0, em, em}
}

// parseSVGViewBox parses the value of a viewBox attribute:
// four numbers separated by whitespace and/or commas.
func parseSVGViewBox(s string) (SVGViewBox, bool) {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\r' || r == '\n'
	})
	if len(fields) != 4 {
		return SVGViewBox{}, false
	}
	var nums [4]float32
	for i, field := range fields {
		v, err := strconv.ParseFloat(field, 32)
		if err != nil {
			return SVGViewBox{}, false
		}
		nums[i] = float32(v)
	}
	if nums[2] <= 0 || nums[3] <= 0 {
		return SVGViewBox{}, false
	}
	return SVGViewBox{nums[0], nums[1], nums[2], nums[3]}, true
}

// parseSVGLength parses a positive width or height attribute value
// expressed in user units, such as "12" or "12px".
func parseSVGLength(s string) (float32, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "px")
	v, err := strconv.ParseFloat(s, 32)
	if err != nil || v <= 0 {
		return 0, false
	}
	return float32(v), true
}

func isXMLSpace(c byte) bool { return c == ' ' || c == '\t' || c == '\r' || c == '\n' }

// svgRootAttributes scans the prolog of the XML document [doc] and returns
// the raw viewBox, width and height attribute values of the root <svg> element.
//
// A dedicated scanner is used instead of [encoding/xml], which would add
// a significant binary size overhead to this core package for the simple
// task of reading the root element attributes.
func svgRootAttributes(doc []byte) (viewBox, width, height string, ok bool) {
	// skip an optional byte order mark
	doc = bytes.TrimPrefix(doc, []byte{0xEF, 0xBB, 0xBF})
	for len(doc) != 0 {
		if isXMLSpace(doc[0]) {
			doc = doc[1:]
			continue
		}
		switch {
		case bytes.HasPrefix(doc, []byte("<?")): // XML declaration or processing instruction
			end := bytes.Index(doc, []byte("?>"))
			if end == -1 {
				return "", "", "", false
			}
			doc = doc[end+2:]
		case bytes.HasPrefix(doc, []byte("<!--")): // comment
			end := bytes.Index(doc, []byte("-->"))
			if end == -1 {
				return "", "", "", false
			}
			doc = doc[end+3:]
		case bytes.HasPrefix(doc, []byte("<!")): // doctype declaration
			end := svgDoctypeEnd(doc)
			if end == -1 {
				return "", "", "", false
			}
			doc = doc[end:]
		case doc[0] == '<': // root start tag
			return parseSVGStartTag(doc)
		default:
			return "", "", "", false
		}
	}
	return "", "", "", false
}

// svgDoctypeEnd returns the index just after the closing '>' of the
// doctype declaration starting [doc], or -1.
func svgDoctypeEnd(doc []byte) int {
	depth := 0 // nesting in the internal subset brackets
	for i := 2; i < len(doc); i++ {
		switch doc[i] {
		case '[':
			depth++
		case ']':
			depth--
		case '>':
			if depth <= 0 {
				return i + 1
			}
		}
	}
	return -1
}

// parseSVGStartTag parses the start tag beginning [doc] and returns
// the raw viewBox, width and height attribute values, or false if the
// element is not an <svg> element or is malformed.
func parseSVGStartTag(doc []byte) (viewBox, width, height string, ok bool) {
	doc = doc[1:] // '<'
	i := 0
	for i < len(doc) && !isXMLSpace(doc[i]) && doc[i] != '>' && doc[i] != '/' {
		i++
	}
	name := doc[:i]
	if j := bytes.LastIndexByte(name, ':'); j != -1 { // namespace prefix
		name = name[j+1:]
	}
	// The element is matched by its local name only: the namespace binding
	// is not verified, since conforming fonts always use the SVG namespace,
	// and an svg element bound to another namespace is too unlikely to be
	// worth rejecting.
	if string(name) != "svg" {
		return "", "", "", false
	}
	doc = doc[i:]
	for {
		for len(doc) != 0 && isXMLSpace(doc[0]) {
			doc = doc[1:]
		}
		if len(doc) == 0 {
			return "", "", "", false
		}
		if doc[0] == '>' || doc[0] == '/' { // end of the start tag
			return viewBox, width, height, true
		}
		// attribute name
		i = 0
		for i < len(doc) && doc[i] != '=' && !isXMLSpace(doc[i]) && doc[i] != '>' && doc[i] != '/' {
			i++
		}
		attr := doc[:i]
		doc = doc[i:]
		for len(doc) != 0 && isXMLSpace(doc[0]) {
			doc = doc[1:]
		}
		if len(doc) == 0 || doc[0] != '=' {
			return "", "", "", false
		}
		doc = doc[1:]
		for len(doc) != 0 && isXMLSpace(doc[0]) {
			doc = doc[1:]
		}
		if len(doc) == 0 || (doc[0] != '"' && doc[0] != '\'') {
			return "", "", "", false
		}
		quote := doc[0]
		doc = doc[1:]
		end := bytes.IndexByte(doc, quote)
		if end == -1 {
			return "", "", "", false
		}
		value := string(doc[:end])
		doc = doc[end+1:]
		switch string(attr) {
		case "viewBox":
			viewBox = value
		case "width":
			width = value
		case "height":
			height = value
		}
	}
}
