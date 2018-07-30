package school

import (
	"fmt"

	"github.com/windmilleng/pets/internal/proc"
	"github.com/windmilleng/pets/internal/service"
)

// A Pets service provider is a function that:
// 1) Takes as input all the services it depends on (expressed as processes), and
// 2) Starts one new process that exposes a service.
//
// TODO(nick): Right now, Pets assumes a 1:1 mapping between a process PID
// and a network port. But obviously, this isn't true for all cases. One process
// might listen on multiple ports. And one process might start multiple sub-processes
// that each listen on their own ports.
//
// It's not clear yet how we should model this situation in the API. Maybe
// there should be a PetsProc for each unique (PID, Port) tuple. Or maybe PetsProc
// should have multiple ports.
type Provider func([]proc.PetsProc) (proc.PetsProc, error)

type ProviderSpec struct {
	inputNames []service.Name
	provider   Provider

	// A human-readable position that tells the user where the provider was declared.
	position string
}

type PetSchool struct {
	procfs proc.ProcFS

	// Each ProviderSpec is mapped by its output name.
	providers map[service.Key]ProviderSpec
}

func NewPetSchool(procfs proc.ProcFS) *PetSchool {
	return &PetSchool{
		procfs:    procfs,
		providers: make(map[service.Key]ProviderSpec),
	}
}

func (s *PetSchool) AddProvider(key service.Key, provider Provider, deps []service.Name, position string) error {
	existing, exists := s.providers[key]
	if exists {
		return fmt.Errorf("Duplicate provider for service %q, tier %q\nFirst:  %s\nSecond: %s",
			key.Name, key.Tier, existing.position, position)
	}
	spec := ProviderSpec{
		inputNames: deps,
		provider:   provider,
		position:   position,
	}
	s.providers[key] = spec
	return nil
}

// Find all the services that are currently "healthy". This means they have:
// 1) a live process,
// 2) with an exposed host + port,
// 3) associated with a service key
// In the future, this might include a user-specified health check.
func (s *PetSchool) healthyServices() (map[service.Key]proc.PetsProc, error) {
	result := make(map[service.Key]proc.PetsProc)

	// TODO(nick): Get this info from ProcFS. This will be easier once 'pets list' is ready

	// TODO(nick): Who's responsible for writing service.Key and service.Tier to procfs?

	return result, nil
}

// Bring up the service with the given key, including all its dependencies.
func (s *PetSchool) Up(key service.Key) (proc.PetsProc, error) {
	services, err := s.healthyServices()
	if err != nil {
		return proc.PetsProc{}, err
	}
	return s.up(key, services)
}

func (s *PetSchool) up(key service.Key, petsUp map[service.Key]proc.PetsProc) (proc.PetsProc, error) {
	alreadyRunning, ok := petsUp[key]
	if ok {
		return alreadyRunning, nil
	}

	providerSpec, ok := s.providers[key]
	if !ok {
		return proc.PetsProc{}, fmt.Errorf("No provider found for service %q, tier %q", key.Name, key.Tier)
	}

	// Make sure all the dependencies are up. For simplicity, we bring them up in serial.
	inputProcs := make([]proc.PetsProc, len(providerSpec.inputNames))
	for i, input := range providerSpec.inputNames {
		// Right now, we assume that if {Name: frontend, Tier: local} depends on "backend",
		// then it should bring up {Name: backend, Tier: local}. We need to think
		// more about how users specify service graphs in PETS, either as command-line
		// overrides or as Guice-style modules.
		inputKey := service.Key{
			Name: input,
			Tier: key.Tier,
		}

		inputProc, err := s.up(inputKey, petsUp)
		if err != nil {
			return proc.PetsProc{}, fmt.Errorf(
				"Service %q depends on service %q, but %q failed:\n%v", key.Name, input, input, err)
		}

		inputProcs[i] = inputProc
	}

	// All the inputs are ready! Let the user take over from here.
	result, err := providerSpec.provider(inputProcs)
	if err != nil {
		return proc.PetsProc{}, err
	}
	petsUp[key] = result
	return result, nil
}
