package cron

import (
	"encoding/json"
	"fmt"
	"log"
	"path"
	"runtime"

	"github.com/toolkits/file"
	"github.com/toolkits/logger"

	"github.com/coraldane/ops-common/model"
)

func HandleHeartbeatResponse(respone *model.HeartbeatResponse, lastDesiredAgents []*model.DesiredAgent) {
	if respone.ErrorMessage != "" {
		log.Println("receive error message:", respone.ErrorMessage)
		return
	}

	das := respone.DesiredAgents
	if das == nil || len(das) == 0 {
		return
	}

	ldap := make(map[string]*model.DesiredAgent)
	for _, val := range lastDesiredAgents {
		ldap[val.Name] = val
	}

	osArch := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	for _, da := range das {
		da.FillAttrs(osArch, da.WorkDir)
		HandleDesiredAgent(da, ldap[da.Name])
	}

	bs, _ := json.Marshal(das)
	file.WriteString(path.Join(file.SelfDir(), "desired_agent.json"), string(bs))
}

func HandleDesiredAgent(da *model.DesiredAgent, last *model.DesiredAgent) {
	var err error
	if da.Cmd == "start" {
		err = StartDesiredAgent(da, last)
	} else if da.Cmd == "stop" {
		err = StopDesiredAgent(last)
	} else {
		log.Println("unknown cmd", da)
	}
	if nil != err {
		logger.Error("%s %s fail, %v\n", da.Cmd, da.Name, err)
	}
}
