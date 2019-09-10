package visualize

import (
	"html/template"
	"strings"

	"github.com/alessio/shellescape"
)

type SVGCommand struct {
	X XTime
	W XDuration

	Args    []string
	SubMake *SVGMake
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

func (cmd *SVGCommand) H() YLines {
	return cmd.BaseH() + cmd.SubMake.H()
}

var commandTemplate = template.Must(template.
	New("<x-command>").
	Funcs(funcMap).
	Parse(`<g class="command">
		<rect x="{{ .Data.X.Percent }}" y="{{ .Attrs.Y.EM }}" 
		      width="{{ .Data.W.Percent}}" height="{{ .Data.H.EM }}" />
		<text x="{{ .Data.X.Percent }}" y="{{ .Attrs.Y.EM }}" dominant-baseline="hanging">
			{{ $dy := "0" }}
			{{ range $line := (.Data.Text | split "\n") }}
				<tspan x="{{ $.Data.X.Percent }}" dy="{{ $dy }}" xml:space="preserve">{{ $line }}</tspan>
				{{ $dy = lineHeightEM }}
			{{ end }}
		</text>
		{{ .SubMake.SVG .Data.X (addLines .Attrs.Y .Data.BaseH) }}
	</g>`))

func (cmd *SVGCommand) SVG(X XTime, Y YLines) (template.HTML, error) {
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
