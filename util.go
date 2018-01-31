package targetutil

import "github.com/fsm/fsm"

// GetStateMap converts a fsm.StateMachine into a fsm.StateMap
func GetStateMap(stateMachine fsm.StateMachine) fsm.StateMap {
	stateMap := make(fsm.StateMap, 0)
	for _, buildState := range stateMachine {
		stateMap[buildState(nil, nil).Slug] = buildState
	}
	return stateMap
}

// Step performs a single step through a StateMachine
func Step(uuid, input string, store fsm.Store, emitter fsm.Emitter, stateMap fsm.StateMap) {
	// Get Traverser
	newTraverser := false
	traverser, err := store.FetchTraverser(uuid)
	if err != nil {
		traverser, _ = store.CreateTraverser(uuid)
		traverser.SetCurrentState("start")
		newTraverser = true
	}

	// Get current state
	currentState := stateMap[traverser.CurrentState()](emitter, traverser)
	if newTraverser {
		performEntryAction(currentState, emitter, traverser, stateMap)
	}

	// Transition
	newState := currentState.Transition(input)
	if newState != nil {
		traverser.SetCurrentState(newState.Slug)
		performEntryAction(newState, emitter, traverser, stateMap)
	} else {
		currentState.ReentryAction()
	}
}

func performEntryAction(state *fsm.State, emitter fsm.Emitter, traverser fsm.Traverser, stateMap fsm.StateMap) error {
	err := state.EntryAction()
	if err != nil {
		return err
	}

	// If we switch states in EntryAction, we want to perform
	// the next states EntryAction
	currentState := traverser.CurrentState()
	if currentState != state.Slug {
		shiftedState := stateMap[currentState](emitter, traverser)
		performEntryAction(shiftedState, emitter, traverser, stateMap)
	}
	return nil
}
