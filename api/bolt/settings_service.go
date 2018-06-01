package bolt

import (
	"github.com/chainid-io/dashboard"
	"github.com/chainid-io/dashboard/bolt/internal"

	"github.com/boltdb/bolt"
)

// SettingsService represents a service to manage application settings.
type SettingsService struct {
	store *Store
}

const (
	dbSettingsKey = "SETTINGS"
)

// Settings retrieve the settings object.
func (service *SettingsService) Settings() (*chainid.Settings, error) {
	var data []byte
	err := service.store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(settingsBucketName))
		value := bucket.Get([]byte(dbSettingsKey))
		if value == nil {
			return chainid.ErrSettingsNotFound
		}

		data = make([]byte, len(value))
		copy(data, value)
		return nil
	})
	if err != nil {
		return nil, err
	}

	var settings chainid.Settings
	err = internal.UnmarshalSettings(data, &settings)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// StoreSettings persists a Settings object.
func (service *SettingsService) StoreSettings(settings *chainid.Settings) error {
	return service.store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(settingsBucketName))

		data, err := internal.MarshalSettings(settings)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(dbSettingsKey), data)
		if err != nil {
			return err
		}
		return nil
	})
}
