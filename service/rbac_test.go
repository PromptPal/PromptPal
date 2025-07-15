package service

import (
	"context"
	"testing"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/ent"
)

func setupTestDB(t *testing.T) *ent.Client {
	config.SetupConfig(true)
	InitDB()
	InitRedis(config.GetRuntimeConfig().RedisURL)
	return EntClient
}

func TestRBACService_InitializeRBACData(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	rbacService := NewRBACService(client)
	ctx := context.Background()

	// Test initialization
	err := rbacService.InitializeRBACData(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize RBAC data: %v", err)
	}

	// Verify roles were created
	roles, err := client.Role.Query().All(ctx)
	if err != nil {
		t.Fatalf("Failed to query roles: %v", err)
	}

	expectedRoles := []string{RoleSystemAdmin, RoleProjectAdmin, RoleProjectEdit, RoleProjectView}
	if len(roles) != len(expectedRoles) {
		t.Errorf("Expected %d roles, got %d", len(expectedRoles), len(roles))
	}

	for _, expectedRole := range expectedRoles {
		found := false
		for _, role := range roles {
			if role.Name == expectedRole {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Role %s not found", expectedRole)
		}
	}

	// Verify permissions were created
	permissions, err := client.Permission.Query().All(ctx)
	if err != nil {
		t.Fatalf("Failed to query permissions: %v", err)
	}

	expectedPermissions := []string{
		PermSystemAdmin, PermProjectAdmin, PermProjectEdit, PermProjectView,
		PermPromptCreate, PermPromptEdit, PermPromptDelete, PermPromptView,
		PermMetricsView, PermUserManage, PermProjectManage,
	}

	if len(permissions) < len(expectedPermissions) {
		t.Errorf("Expected at least %d permissions, got %d", len(expectedPermissions), len(permissions))
	}

	// Test idempotency - running again should not fail
	err = rbacService.InitializeRBACData(ctx)
	if err != nil {
		t.Fatalf("Second initialization should not fail: %v", err)
	}
}

func TestRBACService_AssignUserToProject(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	rbacService := NewRBACService(client)
	ctx := context.Background()

	// Initialize RBAC data
	err := rbacService.InitializeRBACData(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize RBAC data: %v", err)
	}

	// Create test user
	testUser, err := client.User.Create().
		SetName("test_service_rbac94").
		SetAddr("test_service_rbac94").
		SetEmail("test_service_rbac94@annatarhe.com").
		SetPhone("").
		SetLang("en").
		SetLevel(1).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test project
	testProject, err := client.Project.Create().
		SetName("Test Project106").
		SetCreatorID(testUser.ID).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Test role assignment
	err = rbacService.AssignUserToProject(ctx, testUser.ID, testProject.ID, RoleProjectAdmin)
	if err != nil {
		t.Fatalf("Failed to assign role: %v", err)
	}

	// Verify assignment
	userRoles, err := rbacService.GetUserProjectRoles(ctx, testUser.ID, testProject.ID)
	if err != nil {
		t.Fatalf("Failed to get user project roles: %v", err)
	}

	if len(userRoles) != 1 {
		t.Errorf("Expected 1 role, got %d", len(userRoles))
	}

	if userRoles[0].Name != RoleProjectAdmin {
		t.Errorf("Expected role %s, got %s", RoleProjectAdmin, userRoles[0].Name)
	}

	// Test idempotency - assigning same role again should not fail
	err = rbacService.AssignUserToProject(ctx, testUser.ID, testProject.ID, RoleProjectAdmin)
	if err != nil {
		t.Fatalf("Idempotent assignment should not fail: %v", err)
	}
}

func TestRBACService_HasPermission(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	rbacService := NewRBACService(client)
	ctx := context.Background()

	// Initialize RBAC data
	err := rbacService.InitializeRBACData(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize RBAC data: %v", err)
	}

	// Create test user with legacy admin level
	adminUser, err := client.User.Create().
		SetName("Admin User156").
		SetAddr("test_service_rbac_admin156@example.com").
		SetEmail("test_service_rbac_admin156@example.com").
		SetPhone("").
		SetLang("en").
		SetLevel(255).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	// Create regular user
	regularUser, err := client.User.Create().
		SetName("Regular User").
		SetAddr("test_service_rbac169").
		SetEmail("test_service_rbac169@annatarhe.com").
		SetPhone("").
		SetLang("en").
		SetLevel(1).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create regular user: %v", err)
	}

	// Create test project
	testProject, err := client.Project.Create().
		SetName("Test Project181").
		SetCreatorID(regularUser.ID).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Test legacy admin permissions (level > 100)
	hasPermission, err := rbacService.HasPermission(ctx, adminUser.ID, nil, PermSystemAdmin)
	if err != nil {
		t.Fatalf("Failed to check admin permission: %v", err)
	}
	if !hasPermission {
		t.Error("Admin user should have system admin permission due to legacy level")
	}

	// Test regular user without RBAC roles
	hasPermission, err = rbacService.HasPermission(ctx, regularUser.ID, nil, PermSystemAdmin)
	if err != nil {
		t.Fatalf("Failed to check regular user permission: %v", err)
	}
	if hasPermission {
		t.Error("Regular user should not have system admin permission")
	}

	// Assign project editor role to regular user
	err = rbacService.AssignUserToProject(ctx, regularUser.ID, testProject.ID, RoleProjectEdit)
	if err != nil {
		t.Fatalf("Failed to assign project editor role: %v", err)
	}

	// Test project-level permissions
	hasPermission, err = rbacService.HasPermission(ctx, regularUser.ID, &testProject.ID, PermProjectEdit)
	if err != nil {
		t.Fatalf("Failed to check project edit permission: %v", err)
	}
	if !hasPermission {
		t.Error("User with project editor role should have project edit permission")
	}

	// Test permission user doesn't have
	hasPermission, err = rbacService.HasPermission(ctx, regularUser.ID, &testProject.ID, PermProjectAdmin)
	if err != nil {
		t.Fatalf("Failed to check project admin permission: %v", err)
	}
	if hasPermission {
		t.Error("User with project editor role should not have project admin permission")
	}

	// EntClient.Project.DeleteOneID(testProject.ID).ExecX(ctx)
	// EntClient.User.DeleteOneID(adminUser.ID).ExecX(ctx)
	// EntClient.User.DeleteOneID(regularUser.ID).ExecX(ctx)
}

