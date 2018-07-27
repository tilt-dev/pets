// Data structures for identifying services (processes with network host/port)
package service

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
