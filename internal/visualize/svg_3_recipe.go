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

func (recipe *SVGRecipe) Dependencies() []string {
	set := make(map[string]struct{})
	for _, cmd := range recipe.Commands {
		for _, dep := range cmd.Raw.RecipeDependencies {
			set[dep] = struct{}{}
		}
	}
	ret := make([]string, 0, len(set))
	for dep := range set {
		ret = append(ret, dep)
	}
	return ret
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
	var max XDuration
	for _, cmd := range recipe.Commands {
		if w := cmd.W(); w > max {
			max = w
		}
	}
	return max
}

func (recipe *SVGRecipe) H() YLines {
	if recipe == nil {
		return 0
	}
	if globalVerboseCommand {
		var sum YLines
		for _, cmd := range recipe.Commands {
			sum += cmd.H()
		}
		return sum
	} else {
		var xCursor time.Time
		var yCursor YLines
		var yPending YLines
		for _, cmd := range recipe.SortedCommands() {
			if cmd.StartTime().After(xCursor) {
				// happy path
				if h := cmd.H(); h > yPending {
					yPending = h
				}
			} else {
				// sad path
				yCursor += yPending
				yPending = cmd.H()
			}
		}
		return yCursor + yPending
	}
}

var recipeTemplate = template.Must(template.
	New("<x-recipe>").
	Funcs(funcMap).
	Parse(`<svg class="recipe"
		    x="{{ .Attrs.X.PercentOf .Data.Parent.W }}" y="{{ .Attrs.Y.EM }}"
		    width="{{ .Data.W.PercentOf .Data.Parent.W }}" height="{{ .Data.H.EM }}">
		<title xml:space="preserve">{{ .Data.Title }}</title>
		<rect class="background" x="0" y="0" width="100%" height="100%" />
		{{ if verboseCommand }}
			{{ $yoff := asYLines 0 }}
			{{ range .Data.SortedCommands }}
				{{ $xoff := (.StartTime.Sub $.Data.StartTime) | asXDuration }}
				{{ .SVG $xoff $yoff }}
				{{ $yoff = $yoff.Add .H }}
			{{ end }}
		{{ else }}
			{{ $xCursor := asXDuration 0 }}
			{{ $yCursor := asYLines 0 }}
			{{ $yPending := asYLines 0 }}
			{{ range .Data.SortedCommands }}
				{{ $xoff := (.StartTime.Sub $.Data.StartTime) | asXDuration }}
				{{ if gt $xoff $xCursor }}
					{{ if gt .H $yPending }}
						{{ $yPending = .H }}
					{{ end }}
				{{ else }}
					{{ $yCursor = $yCursor.Add $yPending }}
					{{ $yPending = .H }}
				{{ end }}
				{{ .SVG $xoff $yCursor }}
			{{ end }}
		{{ end }}
	</svg>`))

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
