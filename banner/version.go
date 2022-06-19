package banner

import (
	"fmt"
)

var (
	versionString   = ""
	versionGitSHA   = ""
	buildTimestamp  = ""
	goVersionString = ""
)

func Version() string {
	return fmt.Sprintf("%s (%v, %v, go %v)", versionString, versionGitSHA, buildTimestamp, goVersionString)
}
