package postgres

// nullable — пустая строка уходит в БД как NULL (uuid и необязательные поля).
func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
