package permission

import (
	"io/ioutil"
	"json"
	"path"
	"os"
)

const (
	FileNameGroups = "groups.json"
	FileNameUsers  = "users.json"
)

var (
	ErrFileGroups = os.NewError("Couldn't find groups.json")
	ErrFileUsers  = os.NewError("Couldn't find users.json")
)

// This is a permission system based on two json files "groups.json" and "users.json".
// It has one world support.
type JsonPermission struct {
	groups Groups
	users  Users
	// TODO: Implement cache. No looking up permissions in groups
}


// The expected folder structure is:
// ./groups.json
// ./users.yml
func LoadJsonPermission(folder string) (*JsonPermission, os.Error) {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return nil, err
	}
	var fileNameUsers string
	var fileNameGroups string
	for _, file := range files {
		fileName := path.Base(file.Name)
		if fileName == FileNameGroups {
			fileNameGroups = file.Name
		} else if fileName == FileNameUsers {
			fileNameUsers = file.Name
		}
	}
	if len(fileNameUsers) == 0 {
		return nil, ErrFileUsers
	} else if len(fileNameGroups) == 0 {
		return nil, ErrFileGroups
	}
	// Load users
	var bytesUsers []byte
	bytesUsers, err = ioutil.ReadFile(fileNameUsers)
	if err != nil {
		return nil, err
	}
	var users Users
	err = json.Unmarshal(bytesUsers, users)
	if err != nil {
		return nil, err
	}
	// Load groups
	var bytesGroups []byte
	bytesGroups, err = ioutil.ReadFile(fileNameGroups)
	if err != nil {
		return nil, err
	}
	var groups Groups
	err = json.Unmarshal(bytesGroups, groups)
	if err != nil {
		return nil, err
	}
	return &JsonPermission{groups: groups, users: users}, nil
}

// Implementation of IPermissions
func (p *JsonPermission) PlayerPermissions(username string) IUserPermissions {
	if user, ok := p.users[username]; ok {
		// TODO Add group support
		// TODO Cache CachedUser in JsonPermission
		return &CachedUser{permissions: user.Permissions}
	}
	return nil
}

// A JsonPermission user with chached permissions.
type CachedUser struct {
	permissions []string
}

// Implementation of IUserPermissions
func (u *CachedUser) Has(node string) bool {
	for _, p := range u.permissions {
		if p == node {
			return true
		}
	}
	return false
}
