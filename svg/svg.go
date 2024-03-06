package svg

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	viewBox  = regexp.MustCompile(`<svg\s*width="[^"]+"\s*height="[^"]+"\s*viewBox="[^"]+"`)
	graphID  = regexp.MustCompile(`<g id="graph\d"`)
	svgClose = regexp.MustCompile(`</svg>`)
)

// massageDotSVG invokes the dot tool to generate an SVG image and alters
// the image to have panning capabilities when viewed in a browser.
func massageDotSVG(input string) ([]byte, error) {
	svgData, err := invokeDot("svg", strings.NewReader(input))
	if err != nil {
		return nil, err
	}

	output := new(bytes.Buffer)
	_, err = output.Write([]byte(massageSVG(svgData.String())))
	if err != nil {
		return nil, err
	}
	return output., nil
}

func invokeDot(format string, input io.Reader) ([]byte, error) {
	output := new(bytes.Buffer)
	cmd := exec.Command("dot", "-T"+format)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = input, output, os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to execute dot. Is Graphviz installed? Error: %v", err)
	}
	return output.Bytes(), nil
}

// massageSVG enhances the SVG output from DOT to provide better
// panning inside a web browser. It uses the svgpan library, which is
// embedded into the svgpan.JSSource variable.
func massageSVG(svg string) string {
	// Work around for dot bug which misses quoting some ampersands,
	// resulting on unparsable SVG.
	svg = strings.Replace(svg, "&;", "&amp;;", -1)

	// Dot's SVG output is
	//
	//    <svg width="___" height="___"
	//     viewBox="___" xmlns=...>
	//    <g id="graph0" transform="...">
	//    ...
	//    </g>
	//    </svg>
	//
	// Change it to
	//
	//    <svg width="100%" height="100%"
	//     xmlns=...>

	//    <script type="text/ecmascript"><![CDATA[` ..$(svgpan.JSSource)... `]]></script>`
	//    <g id="viewport" transform="translate(0,0)">
	//    <g id="graph0" transform="...">
	//    ...
	//    </g>
	//    </g>
	//    </svg>

	if loc := viewBox.FindStringIndex(svg); loc != nil {
		svg = svg[:loc[0]] +
			`<svg width="100%" height="100%"` +
			svg[loc[1]:]
	}

	if loc := graphID.FindStringIndex(svg); loc != nil {
		svg = svg[:loc[0]] +
			`<script type="text/ecmascript"><![CDATA[` + JSSource + `]]></script>` +
			`<g id="viewport" transform="scale(0.5,0.5) translate(0,0)">` +
			svg[loc[0]:]
	}

	if loc := svgClose.FindStringIndex(svg); loc != nil {
		svg = svg[:loc[0]] +
			`</g>` +
			svg[loc[0]:]
	}

	return svg
}
