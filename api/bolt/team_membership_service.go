package bolt

import (
	"github.com/chainid-io/dashboard"
	"github.com/chainid-io/dashboard/bolt/internal"

	"github.com/boltdb/bolt"
)

// TeamMembershipService represents a service for managing TeamMembership objects.
type TeamMembershipService struct {
	store *Store
}

// TeamMembership returns a TeamMembership object by ID
func (service *TeamMembershipService) TeamMembership(ID chainid.TeamMembershipID) (*chainid.TeamMembership, error) {
	var data []byte
	err := service.store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(teamMembershipBucketName))
		value := bucket.Get(internal.Itob(int(ID)))
		if value == nil {
			return chainid.ErrTeamMembershipNotFound
		}

		data = make([]byte, len(value))
		copy(data, value)
		return nil
	})
	if err != nil {
		return nil, err
	}

	var membership chainid.TeamMembership
	err = internal.UnmarshalTeamMembership(data, &membership)
	if err != nil {
		return nil, err
	}
	return &membership, nil
}

// TeamMemberships return an array containing all the TeamMembership objects.
func (service *TeamMembershipService) TeamMemberships() ([]chainid.TeamMembership, error) {
	var memberships = make([]chainid.TeamMembership, 0)
	err := service.store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(teamMembershipBucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var membership chainid.TeamMembership
			err := internal.UnmarshalTeamMembership(v, &membership)
			if err != nil {
				return err
			}
			memberships = append(memberships, membership)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return memberships, nil
}

// TeamMembershipsByUserID return an array containing all the TeamMembership objects where the specified userID is present.
func (service *TeamMembershipService) TeamMembershipsByUserID(userID chainid.UserID) ([]chainid.TeamMembership, error) {
	var memberships = make([]chainid.TeamMembership, 0)
	err := service.store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(teamMembershipBucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var membership chainid.TeamMembership
			err := internal.UnmarshalTeamMembership(v, &membership)
			if err != nil {
				return err
			}
			if membership.UserID == userID {
				memberships = append(memberships, membership)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return memberships, nil
}

// TeamMembershipsByTeamID return an array containing all the TeamMembership objects where the specified teamID is present.
func (service *TeamMembershipService) TeamMembershipsByTeamID(teamID chainid.TeamID) ([]chainid.TeamMembership, error) {
	var memberships = make([]chainid.TeamMembership, 0)
	err := service.store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(teamMembershipBucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var membership chainid.TeamMembership
			err := internal.UnmarshalTeamMembership(v, &membership)
			if err != nil {
				return err
			}
			if membership.TeamID == teamID {
				memberships = append(memberships, membership)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return memberships, nil
}

// UpdateTeamMembership saves a TeamMembership object.
func (service *TeamMembershipService) UpdateTeamMembership(ID chainid.TeamMembershipID, membership *chainid.TeamMembership) error {
	data, err := internal.MarshalTeamMembership(membership)
	if err != nil {
		return err
	}

	return service.store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(teamMembershipBucketName))
		err = bucket.Put(internal.Itob(int(ID)), data)

		if err != nil {
			return err
		}
		return nil
	})
}

// CreateTeamMembership creates a new TeamMembership object.
func (service *TeamMembershipService) CreateTeamMembership(membership *chainid.TeamMembership) error {
	return service.store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(teamMembershipBucketName))

		id, _ := bucket.NextSequence()
		membership.ID = chainid.TeamMembershipID(id)

		data, err := internal.MarshalTeamMembership(membership)
		if err != nil {
			return err
		}

		err = bucket.Put(internal.Itob(int(membership.ID)), data)
		if err != nil {
			return err
		}
		return nil
	})
}

// DeleteTeamMembership deletes a TeamMembership object.
func (service *TeamMembershipService) DeleteTeamMembership(ID chainid.TeamMembershipID) error {
	return service.store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(teamMembershipBucketName))
		err := bucket.Delete(internal.Itob(int(ID)))
		if err != nil {
			return err
		}
		return nil
	})
}

// DeleteTeamMembershipByUserID deletes all the TeamMembership object associated to a UserID.
func (service *TeamMembershipService) DeleteTeamMembershipByUserID(userID chainid.UserID) error {
	return service.store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(teamMembershipBucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var membership chainid.TeamMembership
			err := internal.UnmarshalTeamMembership(v, &membership)
			if err != nil {
				return err
			}
			if membership.UserID == userID {
				err := bucket.Delete(internal.Itob(int(membership.ID)))
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// DeleteTeamMembershipByTeamID deletes all the TeamMembership object associated to a TeamID.
func (service *TeamMembershipService) DeleteTeamMembershipByTeamID(teamID chainid.TeamID) error {
	return service.store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(teamMembershipBucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var membership chainid.TeamMembership
			err := internal.UnmarshalTeamMembership(v, &membership)
			if err != nil {
				return err
			}
			if membership.TeamID == teamID {
				err := bucket.Delete(internal.Itob(int(membership.ID)))
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}
