package cron

import (
	"os"

	"github.com/coraldane/ops-common/utils"
	"github.com/coraldane/ops-updater/g"
)

var (
	HasSudoPermission = false
	CurrentUser       = ""
)

func init() {
	_, err := utils.ExecuteCommand(g.SelfDir, "sudo -n true")
	if nil == err {
		HasSudoPermission = true
	}

	CurrentUser = utils.GetUserByPid(os.Getpid())
}
