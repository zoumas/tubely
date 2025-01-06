package main

import (
	"fmt"
	"testing"
)

func TestGetVideoAspectRatio(t *testing.T) {
	cases := []struct {
		filename    string
		aspectRatio string
	}{
		{
			filename:    "./samples/boots-video-horizontal.mp4",
			aspectRatio: horizontalAspectRatio,
		},
		{
			filename:    "./samples/boots-video-vertical.mp4",
			aspectRatio: verticalAspectRation,
		},
		{
			filename:    "./samples/4-3.mp4",
			aspectRatio: "other",
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			actual, err := getVideoAspectRatio(tc.filename)
			if err != nil {
				t.Errorf("unexpected error %v", err)
				return
			}
			if tc.aspectRatio != actual {
				t.Errorf("expected: %q\nactual: %q\n", tc.aspectRatio, actual)
			}
		})
	}
}
