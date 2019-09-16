package visualize

import (
	"html/template"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type SVGMake struct {
	Parent   *SVGCommand
	Dir      string
	Restarts []*SVGRestart
}

func (m *SVGMake) StartTime() time.Time {
	if m == nil || len(m.Restarts) == 0 {
		return time.Time{}
	}
	return m.Restarts[0].StartTime()
}

func (m *SVGMake) FinishTime() time.Time {
	if m == nil || len(m.Restarts) == 0 {
		return time.Time{}
	}
	return m.Restarts[len(m.Restarts)-1].FinishTime()
}

func (m *SVGMake) W() XDuration {
	if m == nil {
		return 0
	}
	switch globalLayout {
	case "wallclock":
		return XDuration(m.FinishTime().Sub(m.StartTime()))
	case "compact":
		var sum XDuration
		for _, restart := range m.Restarts {
			sum += restart.W()
		}
		return sum
	default:
		panic(errors.Errorf("invalid layout %q", globalLayout))
	}
}

func (m *SVGMake) H() YLines {
	if m == nil {
		return 0
	}
	var max YLines
	for _, restart := range m.Restarts {
		if restartH := restart.H(); restartH > max {
			max = restartH
		}
	}
	return max
}

var makeTemplateWallclock = template.Must(template.
	New("<x-make>").
	Funcs(funcMap).
	Parse(`<g class="make">
		<rect x="{{ .Attrs.X.Percent }}" y="{{ .Attrs.Y.EM }}"
		      width="{{ .Data.W.Percent }}" height="{{ .Data.H.EM }}" />
		{{ range .Data.Restarts }}
			{{ $xoff := (.StartTime.Sub $.Data.StartTime) | asXDuration }}
			{{ .SVG ($.Attrs.X.Add $xoff) ($.Attrs.Y) }}
		{{ end }}
	</g>`))

var makeTemplateCompact = template.Must(template.
	New("<x-make>").
	Funcs(funcMap).
	Parse(`<g class="make">
		<rect x="{{ .Attrs.X.Percent }}" y="{{ .Attrs.Y.EM }}"
		      width="{{ .Data.W.Percent }}" height="{{ .Data.H.EM }}" />
		{{ $xoff := 0 | asXDuration }}
		{{ range .Data.Restarts }}
			{{ .SVG ($.Attrs.X.Add $xoff) ($.Attrs.Y) }}
			{{ $xoff = $xoff.Add .W }}
		{{ end }}
	</g>`))

func (m *SVGMake) SVG(X XDuration, Y YLines) (template.HTML, error) {
	var makeTemplate *template.Template
	switch globalLayout {
	case "wallclock":
		makeTemplate = makeTemplateWallclock
	case "compact":
		makeTemplate = makeTemplateCompact
	default:
		panic(errors.Errorf("invalid layout %q", globalLayout))
	}

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
