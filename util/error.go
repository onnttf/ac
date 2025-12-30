package util

type Error struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Hint  string `json:"hint"`
	Cause error  `json:"-"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Msg
}

// NewError creates a new Error with code, message and hint.
func NewError(code int, msg, hint string) *Error {
	return &Error{
		Code: code,
		Msg:  msg,
		Hint: hint,
	}
}

// WithMsg returns a copy with a new message.
func (e *Error) WithMsg(newmsg string) *Error {
	if e == nil {
		return nil
	}
	return e.WithMsgAndHint(newmsg, e.Hint)
}

// WithHint returns a copy with a new hint.
func (e *Error) WithHint(newHint string) *Error {
	if e == nil {
		return nil
	}
	return e.WithMsgAndHint(e.Msg, newHint)
}

// WithMsgAndHint returns a copy with both message and hint updated.
func (e *Error) WithMsgAndHint(newmsg, newHint string) *Error {
	if e == nil {
		return nil
	}
	if e.Msg == newmsg && e.Hint == newHint {
		return e
	}
	return &Error{
		Code:  e.Code,
		Msg:   newmsg,
		Hint:  newHint,
		Cause: e.Cause,
	}
}

// WithError attaches an underlying cause error.
func (e *Error) WithError(cause error) *Error {
	if e == nil {
		return nil
	}
	if cause == nil {
		return e
	}
	return &Error{
		Code:  e.Code,
		Msg:   e.Msg,
		Hint:  e.Hint,
		Cause: cause,
	}
}
