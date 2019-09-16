package visualize

import (
	"html/template"
	"strings"
	"time"

	"github.com/alessio/shellescape"
)

type SVGCommand struct {
	startTime  time.Time
	finishTime time.Time

	Args     []string
	SubMakes map[string]*SVGMake // key is CURDIR
}

func (cmd *SVGCommand) Text() string {
	args := make([]string, len(cmd.Args))
	for i := range cmd.Args {
		args[i] = shellescape.Quote(cmd.Args[i])
	}
	return strings.Join(args, " ")
}

func (cmd *SVGCommand) BaseH() YLines {
	return YLines(strings.Count(cmd.Text(), "\n") + 1)
}

////////////////////////////////////////////////////////////////////////////////

func (cmd *SVGCommand) StartTime() time.Time {
	if cmd == nil {
		return time.Time{}
	}
	return cmd.startTime
}

func (cmd *SVGCommand) FinishTime() time.Time {
	if cmd == nil {
		return time.Time{}
	}
	return cmd.finishTime
}

func (cmd *SVGCommand) W() XDuration {
	if cmd == nil {
		return 0
	}
	return XDuration(cmd.FinishTime().Sub(cmd.StartTime()))
}

func (cmd *SVGCommand) H() YLines {
	sum := cmd.BaseH()
	for _, submake := range cmd.SubMakes {
		sum += submake.H()
	}
	return sum
}

var commandTemplate = template.Must(template.
	New("<x-command>").
	Funcs(funcMap).
	Parse(`<g class="command">
		<rect x="{{ .Attrs.X.Percent }}" y="{{ .Attrs.Y.EM }}" 
		      width="{{ .Data.W.Percent}}" height="{{ .Data.H.EM }}" />
		<text x="{{ .Attrs.X.Percent }}" y="{{ .Attrs.Y.EM }}" dominant-baseline="hanging">
			{{ $dy := "0" }}
			{{ range $line := (.Data.Text | split "\n") }}
				<tspan x="{{ $.Attrs.X.Percent }}" dy="{{ $dy }}" xml:space="preserve">{{ $line }}</tspan>
				{{ $dy = (asYLines 1).EM }}
			{{ end }}
		</text>
		{{ $yoff := .Data.BaseH }}
		{{ range .Data.SubMakes }}
			{{ $xoff := (.StartTime.Sub $.Data.StartTime) | asXDuration }}
			{{ .SVG ($.Attrs.X.Add $xoff) ($.Attrs.Y.Add $yoff) }}
			{{ $yoff = $yoff.Add .H }}
		{{ end }}
	</g>`))

func (cmd *SVGCommand) SVG(X XDuration, Y YLines) (template.HTML, error) {
	var str strings.Builder
	err := commandTemplate.Execute(&str, map[string]interface{}{
		"Attrs": map[string]interface{}{
			"X": X,
			"Y": Y,
		},
		"Data": cmd,
	})
	if err != nil {
		return "", err
	}
	return template.HTML(str.String()), nil
}
