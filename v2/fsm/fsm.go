package fsm

const ANY = -1

type TransitionFunction func(state int, input interface{}, output interface{}, data interface{})

type Matcher interface {
	Matches(input interface{}) bool
}

type State struct {
	Input    Matcher
	NewState int
	Action   TransitionFunction
}

func match(input interface{}, states []State) *State {
	var state *State
	for _, v := range states {
		if v.Input.Matches(input) {
			state = &v
			break
		}
	}
	return state
}

func Transition(stateTable map[int][]State, currentState int, input interface{}, output interface{}, data interface{}) int {
	newState := currentState
	state := match(input, stateTable[currentState])
	if state == nil {
		state = match(input, stateTable[ANY])
	}
	if state != nil {
		if state.Action != nil {
			state.Action(currentState, input, output, data)
		}
		if state.NewState != ANY {
			newState = state.NewState
		}
	}
	return newState
}
