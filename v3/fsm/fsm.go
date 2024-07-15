package fsm

type Transition[S comparable, E comparable, A any] struct {
	State    S
	Event    E
	Action   A
	NewState S
	Fallback bool
}

func Input[S comparable, E comparable, A any](
	event E, state S, transitions []Transition[S, E, A],
) (A, S) {
	for _, transition := range transitions {
		if transition.State == state &&
			(transition.Event == event || transition.Fallback) {
			return transition.Action, transition.NewState
		}
	}
	panic("No transition found")
}
