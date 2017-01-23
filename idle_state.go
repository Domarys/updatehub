package main

import "time"

type IdleState struct {
	BaseState
	CancellableState

	elapsedTime int
}

func NewIdleState() *IdleState {
	state := &IdleState{
		BaseState:        BaseState{id: EasyFotaStateIdle},
		CancellableState: CancellableState{cancel: make(chan bool)},
	}
	return state
}

func (is *IdleState) Id() EasyFotaState {
	return is.id
}

func (is *IdleState) Cancel() bool {
	return is.CancellableState.Cancel()
}

func (is *IdleState) Handle(fota *EasyFota) State {
	var nextState State

	nextState = is

	go func() {
		for {
			if is.elapsedTime == fota.pollInterval {
				is.elapsedTime = 0
				nextState = NewUpdateCheckState()
				break
			}

			time.Sleep(time.Second)

			is.elapsedTime++
		}

		is.Cancel()
	}()

	is.Wait()

	return nextState
}