package ref

import (
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5/plumbing"
	"go.uber.org/zap"
)

func SortTags(tags []plumbing.ReferenceName, log *zap.SugaredLogger) {
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

		return parsedI.LessThan(parsedJ)
	})
}

func TagsFilter(tags []plumbing.ReferenceName, Constraints *semver.Constraints, log *zap.SugaredLogger) ([]plumbing.ReferenceName, error) {
	var result []plumbing.ReferenceName
	for _, t := range tags {
		shortTag := t.Short()

		parsedTag, err := semver.NewVersion(shortTag)
		if err != nil {
			log.With("tag", shortTag).Info("cannot parse semver for filtering. Ignoring tag.")
			continue
		}

		if Constraints.Check(parsedTag) {
			result = append(result, t)
		}
	}

	return result, nil
}
