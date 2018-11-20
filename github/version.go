package github

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tcnksm/go-latest"
)

// CheckLatestVersion checks to see if we're on the latest version
func CheckLatestVersion(repoOwner, repoName, currentVersion string) error {
	githubTag := &latest.GithubTag{
		Owner:             repoOwner,
		Repository:        repoName,
		FixVersionStrFunc: latest.DeleteFrontV(),
	}

	res, err := latest.Check(githubTag, currentVersion)
	if err != nil {
		return errors.Wrap(err, "Could not fetch release information from github")
	}

	// Latest version
	if !res.Outdated {
		logrus.WithField("current_version", currentVersion).Debug("Already at latest version")
		return nil
	}
	logrus.WithField("current_version", currentVersion).Warnf("%s is not the current version, you sould upgrade to %s", version, res.Current)
	return nil
}
