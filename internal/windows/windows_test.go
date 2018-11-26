package windows_test

import (
	"testing"

	"bitbucket.org/rhagenson/swsc/internal/windows"
)

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
		{5786, 100, 15056328},
	}
	for _, tc := range tt {
		got, _ := windows.GenerateWindows(tc.length, tc.minWin)
		if len(got) != tc.expected {
			t.Errorf("Given len:%d, win:%d, expected %d, got %d\n",
				tc.length, tc.minWin, tc.expected, len(got))
		}
	}
}

func TestWindow(t *testing.T) {
	tt := []struct {
		win *windows.Window
	}{
		{windows.New(0, 0)},    // Start == Stop
		{windows.New(0, 50)},   // 0 <= Start < Stop
		{windows.New(50, 0)},   //  0 <= Stop < Start, orients values such as Start < Stop
		{windows.New(-50, 0)},  // Start < Stop <= 0
		{windows.New(0, -50)},  // Stop < Start <= 0, orients values such as Start < Stop
		{windows.New(-50, 50)}, // Start < 0 < Stop
		{windows.New(50, -50)}, // Stop < 0 < Start, orients values such as Start < Stop
	}

	for _, tc := range tt {
		t.Run("Start", func(t *testing.T) {
			if tc.win.Start() != tc.win[0] {
				t.Errorf("Got: %d, Expected %d", tc.win.Start(), tc.win[0])
			}
		})
		t.Run("Stop", func(t *testing.T) {
			if tc.win.Stop() != tc.win[1] {
				t.Errorf("Got: %d, Expected %d", tc.win.Start(), tc.win[0])
			}
		})
	}
}
