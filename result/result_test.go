package result_test

import (
	"errors"
	"testing"

	"github.com/denvrdata/go-denvr/result"
	"github.com/stretchr/testify/assert"
)

func TestResult(t *testing.T) {
	t.Run("Ok", func(t *testing.T) { assert.True(t, result.Wrap(5, nil).Ok()) })
	t.Run(
		"Panic",
		func(t *testing.T) {
			err := errors.New("Test Error")
			res := result.Wrap(5, err)
			assert.Panics(t, func() { res.Unwrap() }, "Panics with Test Error")
		},
	)
}
