package pkg

import (
	"fmt"
	"os"
	"strings"

	"github.com/nikhilsbhat/common/errors"
)

type ValueFiles []string

func (v *ValueFiles) String() string {
	return fmt.Sprint(*v)
}

//nolint:goerr113
func (v *ValueFiles) Valid() error {
	errStr := strings.Builder{}

	for _, valuesFile := range *v {
		if strings.TrimSpace(valuesFile) != "-" {
			if _, err := os.Stat(valuesFile); os.IsNotExist(err) {
				errStr.WriteString(err.Error())
			}
		}
	}

	if errStr.Len() == 0 {
		return nil
	}

	return &errors.CommonError{Message: errStr.String()}
}

func (v *ValueFiles) Type() string {
	return "ValueFiles"
}

func (v *ValueFiles) Set(value string) error {
	for filePath := range strings.SplitSeq(value, ",") {
		*v = append(*v, filePath)
	}

	return nil
}
