package sarama

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"
	"vendor"
)

func ExampleBroker() {
	broker := vendor.NewBroker("localhost:9092")
	err := broker.Open(nil)
	if err != nil {
		panic(err)
	}

	request := vendor.MetadataRequest{Topics: []string{"myTopic"}}
	response, err := broker.GetMetadata(&request)
	if err != nil {
		_ = broker.Close()
		panic(err)
	}

	fmt.Println("There are", len(response.Topics), "topics active in the cluster.")

	if err = broker.Close(); err != nil {
		panic(err)
	}
}

type mockEncoder struct {
	bytes []byte
}

func (m mockEncoder) encode(pe vendor.packetEncoder) error {
	return pe.putRawBytes(m.bytes)
}

type brokerMetrics struct {
	bytesRead    int
	bytesWritten int
}

func TestBrokerAccessors(t *testing.T) {
	broker := vendor.NewBroker("abc:123")

	if broker.ID() != -1 {
		t.Error("New broker didn't have an ID of -1.")
	}

	if broker.Addr() != "abc:123" {
		t.Error("New broker didn't have the correct address")
	}

	if broker.Rack() != "" {
		t.Error("New broker didn't have an unknown rack.")
	}

	broker.id = 34
	if broker.ID() != 34 {
		t.Error("Manually setting broker ID did not take effect.")
	}

	rack := "dc1"
	broker.rack = &rack
	if broker.Rack() != rack {
		t.Error("Manually setting broker rack did not take effect.")
	}
}

func TestSimpleBrokerCommunication(t *testing.T) {
	for _, tt := range brokerTestTable {
		vendor.Logger.Printf("Testing broker communication for %s", tt.name)
		mb := vendor.NewMockBroker(t, 0)
		mb.Returns(&mockEncoder{tt.response})
		pendingNotify := make(chan brokerMetrics)
		// Register a callback to be notified about successful requests
		mb.SetNotifier(func(bytesRead, bytesWritten int) {
			pendingNotify <- brokerMetrics{bytesRead, bytesWritten}
		})
		broker := vendor.NewBroker(mb.Addr())
		// Set the broker id in order to validate local broker metrics
		broker.id = 0
		conf := vendor.NewConfig()
		conf.Version = tt.version
		err := broker.Open(conf)
		if err != nil {
			t.Fatal(err)
		}
		tt.runner(t, broker)
		// Wait up to 500 ms for the remote broker to process the request and
		// notify us about the metrics
		timeout := 500 * time.Millisecond
		select {
		case mockBrokerMetrics := <-pendingNotify:
			validateBrokerMetrics(t, broker, mockBrokerMetrics)
		case <-time.After(timeout):
			t.Errorf("No request received for: %s after waiting for %v", tt.name, timeout)
		}
		mb.Close()
		err = broker.Close()
		if err != nil {
			t.Error(err)
		}
	}

}

var ErrTokenFailure = errors.New("Failure generating token")

type TokenProvider struct {
	accessToken *vendor.AccessToken
	err         error
}

func (t *TokenProvider) Token() (*vendor.AccessToken, error) {
	return t.accessToken, t.err
}

func newTokenProvider(token *vendor.AccessToken, err error) *TokenProvider {
	return &TokenProvider{
		accessToken: token,
		err:         err,
	}
}

