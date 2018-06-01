package main // import "github.com/chainid-io/dashboard"

import (
	"strings"

	"github.com/chainid-io/dashboard"
	"github.com/chainid-io/dashboard/bolt"
	"github.com/chainid-io/dashboard/cli"
	"github.com/chainid-io/dashboard/cron"
	"github.com/chainid-io/dashboard/crypto"
	"github.com/chainid-io/dashboard/exec"
	"github.com/chainid-io/dashboard/filesystem"
	"github.com/chainid-io/dashboard/git"
	"github.com/chainid-io/dashboard/http"
	"github.com/chainid-io/dashboard/http/client"
	"github.com/chainid-io/dashboard/jwt"
	"github.com/chainid-io/dashboard/ldap"

	"log"
)

func initCLI() *chainid.CLIFlags {
	var cli chainid.CLIService = &cli.Service{}
	flags, err := cli.ParseFlags(chainid.APIVersion)
	if err != nil {
		log.Fatal(err)
	}

	err = cli.ValidateFlags(flags)
	if err != nil {
		log.Fatal(err)
	}
	return flags
}

func initFileService(dataStorePath string) chainid.FileService {
	fileService, err := filesystem.NewService(dataStorePath, "")
	if err != nil {
		log.Fatal(err)
	}
	return fileService
}

func initStore(dataStorePath string) *bolt.Store {
	store, err := bolt.NewStore(dataStorePath)
	if err != nil {
		log.Fatal(err)
	}

	err = store.Open()
	if err != nil {
		log.Fatal(err)
	}

	err = store.Init()
	if err != nil {
		log.Fatal(err)
	}

	err = store.MigrateData()
	if err != nil {
		log.Fatal(err)
	}
	return store
}

func initStackManager(assetsPath string, dataStorePath string, signatureService chainid.DigitalSignatureService, fileService chainid.FileService) (chainid.StackManager, error) {
	return exec.NewStackManager(assetsPath, dataStorePath, signatureService, fileService)
}

func initJWTService(authenticationEnabled bool) chainid.JWTService {
	if authenticationEnabled {
		jwtService, err := jwt.NewService()
		if err != nil {
			log.Fatal(err)
		}
		return jwtService
	}
	return nil
}

func initDigitalSignatureService() chainid.DigitalSignatureService {
	return &crypto.ECDSAService{}
}

func initCryptoService() chainid.CryptoService {
	return &crypto.Service{}
}

func initLDAPService() chainid.LDAPService {
	return &ldap.Service{}
}

func initGitService() chainid.GitService {
	return &git.Service{}
}

func initEndpointWatcher(endpointService chainid.EndpointService, externalEnpointFile string, syncInterval string) bool {
	authorizeEndpointMgmt := true
	if externalEnpointFile != "" {
		authorizeEndpointMgmt = false
		log.Println("Using external endpoint definition. Endpoint management via the API will be disabled.")
		endpointWatcher := cron.NewWatcher(endpointService, syncInterval)
		err := endpointWatcher.WatchEndpointFile(externalEnpointFile)
		if err != nil {
			log.Fatal(err)
		}
	}
	return authorizeEndpointMgmt
}

func initStatus(authorizeEndpointMgmt bool, flags *chainid.CLIFlags) *chainid.Status {
	return &chainid.Status{
		Analytics:          !*flags.NoAnalytics,
		Authentication:     !*flags.NoAuth,
		EndpointManagement: authorizeEndpointMgmt,
		Version:            chainid.APIVersion,
	}
}

func initDockerHub(dockerHubService chainid.DockerHubService) error {
	_, err := dockerHubService.DockerHub()
	if err == chainid.ErrDockerHubNotFound {
		dockerhub := &chainid.DockerHub{
			Authentication: false,
			Username:       "",
			Password:       "",
		}
		return dockerHubService.StoreDockerHub(dockerhub)
	} else if err != nil {
		return err
	}

	return nil
}

