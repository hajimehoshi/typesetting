// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package font

// SVGViewBoxForTest exposes [svgViewBox] for tests.
func SVGViewBoxForTest(doc []byte, upem uint16) SVGViewBox { return svgViewBox(doc, upem) }