func TestSASLOAuthBearer(t *testing.T) {

	testTable := []struct {
		name             string
		mockAuthErr      vendor.KError // Mock and expect error returned from SaslAuthenticateRequest
		mockHandshakeErr vendor.KError // Mock and expect error returned from SaslHandshakeRequest
		expectClientErr  bool          // Expect an internal client-side error
		tokProvider      *TokenProvider
	}{
		{
			name:             "SASL/OAUTHBEARER OK server response",
			mockAuthErr:      vendor.ErrNoError,
			mockHandshakeErr: vendor.ErrNoError,
			tokProvider:      newTokenProvider(&vendor.AccessToken{Token: "access-token-123"}, nil),
		},
		{
			name:             "SASL/OAUTHBEARER authentication failure response",
			mockAuthErr:      vendor.ErrSASLAuthenticationFailed,
			mockHandshakeErr: vendor.ErrNoError,
			tokProvider:      newTokenProvider(&vendor.AccessToken{Token: "access-token-123"}, nil),
		},
		{
			name:             "SASL/OAUTHBEARER handshake failure response",
			mockAuthErr:      vendor.ErrNoError,
			mockHandshakeErr: vendor.ErrSASLAuthenticationFailed,
			tokProvider:      newTokenProvider(&vendor.AccessToken{Token: "access-token-123"}, nil),
		},
		{
			name:             "SASL/OAUTHBEARER token generation error",
			mockAuthErr:      vendor.ErrNoError,
			mockHandshakeErr: vendor.ErrNoError,
			expectClientErr:  true,
			tokProvider:      newTokenProvider(&vendor.AccessToken{Token: "access-token-123"}, ErrTokenFailure),
		},
		{
			name:             "SASL/OAUTHBEARER invalid extension",
			mockAuthErr:      vendor.ErrNoError,
			mockHandshakeErr: vendor.ErrNoError,
			expectClientErr:  true,
			tokProvider: newTokenProvider(&vendor.AccessToken{
				Token:      "access-token-123",
				Extensions: map[string]string{"auth": "auth-value"},
			}, nil),
		},
	}

	for i, test := range testTable {

		// mockBroker mocks underlying network logic and broker responses
		mockBroker := vendor.NewMockBroker(t, 0)

		mockSASLAuthResponse := vendor.NewMockSaslAuthenticateResponse(t).SetAuthBytes([]byte("response_payload"))
		if test.mockAuthErr != vendor.ErrNoError {
			mockSASLAuthResponse = mockSASLAuthResponse.SetError(test.mockAuthErr)
		}

		mockSASLHandshakeResponse := vendor.NewMockSaslHandshakeResponse(t).SetEnabledMechanisms([]string{vendor.SASLTypeOAuth})
		if test.mockHandshakeErr != vendor.ErrNoError {
			mockSASLHandshakeResponse = mockSASLHandshakeResponse.SetError(test.mockHandshakeErr)
		}

		mockBroker.SetHandlerByMap(map[string]vendor.MockResponse{
			"SaslAuthenticateRequest": mockSASLAuthResponse,
			"SaslHandshakeRequest":    mockSASLHandshakeResponse,
		})

		// broker executes SASL requests against mockBroker
		broker := vendor.NewBroker(mockBroker.Addr())
		broker.requestRate = vendor.NilMeter{}
		broker.outgoingByteRate = vendor.NilMeter{}
		broker.incomingByteRate = vendor.NilMeter{}
		broker.requestSize = vendor.NilHistogram{}
		broker.responseSize = vendor.NilHistogram{}
		broker.responseRate = vendor.NilMeter{}
		broker.requestLatency = vendor.NilHistogram{}

		conf := vendor.NewConfig()
		conf.Net.SASL.Mechanism = vendor.SASLTypeOAuth
		conf.Net.SASL.TokenProvider = test.tokProvider

		broker.conf = conf

		dialer := net.Dialer{
			Timeout:   conf.Net.DialTimeout,
			KeepAlive: conf.Net.KeepAlive,
			LocalAddr: conf.Net.LocalAddr,
		}

		conn, err := dialer.Dial("tcp", mockBroker.listener.Addr().String())

		if err != nil {
			t.Fatal(err)
		}

		broker.conn = conn

		err = broker.authenticateViaSASL()

		if test.mockAuthErr != vendor.ErrNoError {
			if test.mockAuthErr != err {
				t.Errorf("[%d]:[%s] Expected %s auth error, got %s\n", i, test.name, test.mockAuthErr, err)
			}
		} else if test.mockHandshakeErr != vendor.ErrNoError {
			if test.mockHandshakeErr != err {
				t.Errorf("[%d]:[%s] Expected %s handshake error, got %s\n", i, test.name, test.mockHandshakeErr, err)
			}
		} else if test.expectClientErr && err == nil {
			t.Errorf("[%d]:[%s] Expected a client error and got none\n", i, test.name)
		} else if !test.expectClientErr && err != nil {
			t.Errorf("[%d]:[%s] Unexpected error, got %s\n", i, test.name, err)
		}

		mockBroker.Close()
	}
}

