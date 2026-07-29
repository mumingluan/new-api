package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetXuancatMarker(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/457", nil)

	GetXuancatMarker(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.JSONEq(t, `{"457":true}`, recorder.Body.String())
}
