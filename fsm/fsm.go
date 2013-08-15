package fsm

type TransitionFunction func(state int, input interface{}, output interface{}, data interface{})

type Matcher interface {
	Matches(input interface{}) bool
}

type State struct {
	Input    Matcher
	NewState int
	Action   TransitionFunction
}

func Transition(stateTable map[int][]State, currentState int, input interface{}, output interface{}, data interface{}) int {
	newState := currentState
	for _, v := range stateTable[currentState] {
		if v.Input.Matches(input) {
			if v.Action != nil {
				v.Action(currentState, input, output, data)
			}
			newState = v.NewState
			break
		}
	}
	return newState
}
