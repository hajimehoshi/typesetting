// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package font_test

import (
	"bytes"
	"testing"

	td "github.com/go-text/typesetting-utils/opentype"
	"github.com/go-text/typesetting/font"
	tu "github.com/go-text/typesetting/testutils"
)

func TestSVGViewBox(t *testing.T) {
	const upem = 1000
	for _, test := range []struct {
		doc      string
		expected font.SVGViewBox
	}{
		// no viewport information: default to the em square
		{`<svg xmlns="http://www.w3.org/2000/svg" id="glyph1"></svg>`, font.SVGViewBox{0, 0, upem, upem}},
		{``, font.SVGViewBox{0, 0, upem, upem}},
		{`not xml`, font.SVGViewBox{0, 0, upem, upem}},
		{`<svg`, font.SVGViewBox{0, 0, upem, upem}}, // malformed
		// the root element must be an svg element, in the SVG namespace or none
		{`<html viewBox="0 0 128 128"></html>`, font.SVGViewBox{0, 0, upem, upem}},
		{`<svg xmlns="http://example.com/wrong" viewBox="0 0 128 128"></svg>`, font.SVGViewBox{0, 0, upem, upem}},
		{`<w:svg xmlns:w="http://example.com/wrong" viewBox="0 0 128 128"></w:svg>`, font.SVGViewBox{0, 0, upem, upem}},
		{`<svg:svg xmlns:svg="http://www.w3.org/2000/svg" viewBox="0 0 128 128"></svg:svg>`, font.SVGViewBox{0, 0, 128, 128}},
		// viewBox attribute (a missing namespace is tolerated)
		{`<svg viewBox="0 0 128 128"></svg>`, font.SVGViewBox{0, 0, 128, 128}},
		{`<svg viewBox="-10,-20 30,40"></svg>`, font.SVGViewBox{-10, -20, 30, 40}},
		{`<svg viewBox=" 0 , 0 , 1e2 , 50.5 "></svg>`, font.SVGViewBox{0, 0, 100, 50.5}},
		{`<svg viewBox='0 0 128 64'/>`, font.SVGViewBox{0, 0, 128, 64}},
		// invalid viewBox attributes
		{`<svg viewBox="0 0 128"></svg>`, font.SVGViewBox{0, 0, upem, upem}},
		{`<svg viewBox="0 0 0 128"></svg>`, font.SVGViewBox{0, 0, upem, upem}},
		{`<svg viewBox="0 0 128 -1"></svg>`, font.SVGViewBox{0, 0, upem, upem}},
		{`<svg viewBox="a b c d"></svg>`, font.SVGViewBox{0, 0, upem, upem}},
		{`<svg viewBox="NaN 0 128 128"></svg>`, font.SVGViewBox{0, 0, upem, upem}},
		{`<svg viewBox="0 0 Inf 128"></svg>`, font.SVGViewBox{0, 0, upem, upem}},
		{`<svg viewBox="0 0 128 +infinity"></svg>`, font.SVGViewBox{0, 0, upem, upem}},
		// width and height attributes
		{`<svg width="100" height="200"></svg>`, font.SVGViewBox{0, 0, 100, 200}},
		{`<svg width="100px" height="50px"></svg>`, font.SVGViewBox{0, 0, 100, 50}},
		{`<svg width="100"></svg>`, font.SVGViewBox{0, 0, upem, upem}},
		{`<svg width="100%" height="100%"></svg>`, font.SVGViewBox{0, 0, upem, upem}},
		{`<svg width="12pt" height="12pt"></svg>`, font.SVGViewBox{0, 0, upem, upem}},
		{`<svg width="NaN" height="NaN"></svg>`, font.SVGViewBox{0, 0, upem, upem}},
		{`<svg width="Inf" height="Inf"></svg>`, font.SVGViewBox{0, 0, upem, upem}},
		// viewBox has precedence over width and height
		{`<svg width="100" height="200" viewBox="0 0 128 128"></svg>`, font.SVGViewBox{0, 0, 128, 128}},
		{`<svg viewBox="invalid" width="100" height="200"></svg>`, font.SVGViewBox{0, 0, 100, 200}},
		// content before the root element is skipped
		{`<?xml version="1.0"?><!-- comment --><!DOCTYPE svg><svg viewBox="0 0 128 128"></svg>`, font.SVGViewBox{0, 0, 128, 128}},
	} {
		got := font.SVGViewBoxForTest([]byte(test.doc), upem)
		if got != test.expected {
			t.Fatalf("document %q: expected %v, got %v", test.doc, test.expected, got)
		}
	}
}

func TestGlyphDataSVGViewBox(t *testing.T) {
	// this font has no viewBox attribute in its SVG document:
	// the viewport defaults to the em square
	data, err := td.Files.ReadFile("toys/chromacheck-svg.ttf")
	tu.AssertNoErr(t, err)

	face, err := font.ParseTTF(bytes.NewReader(data))
	tu.AssertNoErr(t, err)

	glyph, ok := face.GlyphDataSVG(1)
	tu.Assert(t, ok)
	tu.Assert(t, glyph.ViewBox == font.SVGViewBox{0, 0, 1024, 1024})
}
