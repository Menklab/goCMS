package models

import (
	"time"
)


/**
* @apiDefine UserAdminDisplay
* @apiSuccess (Response) {number} id
* @apiSuccess (Response) {string} fullName
* @apiSuccess (Response) {string} email
* @apiSuccess (Response) {number} gender 1=male, 2=female
* @apiSuccess (Response) {boolean} enabled true is the user is enabled
* @apiSuccess (Response) {boolean} verified true is the user has verified their primary email address
* @apiSuccess (Response) {number} minAge
* @apiSuccess (Response) {number} maxAge
* @apiSuccess (Response) {string} created
*/
type UserAdminDisplay struct {
	Id              int       `json:"id,omitempty"`
	FullName        string    `json:"fullName,omitempty"`
	Email           string    `json:"email,omitempty"`
	Gender          int       `json:"gender,omitempty"`
	Photo           string    `json:"photo,string,omitempty"`
	Enabled         bool      `json:"enabled,omitempty"`
	Verified         bool      `json:"verified,omitempty"`
	MinAge          int       `json:"minAge,omitempty"`
	MaxAge          int       `json:"maxAge,omitempty"`
	Created         time.Time `json:"created,omitempty"`
}

// helper function to get userAdminDisplay from user object
func (user *User) GetUserAdminDisplay() *UserAdminDisplay {
	userAdminDisplay :=  UserAdminDisplay{
		Id: user.Id,
		Email: user.Email,
		FullName: user.FullName,
		Gender: user.Gender,
		Photo: user.Photo,
		Enabled: user.Enabled,
		Verified: user.Verified,
	}
	return &userAdminDisplay
}