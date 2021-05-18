package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/skycoin/skycoin-services/system-survey/cmd/apps"
	"github.com/skycoin/skycoin-services/system-survey/cmd/dmsgtest"
	"github.com/skycoin/skycoin-services/system-survey/cmd/golang"
	"github.com/skycoin/skycoin-services/system-survey/cmd/hwinfo"
	"github.com/skycoin/skycoin-services/system-survey/cmd/netinfo"
	"net/http"
)

func getRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.POST("/system_survey", handleSystemSurvey)
	return router
}

type systemSurveyResponse struct {
	Network  interface{} `json:"network_info"`
	Hardware interface{} `json:"hardware_info"`
	Golang   interface{} `json:"golang_info"`
	Apps     interface{} `json:"apps"`
	DmsgTest interface{} `json:"dmsg_test"`
}

func handleSystemSurvey(c *gin.Context) {
	res := systemSurveyResponse{
		Network:  netinfo.Get(),
		Hardware: hwinfo.Run(),
		Golang:   golang.Run(),
		Apps:     apps.Run(nil),
		DmsgTest: dmsgtest.Run(),
	}

	b, _ := json.Marshal(res)
	log.WithField("ip", c.ClientIP()).
		WithField("result", string(b)).
		Info("Success")
	c.JSON(http.StatusOK, res)
}
