package bolt

import "github.com/chainid-io/dashboard"

func (m *Migrator) updateSettingsToDBVersion3() error {
	legacySettings, err := m.SettingsService.Settings()
	if err != nil {
		return err
	}

	legacySettings.AuthenticationMethod = chainid.AuthenticationInternal
	legacySettings.LDAPSettings = chainid.LDAPSettings{
		TLSConfig: chainid.TLSConfiguration{},
		SearchSettings: []chainid.LDAPSearchSettings{
			chainid.LDAPSearchSettings{},
		},
	}

	err = m.SettingsService.StoreSettings(legacySettings)
	if err != nil {
		return err
	}

	return nil
}
