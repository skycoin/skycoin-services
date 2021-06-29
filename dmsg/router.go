package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/skycoin/skycoin-services/system-survey/tests/apps"
	"github.com/skycoin/skycoin-services/system-survey/tests/dmsgtest"
	"github.com/skycoin/skycoin-services/system-survey/tests/golang"
	"github.com/skycoin/skycoin-services/system-survey/tests/httptest"
	"github.com/skycoin/skycoin-services/system-survey/tests/hwinfo"
	"github.com/skycoin/skycoin-services/system-survey/tests/netinfo"
	"github.com/skycoin/skycoin-services/system-survey/tests/traceroutetest"
	"net/http"
	"sync"
)

func getRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.POST("/system_survey", handleSystemSurvey)
	return router
}

type systemSurveyResponse struct {
	Network        interface{} `json:"network_info"`
	Hardware       interface{} `json:"hardware_info"`
	Golang         interface{} `json:"golang_info"`
	Apps           interface{} `json:"apps"`
	DmsgTest       interface{} `json:"dmsg_test"`
	TracerouteTest interface{} `json:"traceroute_test"`
	HttpTest       interface{} `json:"http_test"`
}

type systemSurveyRequest struct {
	Apps       []string              `json:"apps"`
	Dmsg       *dmsgtest.Input       `json:"dmsg"`
	Traceroute *traceroutetest.Input `json:"traceroute"`
	Http       *httptest.Input       `json:"http"`
}

func handleSystemSurvey(c *gin.Context) {
	req := systemSurveyRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	res := systemSurveyResponse{
		Network:  netinfo.Get(),
		Hardware: hwinfo.Run(),
		Golang:   golang.Run(),
	}

	wg := &sync.WaitGroup{}
	if req.Apps != nil {
		wg.Add(1)
		go func() {
			res.Apps = apps.Run(req.Apps)
			wg.Done()
		}()
	}

	if req.Dmsg != nil {
		wg.Add(1)
		go func() {
			res.DmsgTest = dmsgtest.Run(req.Dmsg)
			wg.Done()
		}()
	}

	if req.Traceroute != nil {
		wg.Add(1)
		go func() {
			res.TracerouteTest = traceroutetest.Trace(req.Traceroute)
			wg.Done()
		}()
	}

	if req.Http != nil {
		wg.Add(1)
		go func() {
			res.HttpTest = httptest.Run(req.Http)
			wg.Done()
		}()
	}
	wg.Wait()

	b, _ := json.Marshal(res)
	log.WithField("ip", c.ClientIP()).
		WithField("result", string(b)).
		Info("Success")
	c.JSON(http.StatusOK, res)
}
