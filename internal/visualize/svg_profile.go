package visualize

import (
	"html/template"
	"io"
	"time"
)

type SVGProfile struct {
	StartTime  time.Time
	FinishTime time.Time
	Make       *SVGMake
}

func (p *SVGProfile) Duration() time.Duration {
	return p.FinishTime.Sub(p.StartTime)
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
			{{ .Data.Make.SVG zeroTime zeroLines }}
		</g>
	</svg>`))

func (p *SVGProfile) SVG(w io.Writer) error {
	return profileTemplate.Execute(w, map[string]interface{}{
		"Data": p,
	})
}
