// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package font

// SVGDocViewBox exposes [svgViewBox] for tests.
func SVGDocViewBox(doc []byte, upem uint16) SVGViewBox { return svgViewBox(doc, upem) }
