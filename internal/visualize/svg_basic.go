package visualize

import (
	"fmt"
	"html/template"
	"strings"
	"time"
)

const LineHeightEM = 1.2

var funcMap = template.FuncMap{
	"addLines":     func(a, b YLines) YLines { return a + b },
	"zeroLines":    func() YLines { return 0 },
	"zeroTime":     func() XTime { return XTime{} },
	"split":        func(sep, input string) []string { return strings.Split(input, sep) },
	"lineHeightEM": func() string { return YLines(1).EM() },
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
	return fmt.Sprintf("%fem", float64(y)*LineHeightEM)
}
