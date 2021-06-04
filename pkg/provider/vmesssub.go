package provider

import (
	"strings"

	"github.com/Sansui233/proxypool/pkg/tool"
)

type VmessSub struct {
	Base
}

func (sub VmessSub) Provide() string {
	sub.Types = "vmess"
	sub.preFilter()
	var resultBuilder strings.Builder
	for _, p := range *sub.Proxies {
		resultBuilder.WriteString(p.Link() + "\n")
	}
	return tool.Base64EncodeString(resultBuilder.String(), false)
}
