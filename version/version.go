package version

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/p4gefau1t/trojan-go/common"
)

type versionOption struct {
	flag *bool
}

func (*versionOption) Name() string {
	return "help"
}

func (*versionOption) Priority() int {
	return 10
}

func (c *versionOption) Handle() error {
	if *c.flag {
		fmt.Println("Trojan-Go", common.Version, fmt.Sprintf("(%s %s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH))
		fmt.Println("Developed by PageFault (p4gefau1t)")
		fmt.Println("Lisensed under GNU General Public License v3")
		fmt.Println("GitHub Repository: https://github.com/p4gefau1t/trojan-go")
		fmt.Println("Documents: https://p4gefau1t.github.io/trojan-go/")
		return nil
	}
	return common.NewError("not set")
}

func init() {
	common.RegisterOptionHandler(&versionOption{
		flag: flag.Bool("version", false, "Display version and help info"),
	})
}
