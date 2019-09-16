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

	layout *Layout
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

func (r *SVGRestart) Layout() *Layout {
	if r.layout == nil {
		r.layout = new(Layout)
		r.layout.AddRecipes(r.TimeSortedRecipes())
	}
	return r.layout
}

////////////////////////////////////////////////////////////////////////////////

type Layout struct {
	recipes map[string]*SVGRecipe

	xPositions map[*SVGRecipe]XDuration

	rows       []XDuration
	yPositions map[*SVGRecipe]YLines
}

func (l *Layout) AddRecipes(recipes []*SVGRecipe) {
	// establish name-to-struct mapping
	l.recipes = make(map[string]*SVGRecipe, len(recipes))
	for _, recipe := range recipes {
		l.recipes[recipe.Name] = recipe
		// TODO: Somehow also get recipe.AlsoMakes, not just recipe.Name
	}
	// establish struct-to-X mapping
	l.xPositions = make(map[*SVGRecipe]XDuration, len(recipes))
	for _, recipe := range l.recipes {
		l.solveX(recipe)
	}
	// establish struct-to-Y mapping
	//
	// XXX: using arg-recipes so that there's hope of a stable
	// Y-ordering (also depends on being called with
	// .TimeSortedRecipes() instead of .Recipes), until I figure out
	// a real Y-ordering algorithm.
	l.yPositions = make(map[*SVGRecipe]YLines, len(recipes))
	for _, recipe := range recipes {
		x := l.X(recipe)
		w := recipe.W()
		h := recipe.H()
		var y YLines
		for !l.rectAvailable(x, y, w, h) {
			y++
		}
		l.rectAdd(x, y, w, h)
		l.yPositions[recipe] = y
	}
}

func (l *Layout) solveX(recipe *SVGRecipe) XDuration {
	if _, solved := l.xPositions[recipe]; !solved {
		var max XDuration
		depNames := recipe.Dependencies()
		if recipe.Name != "" {
			// include "" (parse-time commands) as a pseudo-dependency
			depNames = append(depNames, "")
		}
		for _, depName := range depNames {
			if depRecipe, depRecipeOK := l.recipes[depName]; depRecipeOK {
				depOffset := l.solveX(depRecipe) + depRecipe.W()
				if depOffset > max {
					max = depOffset
				}
			}
		}
		l.xPositions[recipe] = max
	}
	return l.xPositions[recipe]
}

func (l *Layout) rectAdd(x XDuration, y YLines, w XDuration, h YLines) {
	for YLines(len(l.rows)) < y+h {
		l.rows = append(l.rows, 0)
	}
	for iy := y; iy < y+h; iy++ {
		if l.rows[iy] < x+w {
			l.rows[iy] = x + w
		}
	}
}

func (l Layout) rectAvailable(x XDuration, y YLines, w XDuration, h YLines) bool {
	for iy := y; iy < y+h; iy++ {
		if iy < YLines(len(l.rows)) && l.rows[iy] > x {
			return false
		}
	}
	return true
}

func (l Layout) X(recipe *SVGRecipe) XDuration {
	x, ok := l.xPositions[recipe]
	if !ok {
		panic("non-layed-out recipe")
	}
	return x
}

func (l Layout) Y(recipe *SVGRecipe) YLines {
	y, ok := l.yPositions[recipe]
	if !ok {
		panic("non-layed-out recipe")
	}
	return y
}

func (l Layout) W() XDuration {
	var max XDuration
	for _, w := range l.rows {
		if w > max {
			max = w
		}
	}
	return max
}

func (l Layout) H() YLines {
	return YLines(len(l.rows))
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

func (r *SVGRestart) W() XDuration {
	if r == nil {
		return 0
	}
	switch globalLayout {
	case "wallclock":
		return XDuration(r.FinishTime().Sub(r.StartTime()))
	case "compact":
		return r.Layout().W()
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
		return r.Layout().H()+2
	default:
		panic(errors.Errorf("invalid layout %q", globalLayout))
	}
}

var restartTemplateWallclock = template.Must(template.
	New("<x-restart>").
	Funcs(funcMap).
	Parse(`<svg class="restart"
		    x="{{ .Attrs.X.PercentOf .Data.Parent.W }}" y="{{ .Attrs.Y.EM }}"
		    width="{{ .Data.W.PercentOf .Data.Parent.W }}" height="{{ .Data.H.EM }}">
		<title xml:space="preserve">{{ .Data.Title }}</title>
		<rect class="background" x="0" y="0" width="100%" height="100%" />
		{{ $yoff := 0 | asYLines }}
		{{ range .Data.TimeSortedRecipes }}
			{{ $xoff := (.StartTime.Sub $.Data.StartTime) | asXDuration }}
			{{ .SVG $xoff $yoff }}
			{{ $yoff = $yoff.Add .H }}
		{{ end }}
	</svg>`))

var restartTemplateCompact = template.Must(template.
	New("<x-restart>").
	Funcs(funcMap).
	Parse(`<svg class="restart"
		    x="{{ .Attrs.X.PercentOf .Data.Parent.W }}" y="{{ .Attrs.Y.EM }}"
		    width="{{ .Data.W.PercentOf .Data.Parent.W }}" height="{{ .Data.H.EM }}">
		<title xml:space="preserve">{{ .Data.Title }}</title>
		<rect class="background" x="0" y="0" width="100%" height="100%" />
		{{ range .Data.TimeSortedRecipes }}
			{{ .SVG ($.Data.Layout.X .) (($.Data.Layout.Y .).Add 1) }}
		{{ end }}
	</svg>`))

func (r *SVGRestart) SVG(X XDuration, Y YLines) (template.HTML, error) {
	var restartTemplate *template.Template
	switch globalLayout {
	case "wallclock":
		restartTemplate = restartTemplateWallclock
	case "compact":
		restartTemplate = restartTemplateCompact
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
