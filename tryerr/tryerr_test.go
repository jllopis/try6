package tryerr

import "testing"

func TestErrors(t *testing.T) {
	e := ErrInvalidContext
	t.Logf("e = %s", e)
}
