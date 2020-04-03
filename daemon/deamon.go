package deamon

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/p4gefau1t/trojan-go/common"
)

type DaemonOption struct {
	daemon *bool
	common.OptionHandler
}

func (*DaemonOption) Name() string {
	return "daemon"
}

func (*DaemonOption) Priority() int {
	return 1000
}

func (o *DaemonOption) Handle() error {
	if !*o.daemon {
		return common.NewError("not set")
	}
	args := os.Args[1:]
	i := 0
	for ; i < len(args); i++ {
		if strings.Contains(args[i], "-daemon") {
			args[i] = "-daemon=false"
		}
	}
	cmd := exec.Command(os.Args[0], args...)
	cmd.Start()
	fmt.Println("Trojan-Go is running in the background")
	fmt.Println("[PID]", cmd.Process.Pid)
	os.Exit(0)
	return nil
}

func init() {
	common.RegisterOptionHandler(&DaemonOption{
		daemon: flag.Bool("daemon", false, "run trojan-go as a daemon with -daemon"),
	})
}