// A mock scram client.
type MockSCRAMClient struct {
	done bool
}

func (m *MockSCRAMClient) Begin(userName, password, authzID string) (err error) {
	return nil
}

func (m *MockSCRAMClient) Step(challenge string) (response string, err error) {
	if challenge == "" {
		return "ping", nil
	}
	if challenge == "pong" {
		m.done = true
		return "", nil
	}
	return "", errors.New("failed to authenticate :(")
}

func (m *MockSCRAMClient) Done() bool {
	return m.done
}

var _ vendor.SCRAMClient = &MockSCRAMClient{}

func TestSASLSCRAMSHAXXX(t *testing.T) {
	testTable := []struct {
		name               string
		mockHandshakeErr   vendor.KError
		mockSASLAuthErr    vendor.KError
		expectClientErr    bool
		scramClient        *MockSCRAMClient
		scramChallengeResp string
	}{
		{
			name:               "SASL/SCRAMSHAXXX successfull authentication",
			mockHandshakeErr:   vendor.ErrNoError,
			scramClient:        &MockSCRAMClient{},
			scramChallengeResp: "pong",
		},
		{
			name:               "SASL/SCRAMSHAXXX SCRAM client step error client",
			mockHandshakeErr:   vendor.ErrNoError,
			mockSASLAuthErr:    vendor.ErrNoError,
			scramClient:        &MockSCRAMClient{},
			scramChallengeResp: "gong",
			expectClientErr:    true,
		},
		{
			name:               "SASL/SCRAMSHAXXX server authentication error",
			mockHandshakeErr:   vendor.ErrNoError,
			mockSASLAuthErr:    vendor.ErrSASLAuthenticationFailed,
			scramClient:        &MockSCRAMClient{},
			scramChallengeResp: "pong",
		},
		{
			name:               "SASL/SCRAMSHAXXX unsupported SCRAM mechanism",
			mockHandshakeErr:   vendor.ErrUnsupportedSASLMechanism,
			mockSASLAuthErr:    vendor.ErrNoError,
			scramClient:        &MockSCRAMClient{},
			scramChallengeResp: "pong",
		},
	}

	for i, test := range testTable {

		// mockBroker mocks underlying network logic and broker responses
		mockBroker := vendor.NewMockBroker(t, 0)
		broker := vendor.NewBroker(mockBroker.Addr())
		// broker executes SASL requests against mockBroker
		broker.requestRate = vendor.NilMeter{}
		broker.outgoingByteRate = vendor.NilMeter{}
		broker.incomingByteRate = vendor.NilMeter{}
		broker.requestSize = vendor.NilHistogram{}
		broker.responseSize = vendor.NilHistogram{}
		broker.responseRate = vendor.NilMeter{}
		broker.requestLatency = vendor.NilHistogram{}

		mockSASLAuthResponse := vendor.NewMockSaslAuthenticateResponse(t).SetAuthBytes([]byte(test.scramChallengeResp))
		mockSASLHandshakeResponse := vendor.NewMockSaslHandshakeResponse(t).SetEnabledMechanisms([]string{vendor.SASLTypeSCRAMSHA256, vendor.SASLTypeSCRAMSHA512})

		if test.mockSASLAuthErr != vendor.ErrNoError {
			mockSASLAuthResponse = mockSASLAuthResponse.SetError(test.mockSASLAuthErr)
		}
		if test.mockHandshakeErr != vendor.ErrNoError {
			mockSASLHandshakeResponse = mockSASLHandshakeResponse.SetError(test.mockHandshakeErr)
		}

		mockBroker.SetHandlerByMap(map[string]vendor.MockResponse{
			"SaslAuthenticateRequest": mockSASLAuthResponse,
			"SaslHandshakeRequest":    mockSASLHandshakeResponse,
		})

		conf := vendor.NewConfig()
		conf.Net.SASL.Mechanism = vendor.SASLTypeSCRAMSHA512
		conf.Net.SASL.SCRAMClientGeneratorFunc = func() vendor.SCRAMClient { return test.scramClient }

		broker.conf = conf
		dialer := net.Dialer{
			Timeout:   conf.Net.DialTimeout,
			KeepAlive: conf.Net.KeepAlive,
			LocalAddr: conf.Net.LocalAddr,
		}

		conn, err := dialer.Dial("tcp", mockBroker.listener.Addr().String())

		if err != nil {
			t.Fatal(err)
		}

		broker.conn = conn

		err = broker.authenticateViaSASL()

		if test.mockSASLAuthErr != vendor.ErrNoError {
			if test.mockSASLAuthErr != err {
				t.Errorf("[%d]:[%s] Expected %s SASL authentication error, got %s\n", i, test.name, test.mockHandshakeErr, err)
			}
		} else if test.mockHandshakeErr != vendor.ErrNoError {
			if test.mockHandshakeErr != err {
				t.Errorf("[%d]:[%s] Expected %s handshake error, got %s\n", i, test.name, test.mockHandshakeErr, err)
			}
		} else if test.expectClientErr && err == nil {
			t.Errorf("[%d]:[%s] Expected a client error and got none\n", i, test.name)
		} else if !test.expectClientErr && err != nil {
			t.Errorf("[%d]:[%s] Unexpected error, got %s\n", i, test.name, err)
		}

		mockBroker.Close()
	}
}

