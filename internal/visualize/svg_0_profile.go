package visualize

import (
	"html/template"
	"io"
	"time"

	"github.com/pkg/errors"
)

type SVGProfile struct {
	StartTime  time.Time
	FinishTime time.Time
	Make       *SVGMake
}

func (p *SVGProfile) Duration() time.Duration {
	return p.FinishTime.Sub(p.StartTime)
}

func (p *SVGProfile) W() XDuration {
	switch globalLayout {
	case "wallclock":
		return XDuration(p.Duration())
	case "compact":
		return p.Make.W()
	default:
		panic(errors.Errorf("invalid layout %q", globalLayout))
	}
}

func (p *SVGProfile) H() YLines {
	return p.Make.H()
}

var profileTemplate = template.Must(template.
	New("<x-profile>").
	Funcs(funcMap).
	Parse(`<svg xmlns="http://www.w3.org/2000/svg"
		  width="100%"
		  height="{{ .Data.H.EM }}" >
		<style>
			g.command > rect {
				fill: #AAAAAA;
			}
		</style>
		<g>
			{{ .Data.Make.SVG (asXDuration 0) (asYLines 0) }}
		</g>
	</svg>`))

func (p *SVGProfile) SVG(w io.Writer, layout string) error {
	globalProfile = p
	globalLayout = layout
	return profileTemplate.Execute(w, map[string]interface{}{
		"Data": p,
	})
}
