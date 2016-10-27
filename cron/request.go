package cron

import (
	"encoding/json"
	"path"
	"time"

	f "github.com/toolkits/file"
	"github.com/toolkits/logger"

	"github.com/coraldane/ops-common/model"
	"github.com/coraldane/ops-common/utils"
	"github.com/coraldane/ops-updater/g"
)

func BuildHeartbeatRequest(hostname string, desiredAgents []*model.DesiredAgent) model.HeartbeatRequest {
	req := model.HeartbeatRequest{Hostname: hostname}
	req.Ip = utils.GetLocalIP()
	req.UpdaterVersion = g.VERSION
	req.RunUser = CurrentUser

	realAgents := []*model.RealAgent{}
	now := time.Now().Unix()

	for _, da := range desiredAgents {
		agentDir := path.Join(da.WorkDir, da.Name)
		// 如果目录下没有.version，我们认为这根本不是一个agent
		version := ReadVersion(da.RunUser, agentDir)
		if "" == version {
			logger.Error("read %s/.version fail\n", agentDir)
			continue
		}

		controlFile := path.Join(agentDir, version, "control")
		if !CheckFileExists(da.RunUser, controlFile) {
			logger.Errorln(controlFile, "is nonexistent, user:", da.RunUser)
			continue
		}

		cmd := BuildCommand(da.RunUser, path.Join(agentDir, version, "control"), "status")
		cmd.Dir = f.SelfDir()
		status, err := ExecuteCommandWithOutput(cmd)
		if err != nil {
			status = err.Error()
		}

		realAgent := &model.RealAgent{
			Name:      da.Name,
			Version:   version,
			Status:    status,
			Timestamp: now,
		}
		realAgent.RunUser = da.RunUser
		realAgent.WorkDir = da.WorkDir

		realAgents = append(realAgents, realAgent)
	}

	req.RealAgents = realAgents
	return req
}

func ReadDesiredAgents() []*model.DesiredAgent {
	var desiredAgents []*model.DesiredAgent
	strJson, err := f.ToTrimString(path.Join(f.SelfDir(), "desired_agent.json"))
	if nil != err {
		logger.Errorln("read desired agent file error", err)
		return desiredAgents
	}
	err = json.Unmarshal([]byte(strJson), &desiredAgents)
	if nil != err {
		logger.Errorln("unmarshal json error", strJson, err)
	}

	for _, da := range desiredAgents {
		actualVersion := ReadVersion(da.RunUser, path.Join(da.WorkDir, da.Name))
		if "" != actualVersion {
			da.Version = actualVersion
		}
	}
	return desiredAgents
}
