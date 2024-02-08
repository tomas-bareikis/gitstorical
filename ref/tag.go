package ref

import (
	"sort"

	"github.com/coreos/go-semver/semver"
	"github.com/go-git/go-git/v5/plumbing"
	"go.uber.org/zap"
)

func SortTags(tags []plumbing.ReferenceName, log *zap.SugaredLogger) error {
	sort.Slice(tags, func(i, j int) bool {
		semverI := tags[i].Short()
		semverJ := tags[j].Short()

		parsedI, err := semver.NewVersion(semverI)
		if err != nil {
			log.With("tag", semverI).Warn("cannot parse semver for sorting")
			return false
		}

		parsedJ, err := semver.NewVersion(semverJ)
		if err != nil {
			log.With("tag", semverJ).Warn("cannot parse semver for sorting")
			return false
		}

		return parsedI.LessThan(*parsedJ)
	})

	return nil
}
