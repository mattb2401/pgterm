package pgterm

// SessionContext holds the current session state, including the
// active schema and active database selected by the user.
type SessionContext struct {
	ActiveSchema   string // The currently selected schema (e.g., "public")
	ActiveDatabase string // The currently selected database
}

// session is the global session context initialized with the default schema "public".
var session = &SessionContext{
	ActiveSchema: "public",
}

// SetSchema sets the active schema for the session.
func (s *SessionContext) SetSchema(schema string) {
	s.ActiveSchema = schema
}

// GetSchema returns the currently active schema.
func (s *SessionContext) GetSchema() string {
	return s.ActiveSchema
}

// SetDatabase sets the active database for the session.
func (s *SessionContext) SetDatabase(database string) {
	s.ActiveDatabase = database
}

// GetDatabase returns the currently active database.
func (s *SessionContext) GetDatabase() string {
	return s.ActiveDatabase
}
