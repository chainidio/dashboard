package bolt

import (
	"github.com/chainid-io/dashboard"
	"github.com/chainid-io/dashboard/bolt/internal"

	"github.com/boltdb/bolt"
)

// EndpointGroupService represents a service for managing endpoint groups.
type EndpointGroupService struct {
	store *Store
}

// EndpointGroup returns an endpoint group by ID.
func (service *EndpointGroupService) EndpointGroup(ID chainid.EndpointGroupID) (*chainid.EndpointGroup, error) {
	var data []byte
	err := service.store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(endpointGroupBucketName))
		value := bucket.Get(internal.Itob(int(ID)))
		if value == nil {
			return chainid.ErrEndpointGroupNotFound
		}

		data = make([]byte, len(value))
		copy(data, value)
		return nil
	})
	if err != nil {
		return nil, err
	}

	var endpointGroup chainid.EndpointGroup
	err = internal.UnmarshalEndpointGroup(data, &endpointGroup)
	if err != nil {
		return nil, err
	}
	return &endpointGroup, nil
}

// EndpointGroups return an array containing all the endpoint groups.
func (service *EndpointGroupService) EndpointGroups() ([]chainid.EndpointGroup, error) {
	var endpointGroups = make([]chainid.EndpointGroup, 0)
	err := service.store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(endpointGroupBucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var endpointGroup chainid.EndpointGroup
			err := internal.UnmarshalEndpointGroup(v, &endpointGroup)
			if err != nil {
				return err
			}
			endpointGroups = append(endpointGroups, endpointGroup)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return endpointGroups, nil
}

// CreateEndpointGroup assign an ID to a new endpoint group and saves it.
func (service *EndpointGroupService) CreateEndpointGroup(endpointGroup *chainid.EndpointGroup) error {
	return service.store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(endpointGroupBucketName))

		id, _ := bucket.NextSequence()
		endpointGroup.ID = chainid.EndpointGroupID(id)

		data, err := internal.MarshalEndpointGroup(endpointGroup)
		if err != nil {
			return err
		}

		err = bucket.Put(internal.Itob(int(endpointGroup.ID)), data)
		if err != nil {
			return err
		}
		return nil
	})
}

// UpdateEndpointGroup updates an endpoint group.
func (service *EndpointGroupService) UpdateEndpointGroup(ID chainid.EndpointGroupID, endpointGroup *chainid.EndpointGroup) error {
	data, err := internal.MarshalEndpointGroup(endpointGroup)
	if err != nil {
		return err
	}

	return service.store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(endpointGroupBucketName))
		err = bucket.Put(internal.Itob(int(ID)), data)
		if err != nil {
			return err
		}
		return nil
	})
}

// DeleteEndpointGroup deletes an endpoint group.
func (service *EndpointGroupService) DeleteEndpointGroup(ID chainid.EndpointGroupID) error {
	return service.store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(endpointGroupBucketName))
		err := bucket.Delete(internal.Itob(int(ID)))
		if err != nil {
			return err
		}
		return nil
	})
}
