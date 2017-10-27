package access_control_service

import (
	"log"
	"github.com/gocms-io/gocms/init/repository"
	"github.com/gocms-io/gocms/domain/acl/permissions/permission_model"
)

type IAclService interface {
	RefreshPermissionsCache() error
	GetPermissions() map[string]permission_model.Permission
	IsAuthorized(string, int) bool
}

type AclService struct {
	Permissions       map[string]permission_model.Permission
	RepositoriesGroup *repository.RepositoriesGroup
}

func DefaultAclService(rg *repository.RepositoriesGroup) *AclService {
	aclService := &AclService{
		RepositoriesGroup: rg,
	}

	return aclService

}

func (as *AclService) RefreshPermissionsCache() error {

	// get all permissions
	permissions, err := as.RepositoriesGroup.PermissionsRepository.GetAll()
	if err != nil {
		log.Fatalf("Fatal - Error caching permissions: %s\n", err.Error())
		return err
	}

	permissionsCache := make(map[string]permission_model.Permission, len(*permissions))
	// cache permissions
	for _, permission := range *permissions {
		permissionsCache[permission.Name] = permission
	}

	as.Permissions = permissionsCache
	return nil
}

func (as *AclService) GetPermissions() map[string]permission_model.Permission {
	return as.Permissions
}

func (as *AclService) IsAuthorized(permission string, userId int) bool {
	// get user permissions mapped to user
	activePermissions, err := as.RepositoriesGroup.PermissionsRepository.GetUserPermissions(userId)
	if err != nil {
		log.Printf("Error getting users permissions: %s\n", err.Error())
		return false
	}

	// loop over permissions and see if they match the request one
	for _, permId := range *activePermissions {
		if permId == as.Permissions[permission].Id {
			return true
		}
	}
	return false
}
