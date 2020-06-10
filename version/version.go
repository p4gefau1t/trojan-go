package version

import (
	"flag"
	"fmt"
	"github.com/p4gefau1t/trojan-go/option"
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
		fmt.Println("Licensed under GNU General Public License version 3")
		fmt.Println("GitHub Repository:\thttps://github.com/p4gefau1t/trojan-go")
		fmt.Println("Trojan-Go Documents:\thttps://p4gefau1t.github.io/trojan-go/")
		return nil
	}
	return common.NewError("not set")
}

func init() {
	option.RegisterHandler(&versionOption{
		flag: flag.Bool("version", false, "Display version and help info"),
	})
}
