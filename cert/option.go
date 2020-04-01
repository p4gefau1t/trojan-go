package cert

import (
	"flag"

	"github.com/p4gefau1t/trojan-go/common"
)

type certOption struct {
	args *string
	common.OptionHandler
}

func (*certOption) Name() string {
	return "cert"
}

func (*certOption) Priority() int {
	return 10
}

func (c *certOption) Handle() error {
	switch *c.args {
	case "request":
		RequestCertGuide()
		return nil
	case "renew":
		RenewCertGuide()
		return nil
	case "INVALID":
		return common.NewError("not specified")
	default:
		err := common.NewError("invalid args " + *c.args)
		logger.Error(err)
		return common.NewError("invalid args")
	}
}

func init() {
	common.RegisterOptionHandler(&certOption{
		args: flag.String("cert", "INVALID", "Simple letsencrpyt cert acme client. Use \"-cert request\" to request a cert or \"-cert renew\" to renew a cert"),
	})
}
