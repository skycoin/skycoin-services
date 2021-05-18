package main

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func init() {
	gin.SetMode(gin.TestMode)
	logFileName = "uuid-test.log"
	configureLogger()
}

func TestGetRouter(t *testing.T) {
	t.Cleanup(clearLogFile)
	r := getRouter()
	assert.IsType(t, &gin.Engine{}, r)
}

func TestHandleSystemSurvey(t *testing.T) {
	t.Cleanup(clearLogFile)
	r := getRouter()
	req := httptest.NewRequest(http.MethodPost, "/system_survey", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.FileExists(t, logFileName)
}

func clearLogFile() {
	_ = os.Remove(logFileName)
}
