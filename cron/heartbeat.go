package cron

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/toolkits/logger"
	"github.com/toolkits/net/httplib"

	"github.com/coraldane/ops-common/model"
	"github.com/coraldane/ops-common/utils"
	"github.com/coraldane/ops-updater/g"
)

func Heartbeat() {
	SleepRandomDuration()
	for {
		heartbeat()
		d := time.Duration(g.Config().Interval) * time.Second
		time.Sleep(d)
	}
}

func heartbeat() {
	hostname, err := utils.Hostname(g.Config().Hostname)
	if err != nil {
		return
	}

	desiredAgents := ReadDesiredAgents()
	heartbeatRequest := BuildHeartbeatRequest(hostname, desiredAgents)
	logger.Debugln("===>>>", heartbeatRequest)

	bs, err := json.Marshal(heartbeatRequest)
	if err != nil {
		logger.Errorln("encode heartbeat request fail", err)
		return
	}

	url := fmt.Sprintf("http://%s/api/heartbeat", g.Config().Server)
	httpRequest := httplib.Post(url).SetTimeout(time.Second*10, time.Minute)
	httpRequest.Body(bs)
	httpResponse, err := httpRequest.Bytes()
	if err != nil {
		log.Printf("curl %s fail %v", url, err)
		return
	}

	var heartbeatResponse model.HeartbeatResponse
	err = json.Unmarshal(httpResponse, &heartbeatResponse)
	if err != nil {
		logger.Errorln("decode heartbeat response fail", err)
		return
	}

	logger.Debugln("<<<<====", heartbeatResponse)

	HandleHeartbeatResponse(&heartbeatResponse, desiredAgents)
}
