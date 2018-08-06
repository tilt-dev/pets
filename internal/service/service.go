// Data structures for identifying services (processes with network host/port)
package service

import (
	"fmt"
	"regexp"
)

// A name for the service, e.g., "my-frontend".
//
// In a 'pets up', only one service of each name is allowed.
type Name string

// A tier for the service, e.g., "local", "prod", "staging"
//
// Kubernetes models this with a much more free-form "add arbitrary tags that
// you can select services on". This is deliberately more limited.
//
type Tier string

// The primary key for a particular way of running a service (e.g., "my-frontend
// running on local")
//
// Only one provider of each Key=(Name, Tier) tuple is allowed in a Petsfile
// and all the Petsfiles it loads.
type Key struct {
	Name `json:",omitempty"`
	Tier `json:",omitempty"`
}

func NewKey(name Name, tier Tier) Key {
	return Key{Name: name, Tier: tier}
}

var validNameMatcher = regexp.MustCompile("^[a-zA-Z_-][a-zA-Z0-9_-]*$")
var validTierMatcher = validNameMatcher

func (n Name) Validate() error {
	if !validNameMatcher.MatchString(string(n)) {
		return fmt.Errorf("Invalid service name %q. Service names must match regexp %q", n, validNameMatcher)
	}
	return nil
}

func (t Tier) Validate() error {
	if !validTierMatcher.MatchString(string(t)) {
		return fmt.Errorf("Invalid service tier %q. Service names must match regexp %q", t, validTierMatcher)
	}
	return nil
}

func (k Key) Validate() error {
	err := k.Name.Validate()
	if err != nil {
		return err
	}
	return k.Tier.Validate()
}

func (k Key) String() string {
	return fmt.Sprintf("%s-%s", k.Name, k.Tier)
}
