// Package secrets validates sensitive configuration inputs without storing
// private material.
package secrets

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

var supportedPublicKeyTypes = map[string]struct{}{
	ssh.KeyAlgoED25519:  {},
	ssh.KeyAlgoECDSA256: {},
	ssh.KeyAlgoECDSA384: {},
	ssh.KeyAlgoECDSA521: {},
	ssh.KeyAlgoRSA:      {},
}

// ValidateSSHPublicKey validates one authorized_keys-style SSH public key.
func ValidateSSHPublicKey(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("SSH public key is empty")
	}

	key, _, options, rest, err := ssh.ParseAuthorizedKey([]byte(value))
	if err != nil {
		return fmt.Errorf("parse SSH public key: %w", err)
	}
	if len(options) != 0 {
		return fmt.Errorf("SSH public key options are not supported")
	}
	if len(bytes.TrimSpace(rest)) != 0 {
		return fmt.Errorf("expected one SSH public key")
	}
	if _, ok := supportedPublicKeyTypes[key.Type()]; !ok {
		return fmt.Errorf("unsupported SSH public key type %q", key.Type())
	}

	return nil
}
