package visualize

import (
	"html/template"
	"strings"
)

type SVGRestart map[string]*SVGRecipe

func (r *SVGRestart) H() YLines {
	if r == nil || *r == nil {
		return 0
	}
	var ret YLines
	for _, recipe := range *r {
		ret += recipe.H()
	}
	return ret
}

var restartTemplate = template.Must(template.
	New("<x-restart>").
	Funcs(funcMap).
	Parse(`<g class="restart">
		{{ $y := .Attrs.Y }}
		{{ range .Data }}{{ .SVG $.Attrs.X $y }}{{ $y = addLines $y .H }}{{ end }}
	</g>`))

func (r *SVGRestart) SVG(X XTime, Y YLines) (template.HTML, error) {
	var str strings.Builder
	err := restartTemplate.Execute(&str, map[string]interface{}{
		"Attrs": map[string]interface{}{
			"X": X,
			"Y": Y,
		},
		"Data": r,
	})
	if err != nil {
		return "", err
	}
	return template.HTML(str.String()), nil
}
