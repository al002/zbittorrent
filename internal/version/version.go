package version

import (
	"fmt"
)

var (
	DefaultHttpUserAgent string
)

func init() {
	const (
		namespace   = "al002"
		packageName = "zbittorrent"
	)
	var (
		version = "unknown"
	)

	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/User-Agent#library_and_net_tool_ua_strings
	DefaultHttpUserAgent = fmt.Sprintf(
		"%v-%v/%v",
		namespace,
		packageName,
		version,
	)
}
