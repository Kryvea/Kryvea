package util

import "net/url"

func JoinUrlPath(base string, elem ...string) string {
	path, err := url.JoinPath(base, elem...)
	if err != nil {
		// Fall back to joining from the root if the base is invalid
		path, err = url.JoinPath("/", elem...)
		if err != nil {
			return "/"
		}
	}
	return path
}
