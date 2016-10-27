package cron

import (
	"fmt"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/toolkits/file"
	"github.com/toolkits/logger"
)

func Control(runUser, workdir, arg string) (string, error) {
	cmd := BuildCommand(runUser, path.Join(workdir, "control"), arg)
	cmd.Dir = file.SelfDir()
	out, err := ExecuteCommandWithOutput(cmd)
	if err != nil {
		logger.Error("cd %s; ./control %s fail %v. output: %s\n", workdir, arg, err, out)
	}
	return out, err
}

func ControlStatus(runUser, workdir string) (string, error) {
	return Control(runUser, workdir, "status")
}

func ControlStart(runUser, workdir string) (string, error) {
	return Control(runUser, workdir, "start")
}

func ControlStop(runUser, workdir string) (string, error) {
	return Control(runUser, workdir, "stop")
}

func BuildCommand(username string, args ...string) *exec.Cmd {
	params := []string{}
	if CurrentUser == username {
		params = append(params, "-c")
		for _, key := range args {
			params = append(params, key)
		}
		return exec.Command("sh", params...)
	} else {
		params = append(params, "-u")
		params = append(params, username)
		for _, key := range args {
			params = append(params, key)
		}
		return exec.Command("sudo", params...)
	}
}

func ExecuteCommandWithOutput(cmd *exec.Cmd) (string, error) {
	bs, err := cmd.CombinedOutput()
	if nil != err && strings.Contains(err.Error(), "exit status") {
		return "", fmt.Errorf(string(bs))
	}
	return string(bs), err
}

func CheckDirectoryExists(username string, fp string) bool {
	return CheckFileOrDirExists(username, fp, "dir")
}

func CheckFileExists(username string, fp string) bool {
	return CheckFileOrDirExists(username, fp, "file")
}

func CheckFileOrDirExists(username, fp, fileType string) bool {
	if CurrentUser == username {
		return file.IsExist(fp)
	} else {
		cmd := BuildCommand(username, "sh", "check_file.sh", fileType, fp)
		cmd.Dir = file.SelfDir()
		strOut, err := ExecuteCommandWithOutput(cmd)
		if nil != err {
			logger.Errorln("check dir exists", strOut, err)
			return false
		}
		result, _ := strconv.ParseBool(strings.Replace(strOut, "\n", "", -1))
		return result
	}
	return false
}

func ReadVersion(username, agentDir string) string {
	versionFile := path.Join(agentDir, ".version")
	cmd := BuildCommand(username, "sh", "read_file.sh", versionFile)
	cmd.Dir = file.SelfDir()
	version, err := ExecuteCommandWithOutput(cmd)
	if err != nil {
		logger.Warn("%s is nonexistent,error: %v", versionFile, err)
		return ""
	}
	return version
}