func TestSASLPlainAuth(t *testing.T) {

	testTable := []struct {
		name             string
		mockAuthErr      vendor.KError // Mock and expect error returned from SaslAuthenticateRequest
		mockHandshakeErr vendor.KError // Mock and expect error returned from SaslHandshakeRequest
		expectClientErr  bool          // Expect an internal client-side error
	}{
		{
			name:             "SASL Plain OK server response",
			mockAuthErr:      vendor.ErrNoError,
			mockHandshakeErr: vendor.ErrNoError,
		},
		{
			name:             "SASL Plain authentication failure response",
			mockAuthErr:      vendor.ErrSASLAuthenticationFailed,
			mockHandshakeErr: vendor.ErrNoError,
		},
		{
			name:             "SASL Plain handshake failure response",
			mockAuthErr:      vendor.ErrNoError,
			mockHandshakeErr: vendor.ErrSASLAuthenticationFailed,
		},
	}

	for i, test := range testTable {

		// mockBroker mocks underlying network logic and broker responses
		mockBroker := vendor.NewMockBroker(t, 0)

		mockSASLAuthResponse := vendor.NewMockSaslAuthenticateResponse(t).
			SetAuthBytes([]byte(`response_payload`))

		if test.mockAuthErr != vendor.ErrNoError {
			mockSASLAuthResponse = mockSASLAuthResponse.SetError(test.mockAuthErr)
		}

		mockSASLHandshakeResponse := vendor.NewMockSaslHandshakeResponse(t).
			SetEnabledMechanisms([]string{vendor.SASLTypePlaintext})

		if test.mockHandshakeErr != vendor.ErrNoError {
			mockSASLHandshakeResponse = mockSASLHandshakeResponse.SetError(test.mockHandshakeErr)
		}

		mockBroker.SetHandlerByMap(map[string]vendor.MockResponse{
			"SaslAuthenticateRequest": mockSASLAuthResponse,
			"SaslHandshakeRequest":    mockSASLHandshakeResponse,
		})

		// broker executes SASL requests against mockBroker
		broker := vendor.NewBroker(mockBroker.Addr())
		broker.requestRate = vendor.NilMeter{}
		broker.outgoingByteRate = vendor.NilMeter{}
		broker.incomingByteRate = vendor.NilMeter{}
		broker.requestSize = vendor.NilHistogram{}
		broker.responseSize = vendor.NilHistogram{}
		broker.responseRate = vendor.NilMeter{}
		broker.requestLatency = vendor.NilHistogram{}

		conf := vendor.NewConfig()
		conf.Net.SASL.Mechanism = vendor.SASLTypePlaintext
		conf.Net.SASL.User = "token"
		conf.Net.SASL.Password = "password"

		broker.conf = conf
		broker.conf.Version = vendor.V1_0_0_0
		dialer := net.Dialer{
			Timeout:   conf.Net.DialTimeout,
			KeepAlive: conf.Net.KeepAlive,
			LocalAddr: conf.Net.LocalAddr,
		}

		conn, err := dialer.Dial("tcp", mockBroker.listener.Addr().String())

		if err != nil {
			t.Fatal(err)
		}

		broker.conn = conn

		err = broker.authenticateViaSASL()

		if test.mockAuthErr != vendor.ErrNoError {
			if test.mockAuthErr != err {
				t.Errorf("[%d]:[%s] Expected %s auth error, got %s\n", i, test.name, test.mockAuthErr, err)
			}
		} else if test.mockHandshakeErr != vendor.ErrNoError {
			if test.mockHandshakeErr != err {
				t.Errorf("[%d]:[%s] Expected %s handshake error, got %s\n", i, test.name, test.mockHandshakeErr, err)
			}
		} else if test.expectClientErr && err == nil {
			t.Errorf("[%d]:[%s] Expected a client error and got none\n", i, test.name)
		} else if !test.expectClientErr && err != nil {
			t.Errorf("[%d]:[%s] Unexpected error, got %s\n", i, test.name, err)
		}

		mockBroker.Close()
	}
}

