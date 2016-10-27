package cron

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/toolkits/file"
	"github.com/toolkits/logger"

	"github.com/coraldane/ops-common/model"
	"github.com/coraldane/ops-common/utils"
)

func StartDesiredAgent(da *model.DesiredAgent, last *model.DesiredAgent) error {
	if err := InsureRunUserExists(da); nil != err {
		return err
	}

	var lastRunUser, lastWorkDir string
	if nil != last {
		lastRunUser = last.RunUser
		lastWorkDir = last.WorkDir
	}

	if err := StopAgentOf(da, lastRunUser, lastWorkDir); err != nil {
		return err
	}

	if err := InsureNewVersion(da); nil != err {
		return err
	}

	if err := ControlStartIn(da.RunUser, da.AgentVersionDir); err != nil {
		logger.Errorln("ControlStartIn error", err)
		return err
	}

	if err := WriteVersion(da); nil != err {
		logger.Errorln("WriteVersion error", err)
		return err
	}
	return nil
}

func InsureNewVersion(da *model.DesiredAgent) error {
	if err := InsureDesiredAgentDirExists(da); err != nil {
		return err
	}

	if err := InsureNewVersionFiles(da); err != nil {
		return err
	}

	if err := Untar(da); err != nil {
		return err
	}

	return nil
}

func WriteVersion(da *model.DesiredAgent) (err error) {
	if CurrentUser == da.RunUser {
		file.WriteString(path.Join(da.AgentDir, ".version"), da.Version)
	} else {
		file.WriteString(path.Join(file.SelfDir(), ".version"), da.Version)

		_, err = utils.ExecuteCommand(file.SelfDir(), fmt.Sprintf("sudo mv .version %s/", da.AgentDir))
		if nil != err {
			return
		}
		_, err = utils.ExecuteCommand(file.SelfDir(), fmt.Sprintf("sudo chown -R %s:%s %s", da.RunUser, da.RunUser, path.Join(da.AgentDir, ".version")))
	}
	return
}

func Untar(da *model.DesiredAgent) error {
	cmd := BuildCommand(da.RunUser, "tar", "zxf", path.Join(da.AgentVersionDir, da.TarballFilename), "-C", da.AgentVersionDir)
	cmd.Dir = file.SelfDir()
	err := cmd.Run()
	if err != nil {
		log.Println("tar zxf", da.TarballFilename, "fail", err)
		return err
	}
	return nil
}

func ControlStartIn(runUser, workdir string) error {
	out, err := ControlStatus(runUser, workdir)
	if err == nil && strings.Contains(out, "started") {
		return nil
	}

	_, err = ControlStart(runUser, workdir)
	if err != nil {
		return err
	}

	time.Sleep(time.Second * 3)

	out, err = ControlStatus(runUser, workdir)
	if err == nil && strings.Contains(out, "started") {
		return nil
	}
	return err
}

func InsureNewVersionFiles(da *model.DesiredAgent) error {
	if FilesReady(da) {
		return nil
	}

	downloadTarballCmd := BuildCommand(da.RunUser, "wget", "-q", da.TarballUrl, "-O", path.Join(da.AgentVersionDir, da.TarballFilename))
	downloadTarballCmd.Dir = file.SelfDir()
	_, err := ExecuteCommandWithOutput(downloadTarballCmd)
	if err != nil {
		logger.Errorln("wget -q", da.TarballUrl, "-O", da.TarballFilename, "fail", err)
		return err
	}

	downloadMd5Cmd := BuildCommand(da.RunUser, "wget", "-q", da.Md5Url, "-O", path.Join(da.AgentVersionDir, da.Md5Filename))
	downloadMd5Cmd.Dir = file.SelfDir()
	_, err = ExecuteCommandWithOutput(downloadMd5Cmd)
	if err != nil {
		log.Println("wget -q", da.Md5Url, "-O", da.Md5Filename, "fail", err)
		return err
	}

	if "" != da.ConfigFileName && "" != da.ConfigRemoteUrl {
		downloadConfigCmd := BuildCommand(da.RunUser, "wget", "-q", da.ConfigRemoteUrl, "-O", path.Join(da.AgentVersionDir, da.ConfigFileName))
		downloadConfigCmd.Dir = file.SelfDir()
		_, err := ExecuteCommandWithOutput(downloadConfigCmd)
		if err != nil {
			logger.Errorln("wget -q", da.ConfigRemoteUrl, "-O", da.ConfigFileName, "fail", err)
		}
		return err
	}

	return Md5sumCheck(da.RunUser, da.AgentVersionDir, da.TarballFilename, da.Md5Filename)
}

