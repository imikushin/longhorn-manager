package manager

import (
	"fmt"
	"strings"
)

type Errs []error

func (errs Errs) Error() string {
	ss := []string{}
	for _, err := range errs {
		ss = append(ss, fmt.Sprintf("%+v", err))
	}
	return strings.Join(ss, "\n\n")
}
