package cron

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/toolkits/file"
	"github.com/toolkits/logger"

	"github.com/coraldane/ops-common/model"
)

func StopDesiredAgent(last *model.DesiredAgent) error {
	if nil == last {
		return nil
	}

	agentVersionDir := path.Join(last.WorkDir, last.Name, last.Version)
	if !CheckFileExists(last.RunUser, path.Join(agentVersionDir, "control")) {
		return fmt.Errorf("control file not exists, agent name: %s", last.Name)
	}
	return ControlStopIn(last.RunUser, agentVersionDir)
}

func StopAgentOf(da *model.DesiredAgent, lastRunUser, lastWorkDir string) error {
	agentDir := path.Join(lastWorkDir, da.Name)
	version := ReadVersion(lastRunUser, agentDir)

	if "" == version {
		version = da.Version
	}
	if version == da.Version && da.RunUser == lastRunUser && da.WorkDir == lastWorkDir {
		// do nothing
		return nil
	}

	versionDir := path.Join(agentDir, version)
	if !CheckDirectoryExists(lastRunUser, versionDir) {
		logger.Warn("user: %s, %s nonexistent", lastRunUser, versionDir)
		return nil
	}

	err := ControlStopIn(lastRunUser, versionDir)
	if nil != err {
		return err
	}

	cmd := BuildCommand(lastRunUser, "rm", "-rf", versionDir)
	cmd.Dir = file.SelfDir()
	_, err = ExecuteCommandWithOutput(cmd)
	return err
}

func ControlStopIn(runUser, workdir string) error {
	if !CheckDirectoryExists(runUser, workdir) {
		return nil
	}

	out, err := ControlStatus(runUser, workdir)
	if err == nil && strings.Contains(out, "stoped") {
		return nil
	}

	_, err = ControlStop(runUser, workdir)
	if err != nil {
		return err
	}

	time.Sleep(time.Second * 3)

	out, err = ControlStatus(runUser, workdir)
	if err == nil && strings.Contains(out, "stoped") {
		return nil
	}

	return err
}
