package misc

import (
	"fmt"
	"testing"
)

func TestCheckError_Should_Panic_On_Error(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		} else {
			fmt.Printf("%v\n", r)
		}
	}()

	CheckError(fmt.Errorf("a wrapped error"), "wrapping error")
}
