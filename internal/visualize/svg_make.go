package visualize

import (
	"html/template"
	"strings"
)

type SVGMake []*SVGRestart

func (m *SVGMake) H() YLines {
	if m == nil {
		return 0
	}
	var ret YLines
	for _, restart := range *m {
		ret += restart.H()
	}
	return ret
}

var makeTemplate = template.Must(template.
	New("<x-make>").
	Funcs(funcMap).
	Parse(`<g class="make">
		{{ $y := .Attrs.Y }}
		{{ range .Data }}{{ .SVG $.Attrs.X $y }}{{ $y = addLines $y .H }}{{ end }}
	</g>`))

func (m *SVGMake) SVG(X XTime, Y YLines) (template.HTML, error) {
	var str strings.Builder
	err := makeTemplate.Execute(&str, map[string]interface{}{
		"Attrs": map[string]interface{}{
			"X": X,
			"Y": Y,
		},
		"Data": m,
	})
	if err != nil {
		return "", err
	}
	return template.HTML(str.String()), nil
}