func TestBuildClientInitialResponse(t *testing.T) {

	testTable := []struct {
		name        string
		token       *vendor.AccessToken
		expected    []byte
		expectError bool
	}{
		{
			name: "Build SASL client initial response with two extensions",
			token: &vendor.AccessToken{
				Token: "the-token",
				Extensions: map[string]string{
					"x": "1",
					"y": "2",
				},
			},
			expected: []byte("n,,\x01auth=Bearer the-token\x01x=1\x01y=2\x01\x01"),
		},
		{
			name:     "Build SASL client initial response with no extensions",
			token:    &vendor.AccessToken{Token: "the-token"},
			expected: []byte("n,,\x01auth=Bearer the-token\x01\x01"),
		},
		{
			name: "Build SASL client initial response using reserved extension",
			token: &vendor.AccessToken{
				Token: "the-token",
				Extensions: map[string]string{
					"auth": "auth-value",
				},
			},
			expected:    []byte(""),
			expectError: true,
		},
	}

	for i, test := range testTable {

		actual, err := vendor.buildClientInitialResponse(test.token)

		if !reflect.DeepEqual(test.expected, actual) {
			t.Errorf("Expected %s, got %s\n", test.expected, actual)
		}
		if test.expectError && err == nil {
			t.Errorf("[%d]:[%s] Expected an error but did not get one", i, test.name)
		}
		if !test.expectError && err != nil {
			t.Errorf("[%d]:[%s] Expected no error but got %s\n", i, test.name, err)
		}
	}
}

