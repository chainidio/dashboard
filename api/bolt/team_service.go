package bolt

import (
	"github.com/chainid-io/dashboard"
	"github.com/chainid-io/dashboard/bolt/internal"

	"github.com/boltdb/bolt"
)

// TeamService represents a service for managing teams.
type TeamService struct {
	store *Store
}

// Team returns a Team by ID
func (service *TeamService) Team(ID chainid.TeamID) (*chainid.Team, error) {
	var data []byte
	err := service.store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(teamBucketName))
		value := bucket.Get(internal.Itob(int(ID)))
		if value == nil {
			return chainid.ErrTeamNotFound
		}

		data = make([]byte, len(value))
		copy(data, value)
		return nil
	})
	if err != nil {
		return nil, err
	}

	var team chainid.Team
	err = internal.UnmarshalTeam(data, &team)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// TeamByName returns a team by name.
func (service *TeamService) TeamByName(name string) (*chainid.Team, error) {
	var team *chainid.Team

	err := service.store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(teamBucketName))
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var t chainid.Team
			err := internal.UnmarshalTeam(v, &t)
			if err != nil {
				return err
			}
			if t.Name == name {
				team = &t
			}
		}

		if team == nil {
			return chainid.ErrTeamNotFound
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return team, nil
}

// Teams return an array containing all the teams.
func (service *TeamService) Teams() ([]chainid.Team, error) {
	var teams = make([]chainid.Team, 0)
	err := service.store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(teamBucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var team chainid.Team
			err := internal.UnmarshalTeam(v, &team)
			if err != nil {
				return err
			}
			teams = append(teams, team)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return teams, nil
}

// UpdateTeam saves a Team.
func (service *TeamService) UpdateTeam(ID chainid.TeamID, team *chainid.Team) error {
	data, err := internal.MarshalTeam(team)
	if err != nil {
		return err
	}

	return service.store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(teamBucketName))
		err = bucket.Put(internal.Itob(int(ID)), data)

		if err != nil {
			return err
		}
		return nil
	})
}

// CreateTeam creates a new Team.
func (service *TeamService) CreateTeam(team *chainid.Team) error {
	return service.store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(teamBucketName))

		id, _ := bucket.NextSequence()
		team.ID = chainid.TeamID(id)

		data, err := internal.MarshalTeam(team)
		if err != nil {
			return err
		}

		err = bucket.Put(internal.Itob(int(team.ID)), data)
		if err != nil {
			return err
		}
		return nil
	})
}

// DeleteTeam deletes a Team.
func (service *TeamService) DeleteTeam(ID chainid.TeamID) error {
	return service.store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(teamBucketName))
		err := bucket.Delete(internal.Itob(int(ID)))
		if err != nil {
			return err
		}
		return nil
	})
}
