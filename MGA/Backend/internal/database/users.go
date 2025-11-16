package database

import (
    "context"
    "database/sql"
    "errors"
    "time"
)

type User struct {
    ID           int64
    Username     string
    Role         string
    PasswordHash sql.NullString
    PIN          sql.NullString
    HomeownerID  sql.NullInt64
    CreatedAt    time.Time
}

type UsersRepository struct {
    db *sql.DB
}

func NewUsersRepository(db *sql.DB) *UsersRepository {
    return &UsersRepository{db: db}
}

func (r *UsersRepository) Create(ctx context.Context, u *User) (int64, error) {
    if r.db == nil {
        return 0, errors.New("repo not initialized")
    }
    res, err := r.db.ExecContext(ctx,
        `INSERT INTO users (username, role, password_hash, pin, homeowner_id) VALUES (?,?,?,?,?)`,
        u.Username, u.Role, u.PasswordHash, u.PIN, u.HomeownerID,
    )
    if err != nil {
        return 0, err
    }
    return res.LastInsertId()
}

func (r *UsersRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
    row := r.db.QueryRowContext(ctx, `SELECT id, username, role, password_hash, pin, homeowner_id, created_at FROM users WHERE username = ?`, username)
    var u User
    if err := row.Scan(&u.ID, &u.Username, &u.Role, &u.PasswordHash, &u.PIN, &u.HomeownerID, &u.CreatedAt); err != nil {
        return nil, err
    }
    return &u, nil
}

func (r *UsersRepository) ListGuestsForHomeowner(ctx context.Context, homeownerID int64) ([]User, error) {
    rows, err := r.db.QueryContext(ctx, `SELECT id, username, role, password_hash, pin, homeowner_id, created_at FROM users WHERE role='guest' AND homeowner_id = ?`, homeownerID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var users []User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.PasswordHash, &u.PIN, &u.HomeownerID, &u.CreatedAt); err != nil {
            return nil, err
        }
        users = append(users, u)
    }
    return users, rows.Err()
}

func (r *UsersRepository) GetByID(ctx context.Context, id int64) (*User, error) {
    row := r.db.QueryRowContext(ctx, `SELECT id, username, role, password_hash, pin, homeowner_id, created_at FROM users WHERE id = ?`, id)
    var u User
    if err := row.Scan(&u.ID, &u.Username, &u.Role, &u.PasswordHash, &u.PIN, &u.HomeownerID, &u.CreatedAt); err != nil {
        return nil, err
    }
    return &u, nil
}

func (r *UsersRepository) ListTechnicians(ctx context.Context) ([]User, error) {
    rows, err := r.db.QueryContext(ctx, `SELECT id, username, role, password_hash, pin, homeowner_id, created_at FROM users WHERE role='technician'`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var users []User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.PasswordHash, &u.PIN, &u.HomeownerID, &u.CreatedAt); err != nil {
            return nil, err
        }
        users = append(users, u)
    }
    return users, rows.Err()
}

func (r *UsersRepository) Delete(ctx context.Context, id int64) error {
    _, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
    return err
}


