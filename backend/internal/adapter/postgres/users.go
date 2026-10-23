package postgres

import (
	"context"

	"smartcredit/internal/domain"
)

type Users struct{ *DB }

const userCols = `id, login, password_hash, role, full_name, coalesce(position, ''), coalesce(scenario_hint, '')`

func (r Users) ByLogin(ctx context.Context, login string) (domain.User, error) {
	return r.one(ctx, `SELECT `+userCols+` FROM users WHERE login = $1`, login)
}

func (r Users) ByID(ctx context.Context, id string) (domain.User, error) {
	return r.one(ctx, `SELECT `+userCols+` FROM users WHERE id = $1`, id)
}

func (r Users) DemoUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := r.q(ctx).Query(ctx, `SELECT `+userCols+` FROM users WHERE scenario_hint IS NOT NULL ORDER BY role, login`)
	if err != nil {
		return nil, err
	}
	return collect(rows, scanUser)
}

func (r Users) one(ctx context.Context, sql string, arg any) (domain.User, error) {
	u, err := scanUser(r.q(ctx).QueryRow(ctx, sql, arg))
	return u, notFound(err)
}

func scanUser(row scanner) (domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Login, &u.PasswordHash, &u.Role, &u.FullName, &u.Position, &u.ScenarioHint)
	return u, err
}
