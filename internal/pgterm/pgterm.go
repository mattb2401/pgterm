package pgterm

type SessionContext struct {
	ActiveSchema   string
	ActiveDatabase string
}

var session = &SessionContext{
	ActiveSchema: "public",
}

func (s *SessionContext) SetSchema(schema string) {
	s.ActiveSchema = schema
}

func (s *SessionContext) GetSchema() string {
	return s.ActiveSchema
}

func (s *SessionContext) SetDatabase(database string) {
	s.ActiveDatabase = database
}

func (s *SessionContext) GetDatabase() string {
	return s.ActiveDatabase
}
