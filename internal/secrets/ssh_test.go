package secrets

import "testing"

const examplePublicKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJ0EkzLnXFTKfQXv2bseD0jLxoc+R6NxRHWqnY15T9i6 example-only@bootwrangler.invalid"

func TestValidateSSHPublicKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:  "valid ed25519 key",
			value: examplePublicKey,
		},
		{
			name:    "empty key",
			value:   "",
			wantErr: true,
		},
		{
			name:    "malformed key",
			value:   "ssh-ed25519 not-base64",
			wantErr: true,
		},
		{
			name:    "authorized key options",
			value:   "command=\"false\" " + examplePublicKey,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateSSHPublicKey(test.value)
			if test.wantErr && err == nil {
				t.Fatal("ValidateSSHPublicKey() error = nil, want error")
			}
			if !test.wantErr && err != nil {
				t.Fatalf("ValidateSSHPublicKey() error = %v, want nil", err)
			}
		})
	}
}
