package cvss

func IsValidVersion(version string) bool {
	for _, v := range CvssVersions {
		if v == version {
			return true
		}
	}
	return false
}
