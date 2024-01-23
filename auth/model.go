package auth

type Registration struct {
	Email    string `json:"email" validate:"nonzero"`
	Name     string `json:"name" validate:"nonzero"`
	Password string `json:"password" validate:"nonzero"`
}

type Login struct {
	Email    string `json:"email" validate:"nonzero"`
	Password string `json:"password" validate:"nonzero"`
}

type User struct {
	ID       uint64 `db:"id" json:"id"`
	Name     string `db:"name" json:"name"`
	Email    string `db:"email" json:"email"`
	Password string `db:"password"`
	Verified bool   `db:"verified" json:"verified"`
}