func initSettings(settingsService chainid.SettingsService, flags *chainid.CLIFlags) error {
	_, err := settingsService.Settings()
	if err == chainid.ErrSettingsNotFound {
		settings := &chainid.Settings{
			LogoURL:                     *flags.Logo,
			DisplayExternalContributors: false,
			AuthenticationMethod:        chainid.AuthenticationInternal,
			LDAPSettings: chainid.LDAPSettings{
				TLSConfig: chainid.TLSConfiguration{},
				SearchSettings: []chainid.LDAPSearchSettings{
					chainid.LDAPSearchSettings{},
				},
			},
			AllowBindMountsForRegularUsers:     true,
			AllowPrivilegedModeForRegularUsers: true,
		}

		if *flags.Templates != "" {
			settings.TemplatesURL = *flags.Templates
		} else {
			settings.TemplatesURL = chainid.DefaultTemplatesURL
		}

		if *flags.Labels != nil {
			settings.BlackListedLabels = *flags.Labels
		} else {
			settings.BlackListedLabels = make([]chainid.Pair, 0)
		}

		return settingsService.StoreSettings(settings)
	} else if err != nil {
		return err
	}

	return nil
}

func retrieveFirstEndpointFromDatabase(endpointService chainid.EndpointService) *chainid.Endpoint {
	endpoints, err := endpointService.Endpoints()
	if err != nil {
		log.Fatal(err)
	}
	return &endpoints[0]
}

func loadAndParseKeyPair(fileService chainid.FileService, signatureService chainid.DigitalSignatureService) error {
	private, public, err := fileService.LoadKeyPair()
	if err != nil {
		return err
	}
	return signatureService.ParseKeyPair(private, public)
}

func generateAndStoreKeyPair(fileService chainid.FileService, signatureService chainid.DigitalSignatureService) error {
	private, public, err := signatureService.GenerateKeyPair()
	if err != nil {
		return err
	}
	privateHeader, publicHeader := signatureService.PEMHeaders()
	return fileService.StoreKeyPair(private, public, privateHeader, publicHeader)
}

func initKeyPair(fileService chainid.FileService, signatureService chainid.DigitalSignatureService) error {
	existingKeyPair, err := fileService.KeyPairFilesExist()
	if err != nil {
		log.Fatal(err)
	}

	if existingKeyPair {
		return loadAndParseKeyPair(fileService, signatureService)
	}
	return generateAndStoreKeyPair(fileService, signatureService)
}

func createTLSSecuredEndpoint(flags *chainid.CLIFlags, endpointService chainid.EndpointService) error {
	tlsConfiguration := chainid.TLSConfiguration{
		TLS:           *flags.TLS,
		TLSSkipVerify: *flags.TLSSkipVerify,
	}

	if *flags.TLS {
		tlsConfiguration.TLSCACertPath = *flags.TLSCacert
		tlsConfiguration.TLSCertPath = *flags.TLSCert
		tlsConfiguration.TLSKeyPath = *flags.TLSKey
	} else if !*flags.TLS && *flags.TLSSkipVerify {
		tlsConfiguration.TLS = true
	}

	endpoint := &chainid.Endpoint{
		Name:            "primary",
		URL:             *flags.EndpointURL,
		GroupID:         chainid.EndpointGroupID(1),
		Type:            chainid.DockerEnvironment,
		TLSConfig:       tlsConfiguration,
		AuthorizedUsers: []chainid.UserID{},
		AuthorizedTeams: []chainid.TeamID{},
		Extensions:      []chainid.EndpointExtension{},
	}

	if strings.HasPrefix(endpoint.URL, "tcp://") {
		tlsConfig, err := crypto.CreateTLSConfigurationFromDisk(tlsConfiguration.TLSCACertPath, tlsConfiguration.TLSCertPath, tlsConfiguration.TLSKeyPath, tlsConfiguration.TLSSkipVerify)
		if err != nil {
			return err
		}

		agentOnDockerEnvironment, err := client.ExecutePingOperation(endpoint.URL, tlsConfig)
		if err != nil {
			return err
		}

		if agentOnDockerEnvironment {
			endpoint.Type = chainid.AgentOnDockerEnvironment
		}
	}

	return endpointService.CreateEndpoint(endpoint)
}

func createUnsecuredEndpoint(endpointURL string, endpointService chainid.EndpointService) error {
	if strings.HasPrefix(endpointURL, "tcp://") {
		_, err := client.ExecutePingOperation(endpointURL, nil)
		if err != nil {
			return err
		}
	}

	endpoint := &chainid.Endpoint{
		Name:            "primary",
		URL:             endpointURL,
		GroupID:         chainid.EndpointGroupID(1),
		Type:            chainid.DockerEnvironment,
		TLSConfig:       chainid.TLSConfiguration{},
		AuthorizedUsers: []chainid.UserID{},
		AuthorizedTeams: []chainid.TeamID{},
		Extensions:      []chainid.EndpointExtension{},
	}

	return endpointService.CreateEndpoint(endpoint)
}

