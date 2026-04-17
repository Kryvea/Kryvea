package util

import "net/url"

func JoinUrlPath(base string, elem ...string) string {
	path, err := url.JoinPath(base, elem...)
	if err != nil {
		url.JoinPath("/", elem...)
	}
	return path
}