func TestRBACService_RemoveUserFromProject(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	rbacService := NewRBACService(client)
	ctx := context.Background()

	// Initialize RBAC data
	err := rbacService.InitializeRBACData(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize RBAC data: %v", err)
	}

	// Create test user and project
	testUser, err := client.User.Create().
		SetName("Test User").
		SetAddr("test_service_rbac250").
		SetEmail("test_service_rbac250@annatarhe.com").
		SetPhone("").
		SetLang("en").
		SetLevel(1).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	testProject, err := client.Project.Create().
		SetName("Test Project262").
		SetCreatorID(testUser.ID).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Assign role
	err = rbacService.AssignUserToProject(ctx, testUser.ID, testProject.ID, RoleProjectEdit)
	if err != nil {
		t.Fatalf("Failed to assign role: %v", err)
	}

	// Verify role was assigned
	userRoles, err := rbacService.GetUserProjectRoles(ctx, testUser.ID, testProject.ID)
	if err != nil {
		t.Fatalf("Failed to get user project roles: %v", err)
	}
	if len(userRoles) != 1 {
		t.Errorf("Expected 1 role before removal, got %d", len(userRoles))
	}

	// Remove role
	err = rbacService.RemoveUserFromProject(ctx, testUser.ID, testProject.ID, RoleProjectEdit)
	if err != nil {
		t.Fatalf("Failed to remove role: %v", err)
	}

	// Verify role was removed
	userRoles, err = rbacService.GetUserProjectRoles(ctx, testUser.ID, testProject.ID)
	if err != nil {
		t.Fatalf("Failed to get user project roles after removal: %v", err)
	}
	if len(userRoles) != 0 {
		t.Errorf("Expected 0 roles after removal, got %d", len(userRoles))
	}

	// EntClient.Project.DeleteOneID(testProject.ID).ExecX(ctx)
	// EntClient.User.DeleteOneID(testUser.ID).ExecX(ctx)
}

