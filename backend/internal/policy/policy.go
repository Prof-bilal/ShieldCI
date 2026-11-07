package policy

import (
	"errors"
	"fmt"
	"strings"
)

// Policy is the smallest ShieldCI contract shared by the control plane and
// execution adapter. Empty access lists are intentional: the default is deny.
type Policy struct {
	Version     string      `json:"version"`
	Name        string      `json:"name"`
	Filesystem  Filesystem  `json:"filesystem"`
	Network     Network     `json:"network"`
	Environment Environment `json:"environment"`
}

type Filesystem struct {
	Read  []string `json:"read"`
	Write []string `json:"write"`
}

type Network struct {
	Allow []string `json:"allow"`
}

type Environment struct {
	Allow []string `json:"allow"`
}

var ErrInvalid = errors.New("invalid policy")

func (p Policy) Validate() error {
	if strings.TrimSpace(p.Version) == "" {
		return fmt.Errorf("%w: version is required", ErrInvalid)
	}
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalid)
	}
	if strings.ContainsAny(p.Name, "\r\n") {
		return fmt.Errorf("%w: name contains a newline", ErrInvalid)
	}
	for _, path := range append(append([]string{}, p.Filesystem.Read...), p.Filesystem.Write...) {
		if strings.TrimSpace(path) == "" || strings.HasPrefix(path, "~") {
			return fmt.Errorf("%w: filesystem paths must be explicit", ErrInvalid)
		}
	}
	for _, host := range p.Network.Allow {
		if strings.TrimSpace(host) == "" || strings.ContainsAny(host, "/\\") {
			return fmt.Errorf("%w: network entries must be hostnames", ErrInvalid)
		}
	}
	for _, key := range p.Environment.Allow {
		if strings.TrimSpace(key) == "" || strings.ContainsAny(key, "= \t\r\n") {
			return fmt.Errorf("%w: environment entries must be variable names", ErrInvalid)
		}
	}
	return nil
}
