package service

import (
	"context"
	"fmt"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/permission"
	"github.com/PromptPal/PromptPal/ent/role"
	"github.com/PromptPal/PromptPal/ent/user"
	"github.com/PromptPal/PromptPal/ent/userprojectrole"
)

// RBAC Role and Permission constants
const (
	// System Roles
	RoleSystemAdmin  = "SYSTEM_ADMIN"
	RoleProjectAdmin = "PROJECT_ADMIN"
	RoleProjectEdit  = "PROJECT_EDITOR"
	RoleProjectView  = "PROJECT_VIEWER"

	// Permissions
	PermSystemAdmin   = "system:admin"
	PermProjectAdmin  = "project:admin"
	PermProjectEdit   = "project:edit"
	PermProjectView   = "project:view"
	PermPromptCreate  = "prompt:create"
	PermPromptEdit    = "prompt:edit"
	PermPromptDelete  = "prompt:delete"
	PermPromptView    = "prompt:view"
	PermMetricsView   = "metrics:view"
	PermUserManage    = "user:manage"
	PermProjectManage = "project:manage"
)

// RBACService interface defines the RBAC operations
type RBACService interface {
	InitializeRBACData(ctx context.Context) error
	MigrateExistingUsers(ctx context.Context) error
	GetUserProjectRoles(ctx context.Context, userID, projectID int) ([]*ent.Role, error)
	IsProjectOwner(ctx context.Context, userID, projectID int) (bool, error)

	HasPermission(ctx context.Context, userID int, projectID *int, permissionName string) (bool, error)
	AssignUserToProject(ctx context.Context, userID, projectID int, roleName string) error
	RemoveUserFromProject(ctx context.Context, userID, projectID int, roleName string) error
}

type rbacService struct {
	client *ent.Client
}

func NewRBACService(client *ent.Client) RBACService {
	return &rbacService{
		client: client,
	}
}

// InitializeRBACData creates the default roles and permissions
func (s *rbacService) InitializeRBACData(ctx context.Context) error {
	// Create permissions first
	permissions := []struct {
		name        string
		description string
		resource    string
		action      string
	}{
		{PermSystemAdmin, "Full system administration", "system", "admin"},
		{PermProjectAdmin, "Project administration", "project", "admin"},
		{PermProjectEdit, "Edit project content", "project", "edit"},
		{PermProjectView, "View project content", "project", "view"},
		{PermPromptCreate, "Create prompts", "prompt", "create"},
		{PermPromptEdit, "Edit prompts", "prompt", "edit"},
		{PermPromptDelete, "Delete prompts", "prompt", "delete"},
		{PermPromptView, "View prompts", "prompt", "view"},
		{PermMetricsView, "View metrics", "metrics", "view"},
		{PermUserManage, "Manage users", "user", "manage"},
		{PermProjectManage, "Manage projects", "project", "manage"},
	}

	for _, perm := range permissions {
		// Check if permission already exists
		exists, err := s.client.Permission.Query().
			Where(permission.Name(perm.name)).
			Exist(ctx)
		if err != nil {
			return fmt.Errorf("failed to check permission %s: %w", perm.name, err)
		}

		if !exists {
			_, err := s.client.Permission.Create().
				SetName(perm.name).
				SetDescription(perm.description).
				SetResource(perm.resource).
				SetAction(perm.action).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create permission %s: %w", perm.name, err)
			}
		}
	}

	// Create roles and assign permissions
	rolePermissions := map[string][]string{
		RoleSystemAdmin: {
			PermSystemAdmin, PermProjectAdmin, PermProjectEdit, PermProjectView,
			PermPromptCreate, PermPromptEdit, PermPromptDelete, PermPromptView,
			PermMetricsView, PermUserManage, PermProjectManage,
		},
		RoleProjectAdmin: {
			PermProjectAdmin, PermProjectEdit, PermProjectView,
			PermPromptCreate, PermPromptEdit, PermPromptDelete, PermPromptView,
			PermMetricsView, PermUserManage,
		},
		RoleProjectEdit: {
			PermProjectEdit, PermProjectView,
			PermPromptCreate, PermPromptEdit, PermPromptView,
			PermMetricsView,
		},
		RoleProjectView: {
			PermProjectView, PermPromptView, PermMetricsView,
		},
	}

	for roleName, permNames := range rolePermissions {
		// Check if role already exists
		roleEntity, err := s.client.Role.Query().
			Where(role.Name(roleName)).
			Only(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				// This is an unexpected error
				return fmt.Errorf("failed to query role %s: %w", roleName, err)
			}
			// Role doesn't exist, create it
			roleEntity, err = s.client.Role.Create().
				SetName(roleName).
				SetIsSystemRole(true).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create role %s: %w", roleName, err)
			}
		}

		// Get permissions
		perms, err := s.client.Permission.Query().
			Where(permission.NameIn(permNames...)).
			All(ctx)
		if err != nil {
			return fmt.Errorf("failed to query permissions for role %s: %w", roleName, err)
		}

		// Assign permissions to role
		_, err = roleEntity.Update().
			AddPermissions(perms...).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to assign permissions to role %s: %w", roleName, err)
		}
	}

	return nil
}

// HasPermission checks if a user has a specific permission in a project
func (s *rbacService) HasPermission(ctx context.Context, userID int, projectID *int, permissionName string) (bool, error) {
	// Check if user is system admin (legacy level check and new RBAC)
	u, err := s.client.User.Query().
		Where(user.ID(userID)).
		Only(ctx)
	if err != nil {
		return false, err
	}

	// Legacy admin check
	if u.Level > 100 {
		return true, nil
	}

	// System-level permissions (no project required)
	if projectID == nil {
		return s.hasSystemPermission(ctx, userID, permissionName)
	}

	// Project-level permissions
	return s.hasProjectPermission(ctx, userID, *projectID, permissionName)
}

