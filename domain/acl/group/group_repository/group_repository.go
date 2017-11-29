package group_repository

import (
	"github.com/gocms-io/gocms/domain/acl/group/group_model"
	"github.com/gocms-io/gocms/utility/log"
	"github.com/jmoiron/sqlx"
)

type IGroupsRepository interface {
	Add(*group_model.Group) error
	Delete(int64) error
	GetAll() (*[]group_model.Group, error)

	GetUserGroups(userId int64) ([]*group_model.Group, error)
	AddUserToGroup(userId int64, groupId int64) error
	RemoveUserFromGroup(userId int64, groupId int64) error
}

type GroupsRepository struct {
	database *sqlx.DB
}

// DefaultGroupsRepository creates a default groups repository.
func DefaultGroupsRepository(dbx *sqlx.DB) *GroupsRepository {
	groupsRepository := &GroupsRepository{
		database: dbx,
	}

	return groupsRepository
}

// Add adds group to database
func (pr *GroupsRepository) Add(group *group_model.Group) error {

	// insert user
	result, err := pr.database.NamedExec(`
	INSERT INTO gocms_groups (name, description) VALUES (:name, :description)
	`, group)
	if err != nil {
		log.Errorf("Error adding group to db: %s\n", err.Error())
		return err
	}
	id, _ := result.LastInsertId()
	group.Id = id

	return nil

}

// Delete deletes a user group via groupId
func (pr *GroupsRepository) Delete(groupId int64) error {

	_, err := pr.database.NamedExec(`
	DELETE FROM gocms_groups WHERE id=:id
	`, map[string]interface{}{"id": groupId})
	if err != nil {
		log.Errorf("Error deleting group %v from database: %s\n", groupId, err.Error())
		return err
	}

	return nil
}

// GetAll get all groups
func (pr *GroupsRepository) GetAll() (*[]group_model.Group, error) {
	var groups []group_model.Group
	err := pr.database.Select(&groups, "SELECT * FROM gocms_groups")
	if err != nil {
		log.Errorf("Error getting groups from database: %s\n", err.Error())
		return nil, err
	}
	return &groups, nil
}

// GetUserGroups get groups assigned to a given user via userId
func (pr *GroupsRepository) GetUserGroups(userId int64) ([]*group_model.Group, error) {
	var userGroups []*group_model.Group
	err := pr.database.Select(&userGroups, `
	SELECT groupId as id, name, description
	FROM (
		SELECT groupId from gocms_users_to_groups
		WHERE userId = ?
	) as groupIds
	JOIN gocms_groups as grps
	ON groupIds.groupId = grps.id
	`, userId)
	if err != nil {
		log.Errorf("Error getting all groups for user %v from database: %s\n", userId, err.Error())
		return nil, err
	}
	return userGroups, nil
}

// AddUserToGroup adds a user to the group via userId and groupId
func (pr *GroupsRepository) AddUserToGroup(userId int64, groupId int64) error {

	// insert user
	_, err := pr.database.NamedExec(`
	INSERT INTO gocms_users_to_groups (userId, groupId) VALUES (:userId, :groupId)
	`, map[string]interface{}{"userId": userId, "groupId": groupId})
	if err != nil {
		log.Errorf("Error adding user %v to group %v: %s\n", userId, groupId, err.Error())
		return err
	}
	return nil

}

// RemoveUserFromGroup removes a user from the group via userId and groupId
func (pr *GroupsRepository) RemoveUserFromGroup(userId int64, groupId int64) error {

	_, err := pr.database.NamedExec(`
	DELETE FROM gocms_users_to_groups
	WHERE userId=:userId
	AND groupId=:groupId
	`, map[string]interface{}{"userId": userId, "groupId": groupId})
	if err != nil {
		log.Errorf("Error deleting user %v to group %v: %s\n", userId, groupId, err.Error())
		return err
	}

	return nil
}
