package visualize

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"time"

	"github.com/alessio/shellescape"
)

type SVGCommand struct {
	Parent   *SVGRecipe
	Raw      RawCommand
	SubMakes map[string]*SVGMake // key is CURDIR
}

func (cmd *SVGCommand) Text() string {
	if cmd == nil {
		return ""
	}
	args := make([]string, len(cmd.Raw.Args))
	for i := range cmd.Raw.Args {
		args[i] = shellescape.Quote(cmd.Raw.Args[i])
	}
	return strings.Join(args, " ")
}

func (cmd *SVGCommand) Title() string {
	target, err := filepath.Rel(globalProfile.Make.Dir, cmd.Raw.RecipeTarget)
	if err != nil {
		target = cmd.Raw.RecipeTarget
	}
	return fmt.Sprintf("Make/Restart/Recipe/Command\n"+
		"Target: %q\n"+
		"Duration: %s\n"+
		"Command: \n%s",
		target,
		cmd.FinishTime().Sub(cmd.StartTime()),
		cmd.Text())
}

func (cmd *SVGCommand) BaseH() YLines {
	if globalVerboseCommand {
		return YLines(strings.Count(cmd.Text(), "\n") + 1)
	} else {
		return 1
	}
}

////////////////////////////////////////////////////////////////////////////////

func (cmd *SVGCommand) StartTime() time.Time {
	if cmd == nil {
		return time.Time{}
	}
	return cmd.Raw.StartTime
}

func (cmd *SVGCommand) FinishTime() time.Time {
	if cmd == nil {
		return time.Time{}
	}
	return cmd.Raw.FinishTime
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
	Parse(`<svg class="command"
		    x="{{ .Attrs.X.PercentOf .Data.Parent.W }}" y="{{ .Attrs.Y.EM }}"
		    width="{{ .Data.W.PercentOf .Data.Parent.W }}" height="{{ .Data.H.EM }}">
		<title xml:space="preserve">{{ .Data.Title }}</title>
		<rect class="background" x="0" y="0" width="100%" height="100%" />
		<text x="0" y="0" dominant-baseline="hanging">
			{{ $dy := "0" }}
			{{ range $line := (.Data.Text | split "\n") }}
				<tspan x="0" dy="{{ $dy }}" xml:space="preserve">{{ $line }}</tspan>
				{{ $dy = (asYLines 1).EM }}
			{{ end }}
		</text>
		{{ $yoff := .Data.BaseH }}
		{{ range .Data.SubMakes }}
			{{ $xoff := (.StartTime.Sub $.Data.StartTime) | asXDuration }}
			{{ .SVG $xoff $yoff }}
			{{ $yoff = $yoff.Add .H }}
		{{ end }}
	</svg>`))

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
