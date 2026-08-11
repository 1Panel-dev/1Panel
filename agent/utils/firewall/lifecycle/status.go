package lifecycle

import (
	"errors"
	"sync"
)

type State struct {
	Name     string
	IsActive bool
}

type Status struct {
	State
	Version string
}

func LoadState(client Client) (State, error) {
	state := State{Name: client.Name()}
	var err error
	state.IsActive, err = client.Status()
	return state, err
}

func LoadStatus(client Client) (Status, error) {
	status := Status{Version: "-"}
	var versionErr error
	var stateErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		status.Version, versionErr = client.Version()
	}()
	go func() {
		defer wg.Done()
		status.State, stateErr = LoadState(client)
	}()
	wg.Wait()
	return status, errors.Join(versionErr, stateErr)
}
