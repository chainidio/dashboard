package cli

import (
	"github.com/chainid-io/dashboard"

	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"strings"
)

type pairList []chainid.Pair

// Set implementation for a list of chainid.Pair
func (l *pairList) Set(value string) error {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("expected NAME=VALUE got '%s'", value)
	}
	p := new(chainid.Pair)
	p.Name = parts[0]
	p.Value = parts[1]
	*l = append(*l, *p)
	return nil
}

// String implementation for a list of pair
func (l *pairList) String() string {
	return ""
}

// IsCumulative implementation for a list of pair
func (l *pairList) IsCumulative() bool {
	return true
}

func pairs(s kingpin.Settings) (target *[]chainid.Pair) {
	target = new([]chainid.Pair)
	s.SetValue((*pairList)(target))
	return
}
