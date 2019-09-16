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
		<defs>
			<filter id="inset-shadow-black">
				<!-- We implicitly start with an opaque black rectangle -->
				<!-- Set the alpha-chanel to an inverted copy of the source alpha-channel -->
				<feComponentTransfer in="SourceAlpha">
					<feFuncA type="table" tableValues="1 0" />
				</feComponentTransfer>
				<!-- Blur it; bleed the shadow in to the image -->
				<feGaussianBlur stdDeviation="4" />
				<!-- Clip to the source image; only leave what blead in to the image -->
				<feComposite in2="SourceAlpha" operator="in" />
				<!-- Dye it black -->
				<feOffset dx="0" dy="0" result="shape"/>
				<feFlood flood-color="#000000" result="color"/>
				<feComposite in="color" in2="shape" operator="in" />
				<!-- Overlay the shadow on top of the source image -->
				<feMerge>
					<feMergeNode in="SourceGraphic" />
					<feMergeNode />
				</feMerge>
			</filter>
			<filter id="inset-shadow-red">
				<!-- We implicitly start with an opaque black rectangle -->
				<!-- Set the alpha-chanel to an inverted copy of the source alpha-channel -->
				<feComponentTransfer in="SourceAlpha">
					<feFuncA type="table" tableValues="1 0" />
				</feComponentTransfer>
				<!-- Blur it; bleed the shadow in to the image -->
				<feGaussianBlur stdDeviation="4" />
				<!-- Clip to the source image; only leave what blead in to the image -->
				<feComposite in2="SourceAlpha" operator="in" />
				<!-- Dye it red -->
				<feOffset dx="0" dy="0" result="shape"/>
				<feFlood flood-color="#FF0000" result="color"/>
				<feComposite in="color" in2="shape" operator="in" />
				<!-- Overlay the shadow on top of the source image -->
				<feMerge>
					<feMergeNode in="SourceGraphic" />
					<feMergeNode />
				</feMerge>
			</filter>
			<filter id="inset-shadow-green">
				<!-- We implicitly start with an opaque black rectangle -->
				<!-- Set the alpha-chanel to an inverted copy of the source alpha-channel -->
				<feComponentTransfer in="SourceAlpha">
					<feFuncA type="table" tableValues="1 0" />
				</feComponentTransfer>
				<!-- Blur it; bleed the shadow in to the image -->
				<feGaussianBlur stdDeviation="1" />
				<!-- Clip to the source image; only leave what blead in to the image -->
				<feComposite in2="SourceAlpha" operator="in" />
				<!-- Dye it red -->
				<feOffset dx="0" dy="0" result="shape"/>
				<feFlood flood-color="#AAFFAA" result="color"/>
				<feComposite in="color" in2="shape" operator="in" />
				<!-- Overlay the shadow on top of the source image -->
				<feMerge>
					<feMergeNode in="SourceGraphic" />
					<feMergeNode />
				</feMerge>
			</filter>
		</defs>
		<style>
			* svg {
				overflow: {{ if verboseCommand }}visible{{ else }}hidden{{ end }};
			}
			svg.make                  { filter: url(#inset-shadow-black); }
			svg.make > .background    { fill: #CCCCCC;	}

			svg.restart               { filter: url(#inset-shadow-red); }
			svg.restart > .background { fill: #999999; }

			svg.recipe                { filter: url(#inset-shadow-black); }
			svg.recipe > .background  { fill: #666666; }

			svg.command               { }
			svg.command > .background { fill: #333333; filter: url(#inset-shadow-green); }
			svg.command > text        { fill: #FFFFFF; }
		</style>
		<g>
			{{ .Data.Make.SVG (asXDuration 0) (asYLines 0) }}
		</g>
	</svg>`))

func (p *SVGProfile) SVG(w io.Writer, layout string, verboseCommand bool) error {
	globalProfile = p
	globalLayout = layout
	globalVerboseCommand = verboseCommand
	return profileTemplate.Execute(w, map[string]interface{}{
		"Data": p,
	})
}
