package g

import (
	"log"
	"runtime"
)

//change history
// tianligen 0.0.2  2016-02-03 增加对不同操作系统内核(OSX, 32位)机器的支持
// tianligen 0.0.3  2016-10-19 增加对于新版本ops-meta的支持
const (
	VERSION = "0.0.3"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
