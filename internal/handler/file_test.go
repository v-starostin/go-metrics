package handler_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/v-starostin/go-metrics/internal/handler"
	"github.com/v-starostin/go-metrics/internal/mock"
)

func TestFile(t *testing.T) {
	t.Run("WriteToFile", func(t *testing.T) {
		svc := &mock.Service{}
		f := handler.NewFile1(svc)
		svc.On("WriteToFile").Return(nil)
		assert.NoError(t, f.WriteToFile())
	})

	t.Run("RestoreFromFile", func(t *testing.T) {
		svc := &mock.Service{}
		f := handler.NewFile1(svc)
		svc.On("RestoreFromFile").Return(nil)
		assert.NoError(t, f.RestoreFromFile())
	})
}
