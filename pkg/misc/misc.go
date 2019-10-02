package misc

import "github.com/pkg/errors"

func CheckError(err error, message string) {
	if err != nil {
		panic(errors.Wrap(err, message))
	}
}
