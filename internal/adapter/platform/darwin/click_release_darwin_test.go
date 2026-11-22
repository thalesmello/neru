//go:build darwin

package darwin

import (
	"errors"
	"testing"
)

var errClickRecovery = errors.New("injection failed")

func TestClickRelease_RetriesOnlyMissingSyntheticUps(t *testing.T) {
	tests := []struct {
		name            string
		releasesUntilUp int
		physicalAfter   int
		postFails       bool
		wantPosts       int
		wantErr         bool
	}{
		{name: "normal click", releasesUntilUp: 0, physicalAfter: -1},
		{name: "lost up recovered", releasesUntilUp: 1, physicalAfter: -1, wantPosts: 1},
		{name: "two lost ups recovered", releasesUntilUp: 2, physicalAfter: -1, wantPosts: 2},
		{
			name:            "bounded failure",
			releasesUntilUp: 4,
			physicalAfter:   -1,
			wantPosts:       3,
			wantErr:         true,
		},
		{name: "physical drag is untouched", releasesUntilUp: 4, physicalAfter: 0},
		{name: "physical press during retry", releasesUntilUp: 4, physicalAfter: 1, wantPosts: 1},
		{
			name:            "cannot create recovery event",
			releasesUntilUp: 1,
			physicalAfter:   -1,
			postFails:       true,
			wantPosts:       1,
			wantErr:         true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			posts, pauses, remembered := 0, 0, 0

			err := recoverClickRelease(clickReleaseOps{
				isDown:       func() bool { return posts < testCase.releasesUntilUp },
				physicalDown: func() bool { return testCase.physicalAfter >= 0 && posts >= testCase.physicalAfter },
				postUp: func() error {
					posts++

					if testCase.postFails {
						return errClickRecovery
					}

					return nil
				},
				pause:    func() { pauses++ },
				remember: func() { remembered++ },
			})
			if (err != nil) != testCase.wantErr || posts != testCase.wantPosts {
				t.Fatalf(
					"err=%v, up posts=%d; want error=%v, posts=%d",
					err,
					posts,
					testCase.wantErr,
					testCase.wantPosts,
				)
			}

			if remembered != 0 && !testCase.wantErr || remembered == 0 && testCase.wantErr {
				t.Fatalf(
					"unreleased button bookkeeping=%d, want error=%v",
					remembered,
					testCase.wantErr,
				)
			}

			wantPauses := posts
			if testCase.postFails {
				wantPauses--
			}

			if pauses != wantPauses {
				t.Fatalf("pauses=%d, want %d", pauses, wantPauses)
			}
		})
	}
}
