package permission

import (
	"strings"
	"testing"
)

const (
	testUsersJson  = `{
		"agon": {
			"groups": ["admin"],
			"permissions": [
				"server.web.*",
				"admin.*",
				"server.status"
			]
		},
		"huin": {
			"groups": ["admin"],
			"permissions": [
				"server.*",
				"admin.*"
			]
		}
	}`
	testGroupsJson = `{
		"default": {
			"default": true,
			"permissions": [
				"login",
				"user.commands.help",
				"user.commands.kill",
				"user.commands.me",
				"world.build"
			]
		},
		"admin": {
			"inheritance": ["default"],
			"permissions": [
				"login",
				"admin.commands.give",
				"world.*"
			]
		}
	}`
)

func TestJsonPermission(t *testing.T) {
	usersReader := strings.NewReader(testUsersJson)
	groupsReader := strings.NewReader(testGroupsJson)

	perm, err := LoadJsonPermission(usersReader, groupsReader)
	if err != nil {
		t.Fatalf("Error while loading JsonPermission: %s", err)
	}

	type Test struct {
		username    string
		permission  string
		expectedHas bool
	}

	tests := []Test{
		// Check User permissions
		{"agon", "server.status", true},
		// Check User permissions from groups
		{"agon", "admin.commands.give", true},
		// Check if User has no permission
		{"huin", "this.node.does.not.exist.tm", false},
		// Wildcard check
		{"huin", "server.stop", true},
	}

	for i := range tests {
		test := &tests[i]
		result := perm.UserPermissions(test.username).Has(test.permission)
		if test.expectedHas != result {
			if test.expectedHas {
				t.Error("User %s should have node %s", test.username, test.permission)
			} else {
				t.Error("User %s should *not* have node %s", test.username, test.permission)
			}
		}
	}
}