// We're not testing encoding/decoding here, so most of the requests/responses will be empty for simplicity's sake
var brokerTestTable = []struct {
	version  vendor.KafkaVersion
	name     string
	response []byte
	runner   func(*testing.T, *vendor.Broker)
}{
	{vendor.V0_10_0_0,
		"MetadataRequest",
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		func(t *testing.T, broker *vendor.Broker) {
			request := vendor.MetadataRequest{}
			response, err := broker.GetMetadata(&request)
			if err != nil {
				t.Error(err)
			}
			if response == nil {
				t.Error("Metadata request got no response!")
			}
		}},

	{vendor.V0_10_0_0,
		"ConsumerMetadataRequest",
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 't', 0x00, 0x00, 0x00, 0x00},
		func(t *testing.T, broker *vendor.Broker) {
			request := vendor.ConsumerMetadataRequest{}
			response, err := broker.GetConsumerMetadata(&request)
			if err != nil {
				t.Error(err)
			}
			if response == nil {
				t.Error("Consumer Metadata request got no response!")
			}
		}},

	{vendor.V0_10_0_0,
		"ProduceRequest (NoResponse)",
		[]byte{},
		func(t *testing.T, broker *vendor.Broker) {
			request := vendor.ProduceRequest{}
			request.RequiredAcks = vendor.NoResponse
			response, err := broker.Produce(&request)
			if err != nil {
				t.Error(err)
			}
			if response != nil {
				t.Error("Produce request with NoResponse got a response!")
			}
		}},

	{vendor.V0_10_0_0,
		"ProduceRequest (WaitForLocal)",
		[]byte{0x00, 0x00, 0x00, 0x00},
		func(t *testing.T, broker *vendor.Broker) {
			request := vendor.ProduceRequest{}
			request.RequiredAcks = vendor.WaitForLocal
			response, err := broker.Produce(&request)
			if err != nil {
				t.Error(err)
			}
			if response == nil {
				t.Error("Produce request without NoResponse got no response!")
			}
		}},

	{vendor.V0_10_0_0,
		"FetchRequest",
		[]byte{0x00, 0x00, 0x00, 0x00},
		func(t *testing.T, broker *vendor.Broker) {
			request := vendor.FetchRequest{}
			response, err := broker.Fetch(&request)
			if err != nil {
				t.Error(err)
			}
			if response == nil {
				t.Error("Fetch request got no response!")
			}
		}},

	{vendor.V0_10_0_0,
		"OffsetFetchRequest",
		[]byte{0x00, 0x00, 0x00, 0x00},
		func(t *testing.T, broker *vendor.Broker) {
			request := vendor.OffsetFetchRequest{}
			response, err := broker.FetchOffset(&request)
			if err != nil {
				t.Error(err)
			}
			if response == nil {
				t.Error("OffsetFetch request got no response!")
			}
		}},

	{vendor.V0_10_0_0,
		"OffsetCommitRequest",
		[]byte{0x00, 0x00, 0x00, 0x00},
		func(t *testing.T, broker *vendor.Broker) {
			request := vendor.OffsetCommitRequest{}
			response, err := broker.CommitOffset(&request)
			if err != nil {
				t.Error(err)
			}
			if response == nil {
				t.Error("OffsetCommit request got no response!")
			}
		}},

	{vendor.V0_10_0_0,
		"OffsetRequest",
		[]byte{0x00, 0x00, 0x00, 0x00},
		func(t *testing.T, broker *vendor.Broker) {
			request := vendor.OffsetRequest{}
			response, err := broker.GetAvailableOffsets(&request)
			if err != nil {
				t.Error(err)
			}
			if response == nil {
				t.Error("Offset request got no response!")
			}
		}},

	{vendor.V0_10_0_0,
		"JoinGroupRequest",
		[]byte{0x00, 0x17, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		func(t *testing.T, broker *vendor.Broker) {
			request := vendor.JoinGroupRequest{}
			response, err := broker.JoinGroup(&request)
			if err != nil {
				t.Error(err)
			}
			if response == nil {
				t.Error("JoinGroup request got no response!")
			}
		}},

	{vendor.V0_10_0_0,
		"SyncGroupRequest",
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		func(t *testing.T, broker *vendor.Broker) {
			request := vendor.SyncGroupRequest{}
			response, err := broker.SyncGroup(&request)
			if err != nil {
				t.Error(err)
			}
			if response == nil {
				t.Error("SyncGroup request got no response!")
			}
		}},

	{vendor.V0_10_0_0,
		"LeaveGroupRequest",
		[]byte{0x00, 0x00},
		func(t *testing.T, broker *vendor.Broker) {
			request := vendor.LeaveGroupRequest{}
			response, err := broker.LeaveGroup(&request)
			if err != nil {
				t.Error(err)
			}
			if response == nil {
				t.Error("LeaveGroup request got no response!")
			}
		}},

	{vendor.V0_10_0_0,
		"HeartbeatRequest",
		[]byte{0x00, 0x00},
		func(t *testing.T, broker *vendor.Broker) {
			request := vendor.HeartbeatRequest{}
			response, err := broker.Heartbeat(&request)
			if err != nil {
				t.Error(err)
			}
			if response == nil {
				t.Error("Heartbeat request got no response!")
			}
		}},

	{vendor.V0_10_0_0,
		"ListGroupsRequest",
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		func(t *testing.T, broker *vendor.Broker) {
			request := vendor.ListGroupsRequest{}
			response, err := broker.ListGroups(&request)
			if err != nil {
				t.Error(err)
			}
			if response == nil {
				t.Error("ListGroups request got no response!")
			}
		}},

	{vendor.V0_10_0_0,
		"DescribeGroupsRequest",
		[]byte{0x00, 0x00, 0x00, 0x00},
		func(t *testing.T, broker *vendor.Broker) {
			request := vendor.DescribeGroupsRequest{}
			response, err := broker.DescribeGroups(&request)
			if err != nil {
				t.Error(err)
			}
			if response == nil {
				t.Error("DescribeGroups request got no response!")
			}
		}},

	{vendor.V0_10_0_0,
		"ApiVersionsRequest",
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		func(t *testing.T, broker *vendor.Broker) {
			request := vendor.ApiVersionsRequest{}
			response, err := broker.ApiVersions(&request)
			if err != nil {
				t.Error(err)
			}
			if response == nil {
				t.Error("ApiVersions request got no response!")
			}
		}},

	{vendor.V1_1_0_0,
		"DeleteGroupsRequest",
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		func(t *testing.T, broker *vendor.Broker) {
			request := vendor.DeleteGroupsRequest{}
			response, err := broker.DeleteGroups(&request)
			if err != nil {
				t.Error(err)
			}
			if response == nil {
				t.Error("DeleteGroups request got no response!")
			}
		}},
}

