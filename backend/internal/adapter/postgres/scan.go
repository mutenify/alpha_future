package postgres

type scanner interface{ Scan(dest ...any) error }

// collect читает все строки через scan.
func collect[T any](rows interface {
	scanner
	Next() bool
	Err() error
	Close()
}, scan func(scanner) (T, error)) ([]T, error) {
	defer rows.Close()
	var out []T
	for rows.Next() {
		v, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
