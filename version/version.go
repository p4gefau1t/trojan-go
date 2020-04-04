package version

import (
	"flag"
	"fmt"

	"github.com/p4gefau1t/trojan-go/common"
)

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
		fmt.Println("Trojan-Go", common.Version)
		fmt.Println("Developed by PageFault(p4gefau1t)")
		fmt.Println("Lisensed under GNU General Public License v3")
		fmt.Println("GitHub Repository: https://github.com/p4gefau1t/trojan-go")
		return nil
	}
	return common.NewError("not set")
}

func init() {
	common.RegisterOptionHandler(&versionOption{
		arg: flag.Bool("version", false, "Display version and help info"),
	})
}
