package version

import (
	"flag"
	"os"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
)

var logger = log.New(os.Stdout)

type versionOption struct {
	arg *bool
	common.OptionHandler
}

func (*versionOption) Name() string {
	return "help"
}

func (*versionOption) Priority() int {
	return 0
}

func (c *versionOption) Handle() error {
	if *c.arg {
		logger.Info("Trojan-Go", common.Version)
		logger.Info("Developed by PageFault(p4gefau1t)")
		logger.Info("Lisensed under GNU General Public License v3")
		logger.Info("GitHub Repository: https://github.com/p4gefau1t/trojan-go")
		return nil
	}
	return common.NewError("not set")
}

func init() {
	common.RegisterOptionHandler(&versionOption{
		arg: flag.Bool("version", false, "Display version and help info"),
	})
}
