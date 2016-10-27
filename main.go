package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/toolkits/logger"
	"github.com/toolkits/sys"

	"github.com/coraldane/ops-updater/cron"
	"github.com/coraldane/ops-updater/g"
	"github.com/coraldane/ops-updater/http"
)

func main() {
	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")
	flag.Parse()

	if *version {
		fmt.Println(g.VERSION)
		os.Exit(0)
	}

	if err := g.ParseConfig(*cfg); err != nil {
		log.Fatalln(err)
	}

	g.InitGlobalVariables()
	logger.SetLevelWithDefault(g.Config().LogLevel, "info")

	CheckDependency()

	go http.Start()
	go cron.Heartbeat()

	select {}
}

func CheckDependency() {
	_, err := sys.CmdOut("wget", "--help")
	if err != nil {
		log.Fatalln("dependency wget not found")
	}

	if "darwin" == runtime.GOOS {
		_, err = sys.CmdOut("md5", "/etc/hosts")
		if err != nil {
			log.Fatalln("dependency md5 not found")
		}
	} else {
		_, err = sys.CmdOut("md5sum", "--help")
		if err != nil {
			log.Fatalln("dependency md5sum not found")
		}
	}

	_, err = sys.CmdOut("tar", "--help")
	if err != nil {
		log.Fatalln("dependency tar not found")
	}
}
