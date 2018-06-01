package bolt

import (
	"github.com/boltdb/bolt"
	"github.com/chainid-io/dashboard"
	"github.com/chainid-io/dashboard/bolt/internal"
)

func (m *Migrator) updateResourceControlsToDBVersion2() error {
	legacyResourceControls, err := m.retrieveLegacyResourceControls()
	if err != nil {
		return err
	}

	for _, resourceControl := range legacyResourceControls {
		resourceControl.SubResourceIDs = []string{}
		resourceControl.TeamAccesses = []chainid.TeamResourceAccess{}

		owner, err := m.UserService.User(resourceControl.OwnerID)
		if err != nil {
			return err
		}

		if owner.Role == chainid.AdministratorRole {
			resourceControl.AdministratorsOnly = true
			resourceControl.UserAccesses = []chainid.UserResourceAccess{}
		} else {
			resourceControl.AdministratorsOnly = false
			userAccess := chainid.UserResourceAccess{
				UserID:      resourceControl.OwnerID,
				AccessLevel: chainid.ReadWriteAccessLevel,
			}
			resourceControl.UserAccesses = []chainid.UserResourceAccess{userAccess}
		}

		err = m.ResourceControlService.CreateResourceControl(&resourceControl)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) updateEndpointsToDBVersion2() error {
	legacyEndpoints, err := m.EndpointService.Endpoints()
	if err != nil {
		return err
	}

	for _, endpoint := range legacyEndpoints {
		endpoint.AuthorizedTeams = []chainid.TeamID{}
		err = m.EndpointService.UpdateEndpoint(endpoint.ID, &endpoint)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) retrieveLegacyResourceControls() ([]chainid.ResourceControl, error) {
	legacyResourceControls := make([]chainid.ResourceControl, 0)
	err := m.store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("containerResourceControl"))
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var resourceControl chainid.ResourceControl
			err := internal.UnmarshalResourceControl(v, &resourceControl)
			if err != nil {
				return err
			}
			resourceControl.Type = chainid.ContainerResourceControl
			legacyResourceControls = append(legacyResourceControls, resourceControl)
		}

		bucket = tx.Bucket([]byte("serviceResourceControl"))
		cursor = bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var resourceControl chainid.ResourceControl
			err := internal.UnmarshalResourceControl(v, &resourceControl)
			if err != nil {
				return err
			}
			resourceControl.Type = chainid.ServiceResourceControl
			legacyResourceControls = append(legacyResourceControls, resourceControl)
		}

		bucket = tx.Bucket([]byte("volumeResourceControl"))
		cursor = bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var resourceControl chainid.ResourceControl
			err := internal.UnmarshalResourceControl(v, &resourceControl)
			if err != nil {
				return err
			}
			resourceControl.Type = chainid.VolumeResourceControl
			legacyResourceControls = append(legacyResourceControls, resourceControl)
		}
		return nil
	})
	return legacyResourceControls, err
}
