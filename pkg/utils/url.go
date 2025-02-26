// From https://github.com/anacrolix/missinggo/blob/master/httptoo/url.go

package utils

import "net/url"

func CopyURL(u *url.URL) (ret *url.URL) {
	ret = new(url.URL)
	*ret = *u

	if u.User != nil {
		ret.User = new(url.Userinfo)
		*ret.User = *u.User
	}
	return
}
