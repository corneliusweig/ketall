package client

import "testing"

func TestGetResourceScope(t *testing.T) {
	tests := map[string]struct {
		scope         string
		wantCluster   bool
		wantNamespace bool
		wantErr       bool
	}{
		"no scope": {
			scope:         "",
			wantCluster:   true,
			wantNamespace: true,
			wantErr:       false,
		},
		"namespace scope": {
			scope:         "namespace",
			wantCluster:   false,
			wantNamespace: true,
			wantErr:       false,
		},
		"cluster scope": {
			scope:         "cluster",
			wantCluster:   true,
			wantNamespace: false,
			wantErr:       false,
		},
		"unknown scope": {
			scope:         "unknown",
			wantCluster:   false,
			wantNamespace: false,
			wantErr:       true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotCluster, gotNamespace, gotErr := getResourceScope(test.scope)

			if gotCluster != test.wantCluster {
				t.Fatalf("wrong cluster: got %t, want %t", gotCluster, test.wantNamespace)
			}
			if gotNamespace != test.wantNamespace {
				t.Fatalf("wrong namespace: got %t, want %t", gotNamespace, test.wantNamespace)
			}
			if gotErr != nil && !test.wantErr {
				t.Fatalf("unexpected error: %s", gotErr.Error())
			}
			if gotErr == nil && test.wantErr {
				t.Fatal("expected error, got none")
			}
		})
	}
}
