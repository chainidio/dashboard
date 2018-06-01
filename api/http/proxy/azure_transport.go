package proxy

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/chainid-io/dashboard"
	"github.com/chainid-io/dashboard/http/client"
)

type (
	azureAPIToken struct {
		value          string
		expirationTime time.Time
	}

	// AzureTransport represents a transport used when executing HTTP requests
	// against the Azure API.
	AzureTransport struct {
		credentials *chainid.AzureCredentials
		client      *client.HTTPClient
		token       *azureAPIToken
		mutex       sync.Mutex
	}
)

// NewAzureTransport returns a pointer to an AzureTransport instance.
func NewAzureTransport(credentials *chainid.AzureCredentials) *AzureTransport {
	return &AzureTransport{
		credentials: credentials,
		client:      client.NewHTTPClient(),
	}
}

func (transport *AzureTransport) authenticate() error {
	token, err := transport.client.ExecuteAzureAuthenticationRequest(transport.credentials)
	if err != nil {
		return err
	}

	expiresOn, err := strconv.ParseInt(token.ExpiresOn, 10, 64)
	if err != nil {
		return err
	}

	transport.token = &azureAPIToken{
		value:          token.AccessToken,
		expirationTime: time.Unix(expiresOn, 0),
	}

	return nil
}

func (transport *AzureTransport) retrieveAuthenticationToken() error {
	transport.mutex.Lock()
	defer transport.mutex.Unlock()

	if transport.token == nil {
		return transport.authenticate()
	}

	timeLimit := time.Now().Add(-5 * time.Minute)
	if timeLimit.After(transport.token.expirationTime) {
		return transport.authenticate()
	}

	return nil
}

// RoundTrip is the implementation of the Transport interface.
func (transport *AzureTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	err := transport.retrieveAuthenticationToken()
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+transport.token.value)
	return http.DefaultTransport.RoundTrip(request)
}
