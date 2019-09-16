package visualize

import (
	"fmt"
	"html/template"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type SVGRecipe struct {
	Parent   *SVGRestart
	Name     string
	Commands []*SVGCommand
}

func (recipe *SVGRecipe) Title() string {
	target, err := filepath.Rel(globalProfile.Make.Dir, recipe.Name)
	if err != nil {
		target = recipe.Name
	}
	return fmt.Sprintf("Make/Restart/Recipe\n"+
		"Target: %q\n"+
		"Duration: %s",
		target,
		recipe.FinishTime().Sub(recipe.StartTime()))
}

func (recipe *SVGRecipe) SortedCommands() []*SVGCommand {
	if recipe == nil {
		return nil
	}
	sorted := append([]*SVGCommand(nil), recipe.Commands...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].StartTime().Before(sorted[j].StartTime()) })
	return sorted
}

////////////////////////////////////////////////////////////////////////////////

func (recipe *SVGRecipe) StartTime() time.Time {
	if recipe == nil || len(recipe.Commands) == 0 {
		return time.Time{}
	}
	min := recipe.Commands[0].StartTime()
	for _, command := range recipe.Commands[1:] {
		if commandStart := command.StartTime(); commandStart.Before(min) {
			min = commandStart
		}
	}
	return min
}

func (recipe *SVGRecipe) FinishTime() time.Time {
	if recipe == nil || len(recipe.Commands) == 0 {
		return time.Time{}
	}
	max := recipe.Commands[0].FinishTime()
	for _, command := range recipe.Commands[1:] {
		if commandFinish := command.FinishTime(); commandFinish.After(max) {
			max = commandFinish
		}
	}
	return max
}

func (recipe *SVGRecipe) W() XDuration {
	if recipe == nil {
		return 0
	}
	return XDuration(recipe.FinishTime().Sub(recipe.StartTime()))
}

func (recipe *SVGRecipe) H() YLines {
	if recipe == nil {
		return 0
	}
	var sum YLines
	for _, cmd := range recipe.Commands {
		sum += cmd.H()
	}
	return sum
}

var recipeTemplate = template.Must(template.
	New("<x-recipe>").
	Funcs(funcMap).
	Parse(`<g class="recipe">
		<rect x="{{ .Attrs.X.PercentOf .Data.Parent.Parent.ParentW }}" y="{{ .Attrs.Y.EM }}"
		      width="{{ .Data.W.PercentOf .Data.Parent.Parent.ParentW }}" height="{{ .Data.H.EM }}">
			<title xml:space="preserve">{{ .Data.Title }}</title>
		</rect>
		{{ $yoff := 0 | asYLines }}
		{{ range .Data.SortedCommands }}
			{{ $xoff := (.StartTime.Sub $.Data.StartTime) | asXDuration }}
			{{ .SVG ($.Attrs.X.Add $xoff) ($.Attrs.Y.Add $yoff) }}
			{{ $yoff = $yoff.Add .H }}
		{{ end }}
	</g>`))

func (recipe *SVGRecipe) SVG(X XDuration, Y YLines) (template.HTML, error) {
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
