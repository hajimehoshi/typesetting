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
		{`<html></html>`, font.SVGViewBox{0, 0, upem, upem}},
		{`<svg`, font.SVGViewBox{0, 0, upem, upem}},
		// viewBox attribute
		{`<svg viewBox="0 0 128 128"></svg>`, font.SVGViewBox{0, 0, 128, 128}},
		{`<svg viewBox="-10,-20 30,40"></svg>`, font.SVGViewBox{-10, -20, 30, 40}},
		{`<svg viewBox=" 0 , 0 , 1e2 , 50.5 "></svg>`, font.SVGViewBox{0, 0, 100, 50.5}},
		{`<svg viewBox='0 0 128 64'/>`, font.SVGViewBox{0, 0, 128, 64}},
		{"<svg\tviewBox\t=\r\n\"0 0 128 128\"></svg>", font.SVGViewBox{0, 0, 128, 128}},
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
		// prolog, comments and doctype before the root element
		{"\uFEFF" + `<svg viewBox="0 0 128 128"></svg>`, font.SVGViewBox{0, 0, 128, 128}},
		{`<?xml version="1.0" encoding="UTF-8"?><svg viewBox="0 0 128 128"></svg>`, font.SVGViewBox{0, 0, 128, 128}},
		{`<?xml version="1.0"?>
			<!-- comment with <svg viewBox="0 0 1 1"> inside -->
			<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
			<svg viewBox="0 0 128 128"></svg>`, font.SVGViewBox{0, 0, 128, 128}},
		{`<!DOCTYPE svg [ <!ENTITY foo "bar"> ]><svg viewBox="0 0 128 128"></svg>`, font.SVGViewBox{0, 0, 128, 128}},
		{`<!DOCTYPE svg [ <!ENTITY foo "]>"> ]><svg viewBox="0 0 128 128"></svg>`, font.SVGViewBox{0, 0, 128, 128}},
		{`<!DOCTYPE svg [ <!-- don't mind the ]> here --> ]><svg viewBox="0 0 128 128"></svg>`, font.SVGViewBox{0, 0, 128, 128}},
		// namespace prefix on the root element
		{`<svg:svg xmlns:svg="http://www.w3.org/2000/svg" viewBox="0 0 128 128"></svg:svg>`, font.SVGViewBox{0, 0, 128, 128}},
		// other attributes, possibly with tricky values
		{`<svg xmlns="http://www.w3.org/2000/svg" title="a>b" viewBox="0 0 128 128" enable-background="new"></svg>`, font.SVGViewBox{0, 0, 128, 128}},
	} {
		got := font.SVGDocViewBox([]byte(test.doc), upem)
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
