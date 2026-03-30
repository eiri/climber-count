package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

type MockClient struct {
	resp *http.Response
	err  error
	req  *http.Request
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	m.req = req
	return m.resp, m.err
}

// roundTripFunc lets a plain function satisfy http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// minimalOccupancyHTML returns a minimal HTML page whose <script> block
// contains an occupancy data object for the given gym key.
func minimalOccupancyHTML(key string, count, capacity int) string {
	return fmt.Sprintf(`<!DOCTYPE html><html><body><script>
var data = {
  '%s' : {
    'capacity' : %d,
    'count' : %d,
    'subLabel' : 'Current climber count',
    'lastUpdate' : 'Last updated:&nbspnow  (10:00 AM)'
  },    };
</script></body></html>`, key, capacity, count)
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

			cfg := &Config{
				PGK: "pgk",
				FID: "fid",
			}
			client := NewClient(cfg)
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
				bodyBytes, err := io.ReadAll(body)
				if err != nil {
					t.Errorf("unexpected error on reading a body: %q", err)
				}
				bodyStr := string(bodyBytes)
				if bodyStr != tt.expectedBody {
					t.Errorf("expected body %q but got %q", tt.expectedBody, bodyStr)
				}
				err = body.Close()
				if err != nil {
					t.Errorf("unexpected error on closing a body: %q", err)
				}
			}

			// Check if the User-Agent header is set correctly
			if userAgent := mockClient.req.Header.Get("User-Agent"); userAgent != USER_AGENT {
				t.Errorf("expected User-Agent %q but got %q", USER_AGENT, userAgent)
			}
		})
	}
}

func TestClientRoundTripper_RoundTrip(t *testing.T) {
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })

	http.DefaultTransport = roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("ok")),
		}, nil
	})

	crt := ClientRoundTripper{}
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	resp, err := crt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestClientRoundTripper_RoundTrip_Error(t *testing.T) {
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })

	http.DefaultTransport = roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return nil, errors.New("dial error")
	})

	crt := ClientRoundTripper{}
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	_, err := crt.RoundTrip(req)
	if err == nil {
		t.Fatal("expected error from RoundTrip, got nil")
	}
}

func TestCounters_Success(t *testing.T) {
	page := minimalOccupancyHTML("TST", 5, 50)

	cfg := &Config{PGK: "pgk", FID: "fid"}
	c := NewClient(cfg)
	c.client = &MockClient{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(page)),
		},
	}

	counters, err := c.Counters()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if counters == nil {
		t.Fatal("expected non-nil Counters")
	}
}

func TestCounters_FetchError(t *testing.T) {
	cfg := &Config{PGK: "pgk", FID: "fid"}
	c := NewClient(cfg)
	c.client = &MockClient{err: errors.New("network error")}

	_, err := c.Counters()
	if err == nil {
		t.Fatal("expected error when fetch fails")
	}
}

func TestCounters_NoOccupancyData(t *testing.T) {
	// Valid HTTP response but no occupancy script block — parse() returns ""
	cfg := &Config{PGK: "pgk", FID: "fid"}
	c := NewClient(cfg)
	c.client = &MockClient{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("<html><body>no data</body></html>")),
		},
	}

	_, err := c.Counters()
	if err == nil {
		t.Fatal("expected error when page contains no occupancy data")
	}
}