func validateBrokerMetrics(t *testing.T, broker *vendor.Broker, mockBrokerMetrics brokerMetrics) {
	metricValidators := vendor.newMetricValidators()
	mockBrokerBytesRead := mockBrokerMetrics.bytesRead
	mockBrokerBytesWritten := mockBrokerMetrics.bytesWritten

	// Check that the number of bytes sent corresponds to what the mock broker received
	metricValidators.registerForAllBrokers(broker, vendor.countMeterValidator("incoming-byte-rate", mockBrokerBytesWritten))
	if mockBrokerBytesWritten == 0 {
		// This a ProduceRequest with NoResponse
		metricValidators.registerForAllBrokers(broker, vendor.countMeterValidator("response-rate", 0))
		metricValidators.registerForAllBrokers(broker, vendor.countHistogramValidator("response-size", 0))
		metricValidators.registerForAllBrokers(broker, vendor.minMaxHistogramValidator("response-size", 0, 0))
	} else {
		metricValidators.registerForAllBrokers(broker, vendor.countMeterValidator("response-rate", 1))
		metricValidators.registerForAllBrokers(broker, vendor.countHistogramValidator("response-size", 1))
		metricValidators.registerForAllBrokers(broker, vendor.minMaxHistogramValidator("response-size", mockBrokerBytesWritten, mockBrokerBytesWritten))
	}

	// Check that the number of bytes received corresponds to what the mock broker sent
	metricValidators.registerForAllBrokers(broker, vendor.countMeterValidator("outgoing-byte-rate", mockBrokerBytesRead))
	metricValidators.registerForAllBrokers(broker, vendor.countMeterValidator("request-rate", 1))
	metricValidators.registerForAllBrokers(broker, vendor.countHistogramValidator("request-size", 1))
	metricValidators.registerForAllBrokers(broker, vendor.minMaxHistogramValidator("request-size", mockBrokerBytesRead, mockBrokerBytesRead))

	// Run the validators
	metricValidators.run(t, broker.conf.MetricRegistry)
}
