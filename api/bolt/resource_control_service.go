package bolt

import (
	"github.com/chainid-io/dashboard"
	"github.com/chainid-io/dashboard/bolt/internal"

	"github.com/boltdb/bolt"
)

// ResourceControlService represents a service for managing resource controls.
type ResourceControlService struct {
	store *Store
}

// ResourceControl returns a ResourceControl object by ID
func (service *ResourceControlService) ResourceControl(ID chainid.ResourceControlID) (*chainid.ResourceControl, error) {
	var data []byte
	err := service.store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(resourceControlBucketName))
		value := bucket.Get(internal.Itob(int(ID)))
		if value == nil {
			return chainid.ErrResourceControlNotFound
		}

		data = make([]byte, len(value))
		copy(data, value)
		return nil
	})
	if err != nil {
		return nil, err
	}

	var resourceControl chainid.ResourceControl
	err = internal.UnmarshalResourceControl(data, &resourceControl)
	if err != nil {
		return nil, err
	}
	return &resourceControl, nil
}

// ResourceControlByResourceID returns a ResourceControl object by checking if the resourceID is equal
// to the main ResourceID or in SubResourceIDs
func (service *ResourceControlService) ResourceControlByResourceID(resourceID string) (*chainid.ResourceControl, error) {
	var resourceControl *chainid.ResourceControl

	err := service.store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(resourceControlBucketName))
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var rc chainid.ResourceControl
			err := internal.UnmarshalResourceControl(v, &rc)
			if err != nil {
				return err
			}
			if rc.ResourceID == resourceID {
				resourceControl = &rc
			}
			for _, subResourceID := range rc.SubResourceIDs {
				if subResourceID == resourceID {
					resourceControl = &rc
				}
			}
		}

		if resourceControl == nil {
			return chainid.ErrResourceControlNotFound
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resourceControl, nil
}

// ResourceControls returns all the ResourceControl objects
func (service *ResourceControlService) ResourceControls() ([]chainid.ResourceControl, error) {
	var rcs = make([]chainid.ResourceControl, 0)
	err := service.store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(resourceControlBucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var resourceControl chainid.ResourceControl
			err := internal.UnmarshalResourceControl(v, &resourceControl)
			if err != nil {
				return err
			}
			rcs = append(rcs, resourceControl)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return rcs, nil
}

// CreateResourceControl creates a new ResourceControl object
func (service *ResourceControlService) CreateResourceControl(resourceControl *chainid.ResourceControl) error {
	return service.store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(resourceControlBucketName))
		id, _ := bucket.NextSequence()
		resourceControl.ID = chainid.ResourceControlID(id)
		data, err := internal.MarshalResourceControl(resourceControl)
		if err != nil {
			return err
		}

		err = bucket.Put(internal.Itob(int(resourceControl.ID)), data)
		if err != nil {
			return err
		}
		return nil
	})
}

// UpdateResourceControl saves a ResourceControl object.
func (service *ResourceControlService) UpdateResourceControl(ID chainid.ResourceControlID, resourceControl *chainid.ResourceControl) error {
	data, err := internal.MarshalResourceControl(resourceControl)
	if err != nil {
		return err
	}

	return service.store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(resourceControlBucketName))
		err = bucket.Put(internal.Itob(int(ID)), data)

		if err != nil {
			return err
		}
		return nil
	})
}

// DeleteResourceControl deletes a ResourceControl object by ID
func (service *ResourceControlService) DeleteResourceControl(ID chainid.ResourceControlID) error {
	return service.store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(resourceControlBucketName))
		err := bucket.Delete(internal.Itob(int(ID)))
		if err != nil {
			return err
		}
		return nil
	})
}
