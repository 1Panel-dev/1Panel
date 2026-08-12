package lifecycle

import (
	"errors"
	"sync"
)

import ()

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
	var state State
	var version string
	var stateErr, versionErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		state, stateErr = LoadState(client)
	}()
	go func() {
		defer wg.Done()
		version, versionErr = client.Version()
	}()
	wg.Wait()
	status.State = state
	if stateErr != nil {
		return status, errors.Join(stateErr, versionErr)
	}
	if !status.IsActive {
		return status, nil
	}
	status.Version = version
	return status, versionErr
}
