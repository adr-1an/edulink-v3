package schools

import "testing"

func TestPortalUserActiveStatePermission(t *testing.T) {
	tests := []struct {
		name       string
		wasActive  bool
		isActive   bool
		permission Permission
		changed    bool
	}{
		{name: "unchanged inactive", wasActive: false, isActive: false},
		{name: "activate", wasActive: false, isActive: true, permission: PermissionPortalUserActivate, changed: true},
		{name: "deactivate", wasActive: true, isActive: false, permission: PermissionPortalUserDeactivate, changed: true},
		{name: "unchanged active", wasActive: true, isActive: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permission, changed := PortalUserActiveStatePermission(tt.wasActive, tt.isActive)
			if permission != tt.permission || changed != tt.changed {
				t.Fatalf("got (%q, %t), want (%q, %t)", permission, changed, tt.permission, tt.changed)
			}
		})
	}
}
