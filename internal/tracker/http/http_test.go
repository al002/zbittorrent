package httpTracker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"testing"

	trackerTypes "github.com/al002/zbittorrent/internal/tracker/types"
	"github.com/al002/zbittorrent/pkg/metainfo"
	"github.com/al002/zbittorrent/pkg/types"
)

// Mock HTTP transport for testing
type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

// Test setAnnounceParams function
func TestSetAnnounceParams(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		req      trackerTypes.AnnounceRequest
		expected string
	}{
		{
			name: "Basic parameters",
			url:  "http://tracker.example.com/announce",
			req: trackerTypes.AnnounceRequest{
				InfoHash:   metainfo.Hash{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				PeerID:     types.PeerID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				IP:         net.ParseIP("192.168.1.1"),
				Port:       6881,
				Uploaded:   1024,
				Downloaded: 512,
				Left:       2048,
				Event:      trackerTypes.AnnounceEventEmpty,
				Key:        12345,
				NumWant:    50,
			},
			expected: "http://tracker.example.com/announce?compact=1&downloaded=512&info_hash=%01%02%03%04%05%06%07%08%09%0A%0B%0C%0D%0E%0F%10%11%12%13%14&ip=192.168.1.1&key=12345&left=2048&peer_id=%01%02%03%04%05%06%07%08%09%0A%0B%0C%0D%0E%0F%10%11%12%13%14&port=6881&supportcrypto=1&uploaded=1024",
		},
		{
			name: "With started event",
			url:  "http://tracker.example.com/announce",
			req: trackerTypes.AnnounceRequest{
				InfoHash:   metainfo.Hash{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				PeerID:     types.PeerID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				IP:         net.ParseIP("192.168.1.1"),
				Port:       6881,
				Uploaded:   0,
				Downloaded: 0,
				Left:       10240,
				Event:      trackerTypes.AnnounceEventStarted,
				Key:        12345,
				NumWant:    50,
			},
			expected: "http://tracker.example.com/announce?compact=1&downloaded=0&event=started&info_hash=%01%02%03%04%05%06%07%08%09%0A%0B%0C%0D%0E%0F%10%11%12%13%14&ip=192.168.1.1&key=12345&left=10240&peer_id=%01%02%03%04%05%06%07%08%09%0A%0B%0C%0D%0E%0F%10%11%12%13%14&port=6881&supportcrypto=1&uploaded=0",
		},
		{
			name: "With negative left value",
			url:  "http://tracker.example.com/announce",
			req: trackerTypes.AnnounceRequest{
				InfoHash:   metainfo.Hash{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				PeerID:     types.PeerID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				IP:         net.ParseIP("192.168.1.1"),
				Port:       6881,
				Uploaded:   1024,
				Downloaded: 512,
				Left:       -1,
				Event:      trackerTypes.AnnounceEventEmpty,
				Key:        12345,
				NumWant:    50,
			},
			expected: "http://tracker.example.com/announce?compact=1&downloaded=512&info_hash=%01%02%03%04%05%06%07%08%09%0A%0B%0C%0D%0E%0F%10%11%12%13%14&ip=192.168.1.1&key=12345&left=9223372036854775807&peer_id=%01%02%03%04%05%06%07%08%09%0A%0B%0C%0D%0E%0F%10%11%12%13%14&port=6881&supportcrypto=1&uploaded=1024",
		},
		{
			name: "With existing query parameters",
			url:  "http://tracker.example.com/announce?test=value",
			req: trackerTypes.AnnounceRequest{
				InfoHash:   metainfo.Hash{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				PeerID:     types.PeerID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				IP:         net.ParseIP("192.168.1.1"),
				Port:       6881,
				Uploaded:   1024,
				Downloaded: 512,
				Left:       2048,
				Event:      trackerTypes.AnnounceEventEmpty,
				Key:        12345,
				NumWant:    50,
			},
			expected: "http://tracker.example.com/announce?test=value&compact=1&downloaded=512&info_hash=%01%02%03%04%05%06%07%08%09%0A%0B%0C%0D%0E%0F%10%11%12%13%14&ip=192.168.1.1&key=12345&left=2048&peer_id=%01%02%03%04%05%06%07%08%09%0A%0B%0C%0D%0E%0F%10%11%12%13%14&port=6881&supportcrypto=1&uploaded=1024",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedURL, err := url.Parse(tt.url)
			if err != nil {
				t.Fatalf("Failed to parse URL: %v", err)
			}

			setAnnounceParams(parsedURL, &tt.req)

			if parsedURL.String() != tt.expected {
				t.Errorf("setAnnounceParams() got = %v, want %v", parsedURL.String(), tt.expected)
			}
		})
	}
}

// Test Announce method
func TestAnnounce(t *testing.T) {
	tests := []struct {
		name           string
		req            trackerTypes.AnnounceRequest
		opt            AnnounceOpt
		mockResponse   *http.Response
		mockError      error
		expectedResult AnnounceResponse
		expectedError  bool
	}{
		{
			name: "Successful announce with compact peers",
			req: trackerTypes.AnnounceRequest{
				InfoHash:   metainfo.Hash{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				PeerID:     types.PeerID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				IP:         net.ParseIP("192.168.1.1"),
				Port:       6881,
				Uploaded:   1024,
				Downloaded: 512,
				Left:       2048,
				Event:      trackerTypes.AnnounceEventEmpty,
				Key:        12345,
				NumWant:    50,
			},
			opt: AnnounceOpt{
				UserAgent:  "TestClient/1.0",
				HostHeader: "tracker.example.com",
			},
			mockResponse: &http.Response{
				StatusCode: 200,
				Body: io.NopCloser(bytes.NewBufferString(
					"d8:completei10e10:incompletei5e8:intervali1800e5:peers6:\x01\x02\x03\x04\x1a\x0ae")),
			},
			mockError: nil,
			expectedResult: AnnounceResponse{
				Interval: 1800,
				Seeders:  10,
				Leechers: 5,
				Peers: []Peer{
					{
						IP:   net.IP{1, 2, 3, 4},
						Port: 6666,
					},
				},
			},
			expectedError: false,
		},
		{
			name: "HTTP request error",
			req: trackerTypes.AnnounceRequest{
				InfoHash:   metainfo.Hash{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				PeerID:     types.PeerID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				IP:         net.ParseIP("192.168.1.1"),
				Port:       6881,
				Uploaded:   1024,
				Downloaded: 512,
				Left:       2048,
				Event:      trackerTypes.AnnounceEventEmpty,
				Key:        12345,
				NumWant:    50,
			},
			opt: AnnounceOpt{
				UserAgent:  "TestClient/1.0",
				HostHeader: "tracker.example.com",
			},
			mockResponse:   nil,
			mockError:      fmt.Errorf("connection refused"),
			expectedResult: AnnounceResponse{},
			expectedError:  true,
		},
		{
			name: "Non-200 response",
			req: trackerTypes.AnnounceRequest{
				InfoHash:   metainfo.Hash{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				PeerID:     types.PeerID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				IP:         net.ParseIP("192.168.1.1"),
				Port:       6881,
				Uploaded:   1024,
				Downloaded: 512,
				Left:       2048,
				Event:      trackerTypes.AnnounceEventEmpty,
				Key:        12345,
				NumWant:    50,
			},
			opt: AnnounceOpt{
				UserAgent:  "TestClient/1.0",
				HostHeader: "tracker.example.com",
			},
			mockResponse: &http.Response{
				StatusCode: 404,
				Status:     "404 Not Found",
				Body:       io.NopCloser(bytes.NewBufferString("Not Found")),
			},
			mockError:      nil,
			expectedResult: AnnounceResponse{},
			expectedError:  true,
		},
		{
			name: "Response with failure reason",
			req: trackerTypes.AnnounceRequest{
				InfoHash:   metainfo.Hash{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				PeerID:     types.PeerID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				IP:         net.ParseIP("192.168.1.1"),
				Port:       6881,
				Uploaded:   1024,
				Downloaded: 512,
				Left:       2048,
				Event:      trackerTypes.AnnounceEventEmpty,
				Key:        12345,
				NumWant:    50,
			},
			opt: AnnounceOpt{
				UserAgent:  "TestClient/1.0",
				HostHeader: "tracker.example.com",
			},
			mockResponse: &http.Response{
				StatusCode: 200,
				Body: io.NopCloser(bytes.NewBufferString(
					"d14:failure reason23:Invalid info_hash parametere")),
			},
			mockError:      nil,
			expectedResult: AnnounceResponse{},
			expectedError:  true,
		},
		{
			name: "Successful announce with non-compact peers",
			req: trackerTypes.AnnounceRequest{
				InfoHash:   metainfo.Hash{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				PeerID:     types.PeerID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				IP:         net.ParseIP("192.168.1.1"),
				Port:       6881,
				Uploaded:   1024,
				Downloaded: 512,
				Left:       2048,
				Event:      trackerTypes.AnnounceEventEmpty,
				Key:        12345,
				NumWant:    50,
			},
			opt: AnnounceOpt{
				UserAgent:  "TestClient/1.0",
				HostHeader: "tracker.example.com",
			},
			mockResponse: &http.Response{
				StatusCode: 200,
				Body: io.NopCloser(bytes.NewBufferString(
					"d8:completei10e10:incompletei5e8:intervali1800e5:peersld2:ip9:127.0.0.14:porti6881eeee")),
			},
			mockError: nil,
			expectedResult: AnnounceResponse{
				Interval: 1800,
				Seeders:  10,
				Leechers: 5,
				Peers: []Peer{
					{
						IP:   net.ParseIP("127.0.0.1"),
						Port: 6881,
					},
				},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock HTTP transport
			mockTransport := &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return tt.mockResponse, tt.mockError
				},
			}

			// Create a client with the mock HTTP transport
			client := Client{
				hc: &http.Client{
					Transport: mockTransport,
				},
				url_: &url.URL{Scheme: "http", Host: "tracker.example.com", Path: "/announce"},
			}

			// Call the Announce method
			result, err := client.Announce(context.Background(), tt.req, tt.opt)
			// Check if error was expected
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Check interval
				if result.Interval != tt.expectedResult.Interval {
					t.Errorf("Interval mismatch: got %v, want %v", result.Interval, tt.expectedResult.Interval)
				}

				// Check seeders
				if result.Seeders != tt.expectedResult.Seeders {
					t.Errorf("Seeders mismatch: got %v, want %v", result.Seeders, tt.expectedResult.Seeders)
				}

				// Check leechers
				if result.Leechers != tt.expectedResult.Leechers {
					t.Errorf("Leechers mismatch: got %v, want %v", result.Leechers, tt.expectedResult.Leechers)
				}

				// Check peers length
				if len(result.Peers) != len(tt.expectedResult.Peers) {
					t.Errorf("Peers length mismatch: got %v, want %v", len(result.Peers), len(tt.expectedResult.Peers))
				} else {
					// Check each peer
					for i, peer := range result.Peers {
						expectedPeer := tt.expectedResult.Peers[i]
						if !peer.IP.Equal(expectedPeer.IP) {
							t.Errorf("Peer %d IP mismatch: got %v, want %v", i, peer.IP, expectedPeer.IP)
						}
						if peer.Port != expectedPeer.Port {
							t.Errorf("Peer %d Port mismatch: got %v, want %v", i, peer.Port, expectedPeer.Port)
						}
					}
				}
			}
		})
	}
}