func Md5sumCheck(runUser, workdir, tarfile, md5file string) error {
	var cmd *exec.Cmd
	var md5Actual string
	if "darwin" == runtime.GOOS {
		cmd = BuildCommand(runUser, "md5", "-q", path.Join(workdir, tarfile))
	} else {
		cmd = BuildCommand(runUser, "md5sum", path.Join(workdir, tarfile))
	}
	cmd.Dir = file.SelfDir()
	bs, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cd %s; md5sum -c %s fail", workdir, md5file)
	}
	strMd5file, _ := file.ToString(path.Join(workdir, md5file))
	if "" == strMd5file {
		return fmt.Errorf("md5file is empty")
	}

	if "darwin" == runtime.GOOS {
		md5Actual = strings.Replace(string(bs), "\n", "", -1)
	} else {
		md5Actual = strings.Fields(string(bs))[0]
	}

	md5Except := strings.Fields(strMd5file)[0]

	if md5Actual == md5Except {
		return nil
	}
	return fmt.Errorf("md5Actual:%s, md5Except:%s<<<===end", md5Actual, md5Except)
}

func FilesReady(da *model.DesiredAgent) bool {
	if !CheckFileExists(da.RunUser, da.Md5Filepath) {
		return false
	}

	if !CheckFileExists(da.RunUser, da.TarballFilepath) {
		return false
	}

	if !CheckFileExists(da.RunUser, da.ControlFilepath) {
		return false
	}
	return nil == Md5sumCheck(da.RunUser, da.AgentVersionDir, da.TarballFilename, da.Md5Filename)
}

func InsureDesiredAgentDirExists(da *model.DesiredAgent) error {
	err := InsureUserDir(da.AgentDir, da.RunUser, true)
	if err != nil {
		log.Println("insure dir", da.AgentDir, "fail", err)
		return err
	}

	err = InsureUserDir(da.AgentVersionDir, da.RunUser, false)
	if err != nil {
		log.Println("insure dir", da.AgentVersionDir, "fail", err)
	}
	return err
}

func InsureUserDir(fp, username string, createByRoot bool) error {
	var err error
	if CheckDirectoryExists(username, fp) {
		return nil
	}

	if CurrentUser == username {
		return os.MkdirAll(fp, os.ModePerm)
	} else if HasSudoPermission {
		if createByRoot {
			_, err = utils.ExecuteCommand(file.SelfDir(), fmt.Sprintf("sudo mkdir -p %s", fp))
			if nil != err {
				return err
			}
			_, err = utils.ExecuteCommand(file.SelfDir(), fmt.Sprintf("sudo chown -R %s:%s %s", username, username, fp))
		} else {
			_, err = utils.ExecuteCommand(file.SelfDir(), fmt.Sprintf("sudo -u %s mkdir -p %s", username, fp))
		}
	}
	return err
}

func InsureRunUserExists(da *model.DesiredAgent) error {
	if CurrentUser == da.RunUser { //ops-updater和 Agent运行用户一致
		return nil
	} else if HasSudoPermission {
		if utils.CheckUserExists(da.RunUser) {
			return nil
		}
		_, err := utils.ExecuteCommand(file.SelfDir(), fmt.Sprintf("sudo useradd %s", da.RunUser))
		return err
	}
	return fmt.Errorf("you donot have permission to insure user %s", da.RunUser)
}
