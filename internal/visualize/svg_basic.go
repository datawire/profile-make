package visualize

import (
	"fmt"
	"html/template"
	"strings"
	"time"
)

const LineHeightEM = 1.2

var (
	globalProfile        *SVGProfile
	globalLayout         string
	globalVerboseCommand bool
)

var funcMap = template.FuncMap{
	"asYLines":       func(y int) YLines { return YLines(y) },
	"asXDuration":    func(x time.Duration) XDuration { return XDuration(x) },
	"split":          func(sep, input string) []string { return strings.Split(input, sep) },
	"verboseCommand": func() bool { return globalVerboseCommand },
}

type XDuration time.Duration

func (d XDuration) PercentOf(parent XDuration) string {
	return fmt.Sprintf("%f%%", 100*float64(d)/float64(parent))
}

func (a XDuration) Add(b XDuration) XDuration {
	return a + b
}

type YLines int

func (a YLines) Add(b YLines) YLines {
	return a + b
}

func (y YLines) EM() string {
	return fmt.Sprintf("%fem", float64(y)*LineHeightEM)
}

type SVGElement interface {
	StartTime() time.Time
	FinishTime() time.Time
	W() XDuration
	H() YLines
	SVG(x XDuration, y YLines) (template.HTML, error)
}

var _ SVGElement = &SVGMake{}
var _ SVGElement = &SVGRestart{}
var _ SVGElement = &SVGRecipe{}
var _ SVGElement = &SVGCommand{}
