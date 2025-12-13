package repository

import (
	"database/sql"
	"errors"

	"github.com/fitrinovs/achievement_system/app/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleRepository interface {
	FindAll() ([]model.Role, error)
	FindByID(id uuid.UUID) (*model.Role, error)
	FindByName(name string) (*model.Role, error)
	Create(role *model.Role) error
	Update(role *model.Role) error
	Delete(id uuid.UUID) error
	AssignPermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error
	GetRolePermissions(roleID uuid.UUID) ([]model.Permission, error)
}

type roleRepositoryGORM struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepositoryGORM{db: db}
}

func (r *roleRepositoryGORM) FindAll() ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Preload("Permissions").Find(&roles).Error
	return roles, err
}

func (r *roleRepositoryGORM) FindByID(id uuid.UUID) (*model.Role, error) {
	var role model.Role
	err := r.db.Preload("Permissions").Where("id = ?", id).First(&role).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepositoryGORM) FindByName(name string) (*model.Role, error) {
	var role model.Role
	err := r.db.Preload("Permissions").Where("name = ?", name).First(&role).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepositoryGORM) Create(role *model.Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepositoryGORM) Update(role *model.Role) error {
	return r.db.Save(role).Error
}

func (r *roleRepositoryGORM) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Role{}, id).Error
}

func (r *roleRepositoryGORM) AssignPermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	// First, remove existing permissions
	if err := r.db.Exec("DELETE FROM role_permissions WHERE role_id = ?", roleID).Error; err != nil {
		return err
	}

	// Then, insert new permissions
	for _, permID := range permissionIDs {
		rolePermission := model.RolePermission{
			RoleID:       roleID,
			PermissionID: permID,
		}
		if err := r.db.Create(&rolePermission).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *roleRepositoryGORM) GetRolePermissions(roleID uuid.UUID) ([]model.Permission, error) {
	var permissions []model.Permission

	err := r.db.Table("permissions").
		Select("permissions.*").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error

	return permissions, err
}

type RoleRepositorySQL struct {
	DB *sql.DB
}

func NewRoleRepositorySQL(db *sql.DB) RoleRepository {
	return &RoleRepositorySQL{DB: db}
}

func (r *RoleRepositorySQL) FindAll() ([]model.Role, error) {
	query := `SELECT id, name, description, created_at FROM roles ORDER BY name`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []model.Role
	for rows.Next() {
		var role model.Role
		err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, rows.Err()
}

func (r *RoleRepositorySQL) FindByID(id uuid.UUID) (*model.Role, error) {
	query := `SELECT id, name, description, created_at FROM roles WHERE id = $1`

	var role model.Role
	err := r.DB.QueryRow(query, id).Scan(
		&role.ID, &role.Name, &role.Description, &role.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("role not found")
		}
		return nil, err
	}

	return &role, nil
}

func (r *RoleRepositorySQL) FindByName(name string) (*model.Role, error) {
	query := `SELECT id, name, description, created_at FROM roles WHERE name = $1`

	var role model.Role
	err := r.DB.QueryRow(query, name).Scan(
		&role.ID, &role.Name, &role.Description, &role.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("role not found")
		}
		return nil, err
	}

	return &role, nil
}

func (r *RoleRepositorySQL) Create(role *model.Role) error {
	query := `
		INSERT INTO roles (id, name, description)
		VALUES ($1, $2, $3)
		RETURNING created_at
	`

	if role.ID == uuid.Nil {
		role.ID = uuid.New()
	}

	err := r.DB.QueryRow(query, role.ID, role.Name, role.Description).Scan(&role.CreatedAt)
	return err
}

func (r *RoleRepositorySQL) Update(role *model.Role) error {
	query := `
		UPDATE roles 
		SET name = $1, description = $2
		WHERE id = $3
	`

	result, err := r.DB.Exec(query, role.Name, role.Description, role.ID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("role not found")
	}

	return nil
}

func (r *RoleRepositorySQL) Delete(id uuid.UUID) error {
	query := `DELETE FROM roles WHERE id = $1`

	result, err := r.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("role not found")
	}

	return nil
}

func (r *RoleRepositorySQL) AssignPermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("DELETE FROM role_permissions WHERE role_id = $1", roleID)
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, permID := range permissionIDs {
		_, err = stmt.Exec(roleID, permID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *RoleRepositorySQL) GetRolePermissions(roleID uuid.UUID) ([]model.Permission, error) {
	query := `
		SELECT p.id, p.name, p.resource, p.action, p.description
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
	`

	rows, err := r.DB.Query(query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []model.Permission
	for rows.Next() {
		var perm model.Permission
		err := rows.Scan(&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Description)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}

	return permissions, rows.Err()
}
