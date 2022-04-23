package main

import "testing"

func TestComposeGreeting(t *testing.T) {
	cases := []struct{ name, expected string }{
		{"", "Hello Stranger!"}, {"Mary", "Hello Mary!"},
	}

	for _, c := range cases {
		result := ComposeGreeting(c.name)
		if result != c.expected {
			t.Fatalf("want %s, got %s\n", c.expected, result)
		}
	}
}