func initEndpoint(flags *chainid.CLIFlags, endpointService chainid.EndpointService) error {
	if *flags.EndpointURL == "" {
		return nil
	}

	endpoints, err := endpointService.Endpoints()
	if err != nil {
		return err
	}

	if len(endpoints) > 0 {
		log.Println("Instance already has defined endpoints. Skipping the endpoint defined via CLI.")
		return nil
	}

	if *flags.TLS || *flags.TLSSkipVerify {
		return createTLSSecuredEndpoint(flags, endpointService)
	}
	return createUnsecuredEndpoint(*flags.EndpointURL, endpointService)
}

func main() {
	flags := initCLI()

	fileService := initFileService(*flags.Data)

	store := initStore(*flags.Data)
	defer store.Close()

	jwtService := initJWTService(!*flags.NoAuth)

	cryptoService := initCryptoService()

	digitalSignatureService := initDigitalSignatureService()

	ldapService := initLDAPService()

	gitService := initGitService()

	authorizeEndpointMgmt := initEndpointWatcher(store.EndpointService, *flags.ExternalEndpoints, *flags.SyncInterval)

	err := initKeyPair(fileService, digitalSignatureService)
	if err != nil {
		log.Fatal(err)
	}

	stackManager, err := initStackManager(*flags.Assets, *flags.Data, digitalSignatureService, fileService)
	if err != nil {
		log.Fatal(err)
	}

	err = initSettings(store.SettingsService, flags)
	if err != nil {
		log.Fatal(err)
	}

	err = initDockerHub(store.DockerHubService)
	if err != nil {
		log.Fatal(err)
	}

	applicationStatus := initStatus(authorizeEndpointMgmt, flags)

	err = initEndpoint(flags, store.EndpointService)
	if err != nil {
		log.Fatal(err)
	}

	adminPasswordHash := ""
	if *flags.AdminPasswordFile != "" {
		content, err := fileService.GetFileContent(*flags.AdminPasswordFile)
		if err != nil {
			log.Fatal(err)
		}
		adminPasswordHash, err = cryptoService.Hash(content)
		if err != nil {
			log.Fatal(err)
		}
	} else if *flags.AdminPassword != "" {
		adminPasswordHash = *flags.AdminPassword
	}

	if adminPasswordHash != "" {
		users, err := store.UserService.UsersByRole(chainid.AdministratorRole)
		if err != nil {
			log.Fatal(err)
		}

		if len(users) == 0 {
			log.Printf("Creating admin user with password hash %s", adminPasswordHash)
			user := &chainid.User{
				Username: "admin",
				Role:     chainid.AdministratorRole,
				Password: adminPasswordHash,
			}
			err := store.UserService.CreateUser(user)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Println("Instance already has an administrator user defined. Skipping admin password related flags.")
		}
	}

	var server chainid.Server = &http.Server{
		Status:                 applicationStatus,
		BindAddress:            *flags.Addr,
		AssetsPath:             *flags.Assets,
		AuthDisabled:           *flags.NoAuth,
		EndpointManagement:     authorizeEndpointMgmt,
		UserService:            store.UserService,
		TeamService:            store.TeamService,
		TeamMembershipService:  store.TeamMembershipService,
		EndpointService:        store.EndpointService,
		EndpointGroupService:   store.EndpointGroupService,
		ResourceControlService: store.ResourceControlService,
		SettingsService:        store.SettingsService,
		RegistryService:        store.RegistryService,
		DockerHubService:       store.DockerHubService,
		StackService:           store.StackService,
		StackManager:           stackManager,
		CryptoService:          cryptoService,
		JWTService:             jwtService,
		FileService:            fileService,
		LDAPService:            ldapService,
		GitService:             gitService,
		SignatureService:       digitalSignatureService,
		SSL:                    *flags.SSL,
		SSLCert:                *flags.SSLCert,
		SSLKey:                 *flags.SSLKey,
	}

	log.Printf("Starting Chain Platform %s on %s", chainid.APIVersion, *flags.Addr)
	err = server.Start()
	if err != nil {
		log.Fatal(err)
	}
}
