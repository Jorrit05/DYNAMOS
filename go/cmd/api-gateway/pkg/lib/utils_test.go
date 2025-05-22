package lib

import (
	"reflect"
	"sort"
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

func TestSliceIntersectAndDifference(t *testing.T) {
	testCases := []struct {
		sliceA             []string
		sliceB             []string
		expectedMatched    []string
		expectedNotMatched []string
	}{
		{
			sliceA:             []string{"apple", "banana", "cherry", "apple", "grape"},
			sliceB:             []string{"banana", "cherry", "kiwi", "mango"},
			expectedMatched:    []string{"banana", "cherry"},
			expectedNotMatched: []string{"apple", "grape"},
		},
		{
			sliceA:             []string{"apple", "orange", "grape"},
			sliceB:             []string{"banana", "orange", "kiwi", "mango"},
			expectedMatched:    []string{"orange"},
			expectedNotMatched: []string{"apple", "grape"},
		},
	}

	for _, testCase := range testCases {
		matched, notMatched := SliceIntersectAndDifference(testCase.sliceA, testCase.sliceB)
		sort.Strings(matched)
		sort.Strings(notMatched)
		sort.Strings(testCase.expectedMatched)
		sort.Strings(testCase.expectedNotMatched)
		if !reflect.DeepEqual(matched, testCase.expectedMatched) {
			t.Errorf("expected matched %v, got %v", testCase.expectedMatched, matched)
		}
		if !reflect.DeepEqual(notMatched, testCase.expectedNotMatched) {
			t.Errorf("expected notMatched %v, got %v", testCase.expectedNotMatched, notMatched)
		}
	}
}

func TestCreateMapFromSlice(t *testing.T) {
	slice := []string{"unl1_agent", "unl2_agent", "unl3"}
	expected := map[string]bool{"unl1_agent": true, "unl2_agent": true, "unl3": true}

	result := createMapFromSlice(slice)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestGetMatchedElements(t *testing.T) {
	sliceA := []string{"unl1_agent", "unl2_agent", "unl3"}
	sliceB := []string{"unl2_agent", "unl5"}
	expected := []string{"unl2_agent"}

	result, _ := GetMatchedElements(sliceA, sliceB)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestGetNotMatchedElements(t *testing.T) {
	mapA := map[string]bool{"unl1_agent": true, "unl3": true}
	expected := []string{"unl1_agent", "unl3"}

	result := GetNotMatchedElements(mapA)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestSliceIntersectAndDifference2(t *testing.T) {
	sliceA := []string{"unl1_agent", "unl2_agent", "unl3"}
	sliceB := []string{"unl2_agent", "unl5"}
	expectedMatched := []string{"unl2_agent"}
	expectedNotMatched := []string{"unl1_agent", "unl3"}

	matched, notMatched := SliceIntersectAndDifference(sliceA, sliceB)

	if !reflect.DeepEqual(matched, expectedMatched) {
		t.Errorf("Expected %v, got %v", expectedMatched, matched)
	}

	if !reflect.DeepEqual(notMatched, expectedNotMatched) {
		t.Errorf("Expected %v, got %v", expectedNotMatched, notMatched)
	}
}

func TestSliceIntersectAndDifferenceEmptySlice(t *testing.T) {
	sliceA := []string{}
	sliceB := []string{"unl2_agent", "unl5"}
	expectedMatched := []string{}
	expectedNotMatched := []string{}

	matched, notMatched := SliceIntersectAndDifference(sliceA, sliceB)

	if !reflect.DeepEqual(matched, expectedMatched) {
		t.Errorf("Expected %v, got %v", expectedMatched, matched)
	}

	if !reflect.DeepEqual(notMatched, expectedNotMatched) {
		t.Errorf("Expected %v, got %v", expectedNotMatched, notMatched)
	}
}

func TestSliceIntersectAndDifferenceNoIntersection(t *testing.T) {
	sliceA := []string{"unl1_agent", "unl3"}
	sliceB := []string{"unl2_agent", "unl5"}
	expectedMatched := []string{}
	expectedNotMatched := []string{"unl1_agent", "unl3"}

	matched, notMatched := SliceIntersectAndDifference(sliceA, sliceB)

	if !reflect.DeepEqual(matched, expectedMatched) {
		t.Errorf("Expected %v, got %v", expectedMatched, matched)
	}

	if !reflect.DeepEqual(notMatched, expectedNotMatched) {
		t.Errorf("Expected %v, got %v", expectedNotMatched, notMatched)
	}
}
