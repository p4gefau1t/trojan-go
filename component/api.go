//go:build api || full
// +build api full

package build

import (
	_ "github.com/p4gefau1t/trojan-go/api/control"
	_ "github.com/p4gefau1t/trojan-go/api/service"
)
