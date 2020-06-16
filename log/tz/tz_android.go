// +build android

package tz

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func init() {
	// Fix Termux lack of ZONEINFO by reuse android tzdada
	out := strings.TrimSpace(os.Getenv("TZ"));
	if out == "" {
		if z, err := exec.Command("/system/bin/getprop", "persist.sys.timezone").Output(); err == nil {
			if out = strings.TrimSpace(string(z)); out == "" {
				de("Empty android TZ Env")
			}
		} else {
			de("Error on getting %s, %v", "persist.sys.timezone", err)
		}
	}

	if out != "" {
		if loc, err := time.LoadLocation(out); err == nil {
			time.Local = loc
			de("TZ=%s", out)
		} else {
			de("LoadLocation failed for %s: %v", out, err)
		}
	}

	return
}

func de(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