// hasSystemPermission checks system-level permissions
func (s *rbacService) hasSystemPermission(ctx context.Context, userID int, permissionName string) (bool, error) {
	// Check if user has system admin role
	count, err := s.client.UserProjectRole.Query().
		Where(
			userprojectrole.UserID(userID),
			userprojectrole.HasRoleWith(role.Name(RoleSystemAdmin)),
		).
		Count(ctx)
	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	// Check specific permission
	count, err = s.client.UserProjectRole.Query().
		Where(
			userprojectrole.UserID(userID),
			userprojectrole.HasRoleWith(
				role.HasPermissionsWith(permission.Name(permissionName)),
			),
		).
		Count(ctx)
	return count > 0, err
}

// hasProjectPermission checks project-level permissions
func (s *rbacService) hasProjectPermission(ctx context.Context, userID, projectID int, permissionName string) (bool, error) {
	count, err := s.client.UserProjectRole.Query().
		Where(
			userprojectrole.UserID(userID),
			userprojectrole.ProjectID(projectID),
			userprojectrole.HasRoleWith(
				role.HasPermissionsWith(permission.Name(permissionName)),
			),
		).
		Count(ctx)
	return count > 0, err
}

// AssignUserToProject assigns a role to a user for a specific project
func (s *rbacService) AssignUserToProject(ctx context.Context, userID, projectID int, roleName string) error {
	// Get role
	roleEntity, err := s.client.Role.Query().
		Where(role.Name(roleName)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("role %s not found", roleName)
		}
		return fmt.Errorf("failed to query role %s: %w", roleName, err)
	}

	// Check if assignment already exists
	exists, err := s.client.UserProjectRole.Query().
		Where(
			userprojectrole.UserID(userID),
			userprojectrole.ProjectID(projectID),
			userprojectrole.RoleID(roleEntity.ID),
		).
		Exist(ctx)
	if err != nil {
		return fmt.Errorf("failed to check existing role assignment: %w", err)
	}

	if !exists {
		// Create user-project-role assignment
		_, err = s.client.UserProjectRole.Create().
			SetUserID(userID).
			SetProjectID(projectID).
			SetRoleID(roleEntity.ID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to assign role %s to user %d in project %d: %w", roleName, userID, projectID, err)
		}
	}

	return nil
}

// RemoveUserFromProject removes a user's role from a project
func (s *rbacService) RemoveUserFromProject(ctx context.Context, userID, projectID int, roleName string) error {
	// Get role
	roleEntity, err := s.client.Role.Query().
		Where(role.Name(roleName)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("role %s not found", roleName)
		}
		return fmt.Errorf("failed to query role %s: %w", roleName, err)
	}

	// Remove assignment
	_, err = s.client.UserProjectRole.Delete().
		Where(
			userprojectrole.UserID(userID),
			userprojectrole.ProjectID(projectID),
			userprojectrole.RoleID(roleEntity.ID),
		).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to remove role %s from user %d in project %d: %w", roleName, userID, projectID, err)
	}

	return nil
}

// GetUserProjectRoles returns all roles a user has in a project
func (s *rbacService) GetUserProjectRoles(ctx context.Context, userID, projectID int) ([]*ent.Role, error) {
	roles, err := s.client.Role.Query().
		Where(
			role.HasUserProjectRolesWith(
				userprojectrole.UserID(userID),
				userprojectrole.ProjectID(projectID),
			),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user project roles: %w", err)
	}

	return roles, nil
}

// IsProjectOwner checks if a user is the original creator/owner of a project
func (s *rbacService) IsProjectOwner(ctx context.Context, userID, projectID int) (bool, error) {
	// Query project by ID and check if the creator_id matches the userID
	project, err := s.client.Project.Get(ctx, projectID)
	if err != nil {
		return false, err
	}

	// Check if the project's creator_id matches the userID (0 means no creator set)
	return project.CreatorID != 0 && project.CreatorID == userID, nil
}

// MigrateExistingUsers migrates users from the old level system to RBAC
func (s *rbacService) MigrateExistingUsers(ctx context.Context) error {
	users, err := s.client.User.Query().All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query users for migration: %w", err)
	}

	for _, u := range users {
		// Skip if user already has roles assigned
		count, err := s.client.UserProjectRole.Query().
			Where(userprojectrole.UserID(u.ID)).
			Count(ctx)
		if err != nil {
			return fmt.Errorf("failed to check existing roles for user %d: %w", u.ID, err)
		}
		if count > 0 {
			continue // User already migrated
		}

		// Get user's projects (projects where user is the creator)
		projects, err := u.QueryProjects().All(ctx)
		if err != nil {
			return fmt.Errorf("failed to query projects for user %d: %w", u.ID, err)
		}

		// Assign appropriate roles based on level
		for _, project := range projects {
			var roleName string
			switch {
			case u.Level >= 255:
				roleName = RoleSystemAdmin
			case u.Level > 100:
				roleName = RoleProjectAdmin
			case u.Level >= 1:
				roleName = RoleProjectAdmin // Project owner
			default:
				roleName = RoleProjectView
			}

			err = s.AssignUserToProject(ctx, u.ID, project.ID, roleName)
			if err != nil {
				return fmt.Errorf("failed to assign role %s to user %d in project %d: %w", roleName, u.ID, project.ID, err)
			}
		}
	}

	return nil
}
