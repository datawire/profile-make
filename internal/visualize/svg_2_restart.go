package visualize

import (
	"fmt"
	"html/template"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type SVGRestart struct {
	Parent     *SVGMake
	RestartNum uint
	Recipes    []*SVGRecipe
}

func (r *SVGRestart) Title() string {
	dir, err := filepath.Rel(globalProfile.Make.Dir, r.Parent.Dir)
	if err != nil {
		dir = r.Parent.Dir
	}
	return fmt.Sprintf("Make/Restart\n"+
		"Dir: %q\n"+
		"Restart: %d",
		dir,
		r.RestartNum)
}

func (r *SVGRestart) TimeSortedRecipes() []*SVGRecipe {
	if r == nil {
		return nil
	}
	sorted := append([]*SVGRecipe(nil), r.Recipes...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].StartTime().Before(sorted[j].StartTime()) })
	return sorted
}

////////////////////////////////////////////////////////////////////////////////

func (r *SVGRestart) StartTime() time.Time {
	if r == nil || len(r.Recipes) == 0 {
		return time.Time{}
	}
	min := r.Recipes[0].StartTime()
	for _, recipe := range r.Recipes[1:] {
		if recipeStart := recipe.StartTime(); recipeStart.Before(min) {
			min = recipeStart
		}
	}
	return min
}

func (r *SVGRestart) FinishTime() time.Time {
	if r == nil || len(r.Recipes) == 0 {
		return time.Time{}
	}
	max := r.Recipes[0].FinishTime()
	for _, recipe := range r.Recipes[1:] {
		if recipeFinish := recipe.FinishTime(); recipeFinish.After(max) {
			max = recipeFinish
		}
	}
	return max
}

// TODO: This file is where the majority of the "compact" layout logic
// will go.

func (r *SVGRestart) W() XDuration {
	if r == nil {
		return 0
	}
	switch globalLayout {
	case "wallclock":
		return XDuration(r.FinishTime().Sub(r.StartTime()))
	case "compact":
		panic("TODO")
	default:
		panic(errors.Errorf("invalid layout %q", globalLayout))
	}
}

func (r *SVGRestart) H() YLines {
	if r == nil {
		return 0
	}
	switch globalLayout {
	case "wallclock":
		var sum YLines
		for _, recipe := range r.Recipes {
			sum += recipe.H()
		}
		return sum
	case "compact":
		panic("TODO")
	default:
		panic(errors.Errorf("invalid layout %q", globalLayout))
	}
}

var restartTemplateWallclock = template.Must(template.
	New("<x-restart>").
	Funcs(funcMap).
	Parse(`<g class="restart">
		<rect x="{{ .Attrs.X.Percent }}" y="{{ .Attrs.Y.EM }}"
		      width="{{ .Data.W.Percent }}" height="{{ .Data.H.EM }}">
			<title xml:space="preserve">{{ .Data.Title }}</title>
		</rect>
		{{ $yoff := 0 | asYLines }}
		{{ range .Data.TimeSortedRecipes }}
			{{ $xoff := (.StartTime.Sub $.Data.StartTime) | asXDuration }}
			{{ .SVG ($.Attrs.X.Add $xoff) ($.Attrs.Y.Add $yoff) }}
			{{ $yoff = $yoff.Add .H }}
		{{ end }}
	</g>`))

func (r *SVGRestart) SVG(X XDuration, Y YLines) (template.HTML, error) {
	var restartTemplate *template.Template
	switch globalLayout {
	case "wallclock":
		restartTemplate = restartTemplateWallclock
	case "compact":
		panic("TODO")
	default:
		panic(errors.Errorf("invalid layout %q", globalLayout))
	}

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
