package visualize

import (
	"html/template"
	"strings"
)

type SVGRecipe struct {
	Name     string
	Commands []*SVGCommand
}

func (recipe *SVGRecipe) H() YLines {
	if recipe == nil {
		return 0
	}
	ret := YLines(0)
	for _, cmd := range recipe.Commands {
		ret += cmd.H()
	}
	return ret
}

func (recipe *SVGRecipe) X() XTime {
	if recipe == nil || len(recipe.Commands) == 0 {
		return XTime{}
	}
	min := recipe.Commands[0].X
	for _, cmd := range recipe.Commands {
		if cmd.X.X.Before(min.X) {
			min = cmd.X
		}
	}
	return min
}

func (recipe *SVGRecipe) x2() XTime {
	if recipe == nil || len(recipe.Commands) == 0 {
		return XTime{}
	}
	max := recipe.Commands[0].X
	for _, cmd := range recipe.Commands {
		if cmd.X.X.After(max.X) {
			max = cmd.X
		}
	}
	return max
}

func (recipe *SVGRecipe) W() XDuration {
	x1 := recipe.X()
	x2 := recipe.x2()
	return XDuration{
		Profile: x1.Profile,
		W:       x2.X.Sub(x1.X),
	}
}

var recipeTemplate = template.Must(template.
	New("<x-recipe>").
	Funcs(funcMap).
	Parse(`<g class="recipe">
		<rect x="{{ .Data.X.Percent }}" y="{{ .Attrs.Y.EM }}"
		      width="{{ .Data.W.Percent }}" height="{{ .Data.H.EM }}" />
		{{ $y := .Attrs.Y }}
		{{ range .Data.Commands }}{{ .SVG $.Attrs.X $y }}{{ $y = addLines $y .H }}{{ end }}
	</g>`))

func (recipe *SVGRecipe) SVG(X XTime, Y YLines) (template.HTML, error) {
	var str strings.Builder
	err := recipeTemplate.Execute(&str, map[string]interface{}{
		"Attrs": map[string]interface{}{
			"X": X,
			"Y": Y,
		},
		"Data": recipe,
	})
	if err != nil {
		return "", err
	}
	return template.HTML(str.String()), nil
}
