package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestShouldRetryRejectsCommittedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	_, err := context.Writer.Write([]byte("partial"))
	require.NoError(t, err)

	apiErr := types.NewOpenAIError(
		http.ErrHandlerTimeout,
		types.ErrorCodeBadResponse,
		http.StatusInternalServerError,
	)
	require.False(t, shouldRetry(context, apiErr, 1))
}
