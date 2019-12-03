package main

import (
	"testing"
)

func TestParseStatusLine(t *testing.T) {
	type Output struct {
		key    string
		status string
		pid    string
		group  string
	}

	var tests = []struct {
		input    string
		success  bool
		expected Output
	}{
		{"KEY                       STATUS      PID GROUP", true, Output{"KEY", "STATUS", "PID", "GROUP"}},
		{"OPENIO-account-0          UP         1163 OPENIO,account,0", true, Output{"OPENIO-account-0", "UP", "1163", "OPENIO,account,0"}},
		{"A B C", false, Output{}},
		{"A B C D", true, Output{"A", "B", "C", "D"}},
		{"A B C D E", true, Output{"A", "B", "C", "D"}},
		{" A B C D", true, Output{"A", "B", "C", "D"}},
	}

	for _, test := range tests {
		key, status, pid, group, err := parseStatusLine(test.input)
		if err != nil {
			if test.success {
				t.Errorf("parseStatusLine(%q) => Error: %v", test.input, err)
			} else {
				// pass
			}
			continue
		}

		actual := Output{key, status, pid, group}
		if !test.success {
			t.Errorf("parseStatusLine(%q) => Expected: Error, Actual: %v", test.input, actual)
			continue
		}
		if actual != test.expected {
			t.Errorf("parseStatusLine(%q) => Expected: %v, Actual: %v", test.input, test.expected, actual)
			continue
		}
	}
}
