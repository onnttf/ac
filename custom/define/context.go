package define

// Define a custom type for the context key
type contextKey string

// Implement the fmt.Stringer interface for contextKey
func (k contextKey) String() string {
	return string(k) // Return the string representation of the contextKey
}

const RequestIDKey contextKey = "request_id"

const (
	PrefixSystem   = "system"
	PrefixUser     = "user"
	PrefixRole     = "role"
	PrefixResource = "resource"
)
