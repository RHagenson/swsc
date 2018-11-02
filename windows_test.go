package main

import "testing"

func TestGenerateWindows(t *testing.T) {
	tt := []struct {
		length   int
		minWin   int
		expected int
	}{
		{320, 100, 231},
		{321, 100, 253},
		{322, 100, 276},
		{323, 100, 300},
		{324, 100, 325},
		{325, 100, 351},

		{325, 101, 276},
		{325, 102, 210},
		{325, 103, 153},
		{325, 104, 105},
		{325, 105, 66},

		{5786, 50, 15890703},
	}
	for _, tc := range tt {
		got, _ := generateWindows(tc.length, tc.minWin)
		if len(got) != tc.expected {
			t.Errorf("Given len:%d, win:%d, expected %d, got %d\n",
				tc.length, tc.minWin, tc.expected, len(got))
		}
	}
}