func TestRBACService_IsProjectOwner(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	rbacService := NewRBACService(client)
	ctx := context.Background()

	// Create test users
	owner, err := client.User.Create().
		SetName("Owner").
		SetAddr("test_service_owner313").
		SetEmail("test_service_owner313@annatarhe.com").
		SetPhone("").
		SetLang("en").
		SetLevel(1).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create owner user: %v", err)
	}

	otherUser, err := client.User.Create().
		SetName("Other User").
		SetAddr("test_service_other325@example.com").
		SetEmail("test_service_other325@example.com").
		SetPhone("").
		SetLang("en").
		SetLevel(1).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create other user: %v", err)
	}

	// Create test project
	testProject, err := client.Project.Create().
		SetName("Test Project337").
		SetCreatorID(owner.ID).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Test owner
	isOwner, err := rbacService.IsProjectOwner(ctx, owner.ID, testProject.ID)
	if err != nil {
		t.Fatalf("Failed to check project ownership: %v", err)
	}
	if !isOwner {
		t.Error("Owner should be recognized as project owner")
	}

	// Test non-owner
	isOwner, err = rbacService.IsProjectOwner(ctx, otherUser.ID, testProject.ID)
	if err != nil {
		t.Fatalf("Failed to check project ownership for non-owner: %v", err)
	}
	if isOwner {
		t.Error("Non-owner should not be recognized as project owner")
	}

	// EntClient.Project.DeleteOneID(testProject.ID).ExecX(ctx)
	// EntClient.User.DeleteOneID(owner.ID).ExecX(ctx)
	// EntClient.User.DeleteOneID(otherUser.ID).ExecX(ctx)
}

func TestRBACService_MigrateExistingUsers(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	rbacService := NewRBACService(client)
	ctx := context.Background()

	// Initialize RBAC data
	err := rbacService.InitializeRBACData(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize RBAC data: %v", err)
	}

	// Create test users with different levels
	adminUser, err := client.User.Create().
		SetName("Admin User").
		SetAddr("test_service_admin_383@example.com").
		SetEmail("test_service_admin_383@example.com").
		SetPhone("").
		SetLang("en").
		SetLevel(255).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	regularUser, err := client.User.Create().
		SetName("Regular User").
		SetAddr("test_service_regular_395@example.com").
		SetEmail("test_service_regular_395@example.com").
		SetPhone("").
		SetLang("en").
		SetLevel(1).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create regular user: %v", err)
	}

	// Create projects
	adminProject, err := client.Project.Create().
		SetName("Admin Project407").
		SetCreatorID(adminUser.ID).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create admin project: %v", err)
	}

	userProject, err := client.Project.Create().
		SetName("User Project415").
		SetCreatorID(regularUser.ID).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create user project: %v", err)
	}

	// Run migration
	err = rbacService.MigrateExistingUsers(ctx)
	if err != nil {
		t.Fatalf("Failed to migrate existing users: %v", err)
	}

	// Verify admin user roles
	adminRoles, err := rbacService.GetUserProjectRoles(ctx, adminUser.ID, adminProject.ID)
	if err != nil {
		t.Fatalf("Failed to get admin user roles: %v", err)
	}
	if len(adminRoles) == 0 {
		t.Error("Admin user should have roles after migration")
	}

	// Verify regular user roles
	userRoles, err := rbacService.GetUserProjectRoles(ctx, regularUser.ID, userProject.ID)
	if err != nil {
		t.Fatalf("Failed to get regular user roles: %v", err)
	}
	if len(userRoles) == 0 {
		t.Error("Regular user should have roles after migration")
	}

	// Test idempotency - running migration again should not fail or duplicate roles
	err = rbacService.MigrateExistingUsers(ctx)
	if err != nil {
		t.Fatalf("Second migration should not fail: %v", err)
	}

	// Clean up
	// EntClient.Project.DeleteOneID(adminProject.ID).ExecX(ctx)
	// EntClient.Project.DeleteOneID(userProject.ID).ExecX(ctx)
	// EntClient.User.DeleteOneID(adminUser.ID).ExecX(ctx)
	// EntClient.User.DeleteOneID(regularUser.ID).ExecX(ctx)
}
