package visualize

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/alessio/shellescape"
)

var funcMap = template.FuncMap{
	"addLines":  func(a, b YLines) YLines { return a + b },
	"zeroLines": func() YLines { return 0 },
	"split":     func(sep, input string) []string { return strings.Split(input, sep) },
}

type XTime struct {
	Profile *SVGProfile
	X       time.Time
}

func (x XTime) Duration() time.Duration {
	return x.X.Sub(x.Profile.StartTime)
}

func (x XTime) Percent() string {
	return fmt.Sprintf("%f%%", 100*float64(x.Duration())/float64(x.Profile.Duration()))
}

type XDuration struct {
	Profile *SVGProfile
	W       time.Duration
}

func (w XDuration) Duration() time.Duration {
	return w.W
}

func (w XDuration) Percent() string {
	return fmt.Sprintf("%f%%", 100*float64(w.Duration())/float64(w.Profile.Duration()))
}

type YLines int

func (y YLines) EM() string {
	return fmt.Sprintf("%dem", y)
}

type SVGProfile struct {
	StartTime  time.Time
	FinishTime time.Time
	Commands   []SVGCommand
}

func (p *SVGProfile) Duration() time.Duration {
	return p.FinishTime.Sub(p.StartTime)
}

func (p *SVGProfile) H() YLines {
	var ret YLines
	for _, cmd := range p.Commands {
		ret += cmd.H()
	}
	return ret
}

type SVGCommand struct {
	X XTime
	W XDuration

	Y YLines

	Args        []string
	SubCommands []SVGCommand
}

func (cmd SVGCommand) Text() string {
	args := make([]string, len(cmd.Args))
	for i := range cmd.Args {
		args[i] = shellescape.Quote(cmd.Args[i])
	}
	return strings.Join(args, " ")
}

func (cmd SVGCommand) BaseH() YLines {
	return YLines(strings.Count(cmd.Text(), "\n") + 1)
}

func (cmd SVGCommand) H() YLines {
	ret := cmd.BaseH()
	for _, subcmd := range cmd.SubCommands {
		ret += subcmd.H()
	}
	return ret
}

func (cmd SVGCommand) SVG(Y YLines) (template.HTML, error) {
	cmd.Y = Y
	var str strings.Builder
	err := template.Must(template.
		New("<x-command>").
		Funcs(funcMap).
		Parse(`<g>
			<rect x="{{ .X.Percent }}" width="{{ .W.Percent}}" y="{{ .Y.EM }}" height="{{ .H.EM }}" style="fill: #AAAAAA; border: #FF0000" />
			<text x="{{ .X.Percent }}" y="{{ .Y.EM }}" color="#000000" dominant-baseline="hanging">
				{{ $dy := "0" }}
				{{ range $line := (.Text | split "\n") }}
					<tspan x="{{ $.X.Percent }}" dy="{{ $dy }}" xml:space="preserve">{{ $line }}</tspan>
					{{ $dy = "1em" }}
				{{ end }}
			</text>
			{{ $y := addLines .Y .BaseH }}
			{{ range $subcmd := .SubCommands }}{{ $subcmd.SVG $y }}{{ $y = addLines $y $subcmd.H }}{{ end }}
		</g>`)).Execute(&str, cmd)
	if err != nil {
		return "", err
	}
	return template.HTML(str.String()), nil
}

var mainTemplate = template.Must(template.
	New("<x-profile>").
	Funcs(funcMap).
	Parse(`{{ "" -}}
<svg xmlns="http://www.w3.org/2000/svg"
  width="100%"
  height="{{ .ProfileData.H.EM }}"
  >
	{{ $y := zeroLines }}
	{{ range $cmd := .ProfileData.Commands }}{{ $cmd.SVG $y }}{{ $y = addLines $y $cmd.H }}{{ end }}
</svg>`))
