package util

type Error struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Hint  string `json:"hint"`
	Cause error  `json:"-"`
}

func (e *Error) Error() string {
	return e.Msg
}

func NewError(code int, msg, hint string) *Error {
	return &Error{
		Code: code,
		Msg:  msg,
		Hint: hint,
	}
}

func (e *Error) Withmsg(newmsg string) *Error {
	if e == nil {
		return nil
	}
	return e.WithmsgAndHint(newmsg, e.Hint)
}

func (e *Error) WithHint(newHint string) *Error {
	if e == nil {
		return nil
	}
	return e.WithmsgAndHint(e.Msg, newHint)
}

func (e *Error) WithmsgAndHint(newmsg, newHint string) *Error {
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
