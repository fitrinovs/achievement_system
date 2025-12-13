package repository

import (
	"database/sql"
	"errors"

	"github.com/fitrinovs/achievement_system/app/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByUsername(username string) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	FindByID(id uuid.UUID) (*model.User, error)
	Create(user *model.User) error
	Update(user *model.User) error
	Delete(id uuid.UUID) error
	GetUserPermissions(userID uuid.UUID) ([]string, error)
}

// GORM Implementation
type userRepositoryGORM struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryGORM{db: db}
}

// SQL Implementation
type UserRepositorySQL struct {
	DB *sql.DB
}

func NewUserRepositorySQL(db *sql.DB) UserRepository {
	return &UserRepositorySQL{DB: db}
}

// ============ GORM IMPLEMENTATION ============

func (r *userRepositoryGORM) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Role.Permissions").
		Where("username = ? AND is_active = ?", username, true).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryGORM) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Role.Permissions").
		Where("email = ? AND is_active = ?", email, true).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryGORM) FindByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Role.Permissions").
		Where("id = ? AND is_active = ?", id, true).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryGORM) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *userRepositoryGORM) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *userRepositoryGORM) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.User{}, id).Error
}

func (r *userRepositoryGORM) GetUserPermissions(userID uuid.UUID) ([]string, error) {
	var permissions []string

	err := r.db.Table("users").
		Select("permissions.name").
		Joins("JOIN roles ON users.role_id = roles.id").
		Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("users.id = ? AND users.is_active = ?", userID, true).
		Pluck("permissions.name", &permissions).Error

	if err != nil {
		return nil, err
	}

	return permissions, nil
}

// ============ SQL IMPLEMENTATION ============

func (r *UserRepositorySQL) FindByUsername(username string) (*model.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.full_name, 
		       u.role_id, u.is_active, u.created_at, u.updated_at,
		       r.id, r.name, r.description
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.username = $1 AND u.is_active = true
	`

	var user model.User
	var role model.Role

	err := r.DB.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		&role.ID, &role.Name, &role.Description,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	user.Role = &role
	return &user, nil
}

func (r *UserRepositorySQL) FindByEmail(email string) (*model.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.full_name, 
		       u.role_id, u.is_active, u.created_at, u.updated_at,
		       r.id, r.name, r.description
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.email = $1 AND u.is_active = true
	`

	var user model.User
	var role model.Role

	err := r.DB.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		&role.ID, &role.Name, &role.Description,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	user.Role = &role
	return &user, nil
}

func (r *UserRepositorySQL) FindByID(id uuid.UUID) (*model.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.full_name, 
		       u.role_id, u.is_active, u.created_at, u.updated_at,
		       r.id, r.name, r.description
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1 AND u.is_active = true
	`

	var user model.User
	var role model.Role

	err := r.DB.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		&role.ID, &role.Name, &role.Description,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	user.Role = &role
	return &user, nil
}

func (r *UserRepositorySQL) Create(user *model.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, full_name, role_id, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	err := r.DB.QueryRow(
		query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.FullName, user.RoleID, user.IsActive,
	).Scan(&user.CreatedAt, &user.UpdatedAt)

	return err
}

func (r *UserRepositorySQL) Update(user *model.User) error {
	query := `
		UPDATE users 
		SET username = $1, email = $2, password_hash = $3, 
		    full_name = $4, role_id = $5, is_active = $6, updated_at = NOW()
		WHERE id = $7
		RETURNING updated_at
	`

	err := r.DB.QueryRow(
		query,
		user.Username, user.Email, user.PasswordHash,
		user.FullName, user.RoleID, user.IsActive, user.ID,
	).Scan(&user.UpdatedAt)

	return err
}

func (r *UserRepositorySQL) Delete(id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *UserRepositorySQL) GetUserPermissions(userID uuid.UUID) ([]string, error) {
	query := `
		SELECT p.name
		FROM users u
		JOIN roles r ON u.role_id = r.id
		JOIN role_permissions rp ON r.id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE u.id = $1 AND u.is_active = true
	`

	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}
