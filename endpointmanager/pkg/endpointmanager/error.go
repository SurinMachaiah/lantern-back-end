package endpointmanager

type ErrOpCompleted struct {
	cause error
}

func NewErrOpCompleted(err error) ErrOpCompleted {
	return ErrOpCompleted{cause: err}
}

func (ErrOpCompleted) Error() string {
	return "the operation completed despite the context ending"
}

func (oc ErrOpCompleted) Cause() error {
	return oc.cause
}

type ErrOpCanceled struct {
	cause error
}

func NewErrOpCanceled(err error) ErrOpCanceled {
	return ErrOpCanceled{cause: err}
}

func (ErrOpCanceled) Error() string {
	return "the operation due to context ending"
}

func (oc ErrOpCanceled) Cause() error {
	return oc.cause
}
