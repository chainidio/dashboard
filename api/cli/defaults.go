// +build !windows

package cli

const (
	defaultBindAddress     = ":9000"
	defaultDataDirectory   = "/data"
	defaultAssetsDirectory = "./"
	defaultNoAuth          = "false"
	defaultNoAnalytics     = "false"
	defaultTLS             = "false"
	defaultTLSSkipVerify   = "false"
	defaultTLSCACertPath   = "/certs/ca.pem"
	defaultTLSCertPath     = "/certs/cert.pem"
	defaultTLSKeyPath      = "/certs/key.pem"
	defaultSSL             = "false"
	defaultSSLCertPath     = "/certs/chainid.crt"
	defaultSSLKeyPath      = "/certs/chainid.key"
	defaultSyncInterval    = "60s"
)
