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

	// Overrides set on the command-line
	overrides map[service.Name]service.Tier
}

func NewPetSchool(procfs proc.ProcFS) *PetSchool {
	return &PetSchool{
		procfs:    procfs,
		providers: make(map[service.Key]ProviderSpec),
		overrides: make(map[service.Name]service.Tier),
	}
}

func (s *PetSchool) AddOverride(name service.Name, tier service.Tier) {
	s.overrides[name] = tier
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
	procs, err := s.procfs.ProcsFromFS()
	if err != nil {
		return nil, fmt.Errorf("school#healthServices: %v", err)
	}

	for _, p := range procs {
		isHealthyServiceProc := p.ServiceName != "" && p.ServiceTier != "" &&
			p.Hostname != "" && p.Port != 0
		if !isHealthyServiceProc {
			continue
		}

		key := p.ServiceKey()
		result[key] = p
	}

	return result, nil
}

// Bring up the service with the given key, including all its dependencies.
func (s *PetSchool) UpByKey(key service.Key) (proc.PetsProc, error) {
	services, err := s.healthyServices()
	if err != nil {
		return proc.PetsProc{}, err
	}
	return s.up(key, services)
}

// Bring up all the services of a given tier. Returns an error if there are no services in this tier.
func (s *PetSchool) UpByTier(tier service.Tier) ([]proc.PetsProc, error) {
	services, err := s.healthyServices()
	if err != nil {
		return nil, err
	}

	allProviders := s.providers
	keys := make([]service.Key, 0)
	for key, _ := range allProviders {
		if key.Tier == tier {
			keys = append(keys, key)
		}
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("No service providers found for tier: %q", tier)
	}

	procs := make([]proc.PetsProc, len(keys))
	for i, key := range keys {
		proc, err := s.up(key, services)
		if err != nil {
			return nil, err
		}
		procs[i] = proc
	}

	return procs, nil
}

func (s *PetSchool) up(key service.Key, petsUp map[service.Key]proc.PetsProc) (proc.PetsProc, error) {
	overrideTier, hasOverride := s.overrides[key.Name]
	if hasOverride {
		key.Tier = overrideTier
	}

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

	err = s.procfs.ModifyProc(result.WithServiceKey(key))
	if err != nil {
		return proc.PetsProc{}, err
	}

	petsUp[key] = result
	return result, nil
}
