package lib

import (
	"strings"
	"testing"

	"gotest.tools/assert"
)

func TestGenerateGuid(t *testing.T) {
	testCases := []struct {
		name     string
		parts    int
		expected int
	}{
		{
			name:     "TestFullUUID",
			parts:    0,
			expected: 4,
		},
		{
			name:     "TestOnePart",
			parts:    1,
			expected: 0,
		},
		{
			name:     "TestTwoParts",
			parts:    2,
			expected: 1,
		},
		{
			name:     "TestThreeParts",
			parts:    3,
			expected: 2,
		},
		{
			name:     "TestOutOfRange",
			parts:    5,
			expected: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			guid := GenerateGuid(tc.parts)
			dashCount := strings.Count(guid, "-")

			if dashCount != tc.expected {
				t.Errorf("Expected %d dashes but got %d for parts = %d", tc.expected, dashCount, tc.parts)
			}
		})
	}
}

func TestLastPartAfterSlash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://example.com/path/to/resource", "resource"},
		{"example.com/some/path", "path"},
		{"no/slashes/here", "here"},
		{"no_slashes", "no_slashes"},
		{"trailing/slash/", ""},
		{"/leading/slash", "slash"},
		{"", ""},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			result := LastPartAfterSlash(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestSplitImageAndTag(t *testing.T) {
	testCases := []struct {
		fullImageName string
		expectedImage string
		expectedTag   string
	}{
		{"anonymize_service:latest", "anonymize_service", "latest"},
		{"anonymize_service", "anonymize_service", "latest"},
		{"anonymize_service:v1.0.0", "anonymize_service", "v1.0.0"},
		{"anonymize_service:1.0", "anonymize_service", "1.0"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.fullImageName, func(t *testing.T) {
			image, tag := SplitImageAndTag(testCase.fullImageName)

			if image != testCase.expectedImage {
				t.Errorf("Expected image '%s', got '%s'", testCase.expectedImage, image)
			}

			if tag != testCase.expectedTag {
				t.Errorf("Expected tag '%s', got '%s'", testCase.expectedTag, tag)
			}
		})
	}
}
