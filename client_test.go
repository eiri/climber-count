package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

// Mocked HTTP Client
type MockClient struct {
	resp *http.Response
	err  error
	req  *http.Request
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	m.req = req // Save the request for inspection
	return m.resp, m.err
}

func TestExtract(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Basic input",
			input: `
				var data = {
					'SBG' : {
				'capacity' : 60,
				'count' : 8,
				'subLabel' : 'Current climber count',
				'lastUpdate' : 'Last updated:&nbspnow  (10:16 AM)'
			},'SBL' : {
				'capacity' : 100,
				'count' : 3,
				'subLabel' : 'Current climber count',
				'lastUpdate' : 'Last updated:&nbsp8 mins ago (10:09 AM)'
			},    };`,
			expected: `{"SBG":{"capacity":60,"count":8,"subLabel":"Currentclimbercount","lastUpdate":"Lastupdated:&nbspnow(10:16AM)"},"SBL":{"capacity":100,"count":3,"subLabel":"Currentclimbercount","lastUpdate":"Lastupdated:&nbsp8minsago(10:09AM)"}}`,
		},
		{
			name:     "Empty input",
			input:    "",
			expected: "",
		},
		{
			name: "Input without equal sign",
			input: `
				var data {
					'SBG' : {
				'capacity' : 60,
				'count' : 8,
				'subLabel' : 'Current climber count',
				'lastUpdate' : 'Last updated:&nbspnow  (10:16 AM)'
			},'SBL' : {
				'capacity' : 100,
				'count' : 3,
				'subLabel' : 'Current climber count',
				'lastUpdate' : 'Last updated:&nbsp8 mins ago (10:09 AM)'
			},    };`,
			expected: "",
		},
		{
			name: "Input with only semicolon",
			input: `
				;`,
			expected: "",
		},
		{
			name: "Input with unicode spaces",
			input: `
				var data = {
					'SBG' : {
				'capacity' : 60,
				'count' : 8,
				'subLabel' : 'Current climber count',
				'lastUpdate' : 'Last updated:&nbspnow  (10:16 AM)'
			},'SBL' : {
				'capacity' : 100,
				'count' : 3,
				'subLabel' : 'Current climber count',
				'lastUpdate' : 'Last updated:&nbsp8 mins ago (10:09 AM)'
			},    };`,
			expected: `{"SBG":{"capacity":60,"count":8,"subLabel":"Currentclimbercount","lastUpdate":"Lastupdated:&nbspnow(10:16AM)"},"SBL":{"capacity":100,"count":3,"subLabel":"Currentclimbercount","lastUpdate":"Lastupdated:&nbsp8minsago(10:09AM)"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extract(tc.input)
			if result != tc.expected {
				t.Errorf("expected %q but got %q", tc.expected, result)
			}
		})
	}
}

func TestParse(t *testing.T) {
	htmlContent := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Test Page</title>
	</head>
	<body>
		<div>Some content here</div>
		<script>
			var data = {
				'SBG' : {
			'capacity' : 60,
			'count' : 8,
			'subLabel' : 'Current climber count',
			'lastUpdate' : 'Last updated:&nbspnow  (10:16 AM)'
		},'SBL' : {
			'capacity' : 100,
			'count' : 3,
			'subLabel' : 'Current climber count',
			'lastUpdate' : 'Last updated:&nbsp8 mins ago (10:09 AM)'
		},    };

		function DoSomething(p1, p2) {
  			return p1 * p2;
     	}
		</script>
	</body>
	</html>
	`

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	result := parse(doc)
	expected := `{"SBG":{"capacity":60,"count":8,"subLabel":"Currentclimbercount","lastUpdate":"Lastupdated:&nbspnow(10:16AM)"},"SBL":{"capacity":100,"count":3,"subLabel":"Currentclimbercount","lastUpdate":"Lastupdated:&nbsp8minsago(10:09AM)"}}`

	if result != expected {
		t.Errorf("expected %q but got %q", expected, result)
	}
}

func TestFetch(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  *http.Response
		mockError     error
		expectedError string
		expectedBody  string
	}{
		{
			name: "Successful fetch",
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("Success")),
			},
			mockError:     nil,
			expectedError: "",
			expectedBody:  "Success",
		},
		{
			name: "HTTP error response",
			mockResponse: &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("Not Found")),
			},
			mockError:     nil,
			expectedError: "failed to fetch URL: , status: 404",
			expectedBody:  "",
		},
		{
			name:          "Request error",
			mockResponse:  nil,
			mockError:     fmt.Errorf("request error"),
			expectedError: "request error",
			expectedBody:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClient{
				resp: tt.mockResponse,
				err:  tt.mockError,
			}

			client := NewClient("uid", "fid")
			client.client = mockClient

			body, err := client.fetch()

			if tt.expectedError != "" {
				if err == nil || err.Error() != tt.expectedError {
					t.Errorf("expected error %q but got %q", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if body != nil {
				bodyBytes, _ := io.ReadAll(body)
				bodyStr := string(bodyBytes)
				if bodyStr != tt.expectedBody {
					t.Errorf("expected body %q but got %q", tt.expectedBody, bodyStr)
				}
				body.Close()
			}

			// Check if the User-Agent header is set correctly
			if userAgent := mockClient.req.Header.Get("User-Agent"); userAgent != USER_AGENT {
				t.Errorf("expected User-Agent %q but got %q", USER_AGENT, userAgent)
			}
		})
	}
}
