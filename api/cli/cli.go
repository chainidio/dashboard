package cli

import (
	"time"

	"github.com/chainid-io/dashboard"

	"os"
	"path/filepath"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

// Service implements the CLIService interface
type Service struct{}

const (
	errInvalidEndpointProtocol       = chainid.Error("Invalid endpoint protocol: Chain Platform only supports unix:// or tcp://")
	errSocketNotFound                = chainid.Error("Unable to locate Unix socket")
	errEndpointsFileNotFound         = chainid.Error("Unable to locate external endpoints file")
	errInvalidSyncInterval           = chainid.Error("Invalid synchronization interval")
	errEndpointExcludeExternal       = chainid.Error("Cannot use the -H flag mutually with --external-endpoints")
	errNoAuthExcludeAdminPassword    = chainid.Error("Cannot use --no-auth with --admin-password or --admin-password-file")
	errAdminPassExcludeAdminPassFile = chainid.Error("Cannot use --admin-password with --admin-password-file")
)

// ParseFlags parse the CLI flags and return a chainid.Flags struct
func (*Service) ParseFlags(version string) (*chainid.CLIFlags, error) {
	kingpin.Version(version)

	flags := &chainid.CLIFlags{
		Addr:              kingpin.Flag("bind", "Address and port to serve Chain Platform").Default(defaultBindAddress).Short('p').String(),
		Assets:            kingpin.Flag("assets", "Path to the assets").Default(defaultAssetsDirectory).Short('a').String(),
		Data:              kingpin.Flag("data", "Path to the folder where the data is stored").Default(defaultDataDirectory).Short('d').String(),
		EndpointURL:       kingpin.Flag("host", "Endpoint URL").Short('H').String(),
		ExternalEndpoints: kingpin.Flag("external-endpoints", "Path to a file defining available endpoints").String(),
		NoAuth:            kingpin.Flag("no-auth", "Disable authentication").Default(defaultNoAuth).Bool(),
		NoAnalytics:       kingpin.Flag("no-analytics", "Disable Analytics in app").Default(defaultNoAnalytics).Bool(),
		TLS:               kingpin.Flag("tlsverify", "TLS support").Default(defaultTLS).Bool(),
		TLSSkipVerify:     kingpin.Flag("tlsskipverify", "Disable TLS server verification").Default(defaultTLSSkipVerify).Bool(),
		TLSCacert:         kingpin.Flag("tlscacert", "Path to the CA").Default(defaultTLSCACertPath).String(),
		TLSCert:           kingpin.Flag("tlscert", "Path to the TLS certificate file").Default(defaultTLSCertPath).String(),
		TLSKey:            kingpin.Flag("tlskey", "Path to the TLS key").Default(defaultTLSKeyPath).String(),
		SSL:               kingpin.Flag("ssl", "Secure Chain Platform instance using SSL").Default(defaultSSL).Bool(),
		SSLCert:           kingpin.Flag("sslcert", "Path to the SSL certificate used to secure the Chain Platform instance").Default(defaultSSLCertPath).String(),
		SSLKey:            kingpin.Flag("sslkey", "Path to the SSL key used to secure the Chain Platform instance").Default(defaultSSLKeyPath).String(),
		SyncInterval:      kingpin.Flag("sync-interval", "Duration between each synchronization via the external endpoints source").Default(defaultSyncInterval).String(),
		AdminPassword:     kingpin.Flag("admin-password", "Hashed admin password").String(),
		AdminPasswordFile: kingpin.Flag("admin-password-file", "Path to the file containing the password for the admin user").String(),
		Labels:            pairs(kingpin.Flag("hide-label", "Hide containers with a specific label in the UI").Short('l')),
		Logo:              kingpin.Flag("logo", "URL for the logo displayed in the UI").String(),
		Templates:         kingpin.Flag("templates", "URL to the templates (apps) definitions").Short('t').String(),
	}

	kingpin.Parse()

	if !filepath.IsAbs(*flags.Assets) {
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		*flags.Assets = filepath.Join(filepath.Dir(ex), *flags.Assets)
	}

	return flags, nil
}

// ValidateFlags validates the values of the flags.
func (*Service) ValidateFlags(flags *chainid.CLIFlags) error {

	if *flags.EndpointURL != "" && *flags.ExternalEndpoints != "" {
		return errEndpointExcludeExternal
	}

	err := validateEndpointURL(*flags.EndpointURL)
	if err != nil {
		return err
	}

	err = validateExternalEndpoints(*flags.ExternalEndpoints)
	if err != nil {
		return err
	}

	err = validateSyncInterval(*flags.SyncInterval)
	if err != nil {
		return err
	}

	if *flags.NoAuth && (*flags.AdminPassword != "" || *flags.AdminPasswordFile != "") {
		return errNoAuthExcludeAdminPassword
	}

	if *flags.AdminPassword != "" && *flags.AdminPasswordFile != "" {
		return errAdminPassExcludeAdminPassFile
	}

	return nil
}

func validateEndpointURL(endpointURL string) error {
	if endpointURL != "" {
		if !strings.HasPrefix(endpointURL, "unix://") && !strings.HasPrefix(endpointURL, "tcp://") {
			return errInvalidEndpointProtocol
		}

		if strings.HasPrefix(endpointURL, "unix://") {
			socketPath := strings.TrimPrefix(endpointURL, "unix://")
			if _, err := os.Stat(socketPath); err != nil {
				if os.IsNotExist(err) {
					return errSocketNotFound
				}
				return err
			}
		}
	}
	return nil
}

func validateExternalEndpoints(externalEndpoints string) error {
	if externalEndpoints != "" {
		if _, err := os.Stat(externalEndpoints); err != nil {
			if os.IsNotExist(err) {
				return errEndpointsFileNotFound
			}
			return err
		}
	}
	return nil
}

func validateSyncInterval(syncInterval string) error {
	if syncInterval != defaultSyncInterval {
		_, err := time.ParseDuration(syncInterval)
		if err != nil {
			return errInvalidSyncInterval
		}
	}
	return nil
}
