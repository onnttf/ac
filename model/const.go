package model

type StatusFlag int64

const (
	StatusDisabled StatusFlag = 0
	StatusEnabled  StatusFlag = 1
)

func (s StatusFlag) String() string {
	switch s {
	case StatusDisabled:
		return "disabled"
	case StatusEnabled:
		return "enabled"
	default:
		return "unknown"
	}
}

func (s StatusFlag) Int64() int64 {
	return int64(s)
}

type DeletedFlag int64

const (
	NotDeleted DeletedFlag = 0
	Deleted    DeletedFlag = 1
)

func (d DeletedFlag) String() string {
	switch d {
	case NotDeleted:
		return "notDeleted"
	case Deleted:
		return "deleted"
	default:
		return "unknown"
	}
}

func (d DeletedFlag) Int64() int64 {
	return int64(d)
}
