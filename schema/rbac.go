package schema

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/userprojectrole"
	"github.com/PromptPal/PromptPal/service"
)

// Role resolver

type roleResponse struct {
	r *ent.Role
}

func (r roleResponse) ID() string {
	return strconv.Itoa(r.r.ID)
}

func (r roleResponse) Name() string {
	return r.r.Name
}

func (r roleResponse) Description() *string {
	if r.r.Description == "" {
		return nil
	}
	return &r.r.Description
}

func (r roleResponse) IsSystemRole() bool {
	return r.r.IsSystemRole
}

func (r roleResponse) Permissions(ctx context.Context) ([]permissionResponse, error) {
	perms, err := r.r.QueryPermissions().All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]permissionResponse, len(perms))
	for i, p := range perms {
		result[i] = permissionResponse{p: p}
	}
	return result, nil
}

func (r roleResponse) CreatedAt() string {
	return r.r.CreateTime.Format("2006-01-02T15:04:05Z")
}

func (r roleResponse) UpdatedAt() string {
	return r.r.UpdateTime.Format("2006-01-02T15:04:05Z")
}

// Permission resolver

type permissionResponse struct {
	p *ent.Permission
}

func (p permissionResponse) ID() string {
	return strconv.Itoa(p.p.ID)
}

func (p permissionResponse) Name() string {
	return p.p.Name
}

func (p permissionResponse) Description() *string {
	if p.p.Description == "" {
		return nil
	}
	return &p.p.Description
}

func (p permissionResponse) Resource() *string {
	if p.p.Resource == "" {
		return nil
	}
	return &p.p.Resource
}

func (p permissionResponse) Action() *string {
	if p.p.Action == "" {
		return nil
	}
	return &p.p.Action
}

func (p permissionResponse) CreatedAt() string {
	return p.p.CreateTime.Format("2006-01-02T15:04:05Z")
}

func (p permissionResponse) UpdatedAt() string {
	return p.p.UpdateTime.Format("2006-01-02T15:04:05Z")
}

// UserProjectRole resolver

type userProjectRoleResponse struct {
	upr *ent.UserProjectRole
}

func (upr userProjectRoleResponse) ID() string {
	return strconv.Itoa(upr.upr.ID)
}

func (upr userProjectRoleResponse) User(ctx context.Context) (userResponse, error) {
	u, err := upr.upr.QueryUser().Only(ctx)
	if err != nil {
		return userResponse{}, err
	}
	return userResponse{u: u}, nil
}

func (upr userProjectRoleResponse) Project(ctx context.Context) (projectResponse, error) {
	p, err := upr.upr.QueryProject().Only(ctx)
	if err != nil {
		return projectResponse{}, err
	}
	return projectResponse{p: p}, nil
}

func (upr userProjectRoleResponse) Role(ctx context.Context) (roleResponse, error) {
	r, err := upr.upr.QueryRole().Only(ctx)
	if err != nil {
		return roleResponse{}, err
	}
	return roleResponse{r: r}, nil
}

func (upr userProjectRoleResponse) CreatedAt() string {
	return upr.upr.CreateTime.Format("2006-01-02T15:04:05Z")
}

func (upr userProjectRoleResponse) UpdatedAt() string {
	return upr.upr.UpdateTime.Format("2006-01-02T15:04:05Z")
}

// List response types

type roleListResponse struct {
	count int
	edges []roleResponse
}

func (r roleListResponse) Count() int {
	return r.count
}

func (r roleListResponse) Edges() []roleResponse {
	return r.edges
}

type permissionListResponse struct {
	count int
	edges []permissionResponse
}

func (p permissionListResponse) Count() int {
	return p.count
}

func (p permissionListResponse) Edges() []permissionResponse {
	return p.edges
}

// Query resolvers

func (q QueryResolver) Roles(ctx context.Context) (roleListResponse, error) {
	// Check if user has permission to view roles
	ctxValue, ok := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	if !ok {
		return roleListResponse{}, NewGraphQLHttpError(http.StatusUnauthorized, errors.New("authentication required"))
	}

	hasPermission, err := rbacService.HasPermission(ctx, ctxValue.UserID, nil, service.PermSystemAdmin)
	if err != nil {
		return roleListResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	if !hasPermission {
		return roleListResponse{}, NewGraphQLHttpError(http.StatusForbidden, errors.New("admin privileges required"))
	}

	roles, err := service.EntClient.Role.Query().All(ctx)
	if err != nil {
		return roleListResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	edges := make([]roleResponse, len(roles))
	for i, r := range roles {
		edges[i] = roleResponse{r: r}
	}

	return roleListResponse{
		count: len(roles),
		edges: edges,
	}, nil
}

func (q QueryResolver) Permissions(ctx context.Context) (permissionListResponse, error) {
	// Check if user has permission to view permissions
	ctxValue, ok := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	if !ok {
		return permissionListResponse{}, NewGraphQLHttpError(http.StatusUnauthorized, errors.New("authentication required"))
	}

	hasPermission, err := rbacService.HasPermission(ctx, ctxValue.UserID, nil, service.PermSystemAdmin)
	if err != nil {
		return permissionListResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	if !hasPermission {
		return permissionListResponse{}, NewGraphQLHttpError(http.StatusForbidden, errors.New("admin privileges required"))
	}

	permissions, err := service.EntClient.Permission.Query().All(ctx)
	if err != nil {
		return permissionListResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	edges := make([]permissionResponse, len(permissions))
	for i, p := range permissions {
		edges[i] = permissionResponse{p: p}
	}

	return permissionListResponse{
		count: len(permissions),
		edges: edges,
	}, nil
}

type userProjectRolesArgs struct {
	UserID    *string
	ProjectID *string
}

func (q QueryResolver) UserProjectRoles(ctx context.Context, args userProjectRolesArgs) ([]userProjectRoleResponse, error) {
	// Check if user has permission to view user project roles
	ctxValue, ok := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	if !ok {
		return nil, NewGraphQLHttpError(http.StatusUnauthorized, errors.New("authentication required"))
	}

	query := service.EntClient.UserProjectRole.Query()

	// Apply filters
	if args.UserID != nil {
		userID, err := strconv.Atoi(*args.UserID)
		if err != nil {
			return nil, NewGraphQLHttpError(http.StatusBadRequest, errors.New("invalid user ID"))
		}
		query = query.Where(userprojectrole.UserID(userID))
	}

	if args.ProjectID != nil {
		projectID, err := strconv.Atoi(*args.ProjectID)
		if err != nil {
			return nil, NewGraphQLHttpError(http.StatusBadRequest, errors.New("invalid project ID"))
		}
		query = query.Where(userprojectrole.ProjectID(projectID))

		// Check if user has permission to view this project's roles
		hasPermission, err := rbacService.HasPermission(ctx, ctxValue.UserID, &projectID, service.PermProjectView)
		if err != nil {
			return nil, NewGraphQLHttpError(http.StatusInternalServerError, err)
		}

		if !hasPermission {
			return nil, NewGraphQLHttpError(http.StatusForbidden, errors.New("insufficient permissions"))
		}
	} else {
		// Without project filter, require system admin
		hasPermission, err := rbacService.HasPermission(ctx, ctxValue.UserID, nil, service.PermSystemAdmin)
		if err != nil {
			return nil, NewGraphQLHttpError(http.StatusInternalServerError, err)
		}

		if !hasPermission {
			return nil, NewGraphQLHttpError(http.StatusForbidden, errors.New("admin privileges required"))
		}
	}

	uprs, err := query.All(ctx)
	if err != nil {
		return nil, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	result := make([]userProjectRoleResponse, len(uprs))
	for i, upr := range uprs {
		result[i] = userProjectRoleResponse{upr: upr}
	}
	return result, nil
}

// Mutation resolvers

type assignRoleInput struct {
	UserID    string
	ProjectID string
	RoleName  string
}

type assignRoleArgs struct {
	Input assignRoleInput
}

type assignRoleResponse struct {
	success bool
	message *string
}

func (r assignRoleResponse) Success() bool {
	return r.success
}

func (r assignRoleResponse) Message() *string {
	return r.message
}

func (q QueryResolver) AssignUserToProject(ctx context.Context, args assignRoleArgs) (assignRoleResponse, error) {
	// Check if user has permission to assign roles
	ctxValue, ok := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	if !ok {
		msg := "authentication required"
		return assignRoleResponse{success: false, message: &msg}, nil
	}

	projectID, err := strconv.Atoi(args.Input.ProjectID)
	if err != nil {
		msg := "invalid project ID"
		return assignRoleResponse{success: false, message: &msg}, nil
	}

	hasPermission, err := rbacService.HasPermission(ctx, ctxValue.UserID, &projectID, service.PermProjectAdmin)
	if err != nil {
		msg := "permission check failed"
		return assignRoleResponse{success: false, message: &msg}, nil
	}

	if !hasPermission {
		msg := "insufficient permissions"
		return assignRoleResponse{success: false, message: &msg}, nil
	}

	userID, err := strconv.Atoi(args.Input.UserID)
	if err != nil {
		msg := "invalid user ID"
		return assignRoleResponse{success: false, message: &msg}, nil
	}

	err = rbacService.AssignUserToProject(ctx, userID, projectID, args.Input.RoleName)
	if err != nil {
		msg := err.Error()
		return assignRoleResponse{success: false, message: &msg}, nil
	}

	return assignRoleResponse{success: true, message: nil}, nil
}

type removeRoleInput struct {
	UserID    string
	ProjectID string
	RoleName  string
}

type removeRoleArgs struct {
	Input removeRoleInput
}

type removeRoleResponse struct {
	success bool
	message *string
}

func (r removeRoleResponse) Success() bool {
	return r.success
}

func (r removeRoleResponse) Message() *string {
	return r.message
}

func (q QueryResolver) RemoveUserFromProject(ctx context.Context, args removeRoleArgs) (removeRoleResponse, error) {
	// Check if user has permission to remove roles
	ctxValue, ok := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	if !ok {
		msg := "authentication required"
		return removeRoleResponse{success: false, message: &msg}, nil
	}

	projectID, err := strconv.Atoi(args.Input.ProjectID)
	if err != nil {
		msg := "invalid project ID"
		return removeRoleResponse{success: false, message: &msg}, nil
	}

	hasPermission, err := rbacService.HasPermission(ctx, ctxValue.UserID, &projectID, service.PermProjectAdmin)
	if err != nil {
		msg := "permission check failed"
		return removeRoleResponse{success: false, message: &msg}, nil
	}

	if !hasPermission {
		msg := "insufficient permissions"
		return removeRoleResponse{success: false, message: &msg}, nil
	}

	userID, err := strconv.Atoi(args.Input.UserID)
	if err != nil {
		msg := "invalid user ID"
		return removeRoleResponse{success: false, message: &msg}, nil
	}

	err = rbacService.RemoveUserFromProject(ctx, userID, projectID, args.Input.RoleName)
	if err != nil {
		msg := err.Error()
		return removeRoleResponse{success: false, message: &msg}, nil
	}

	return removeRoleResponse{success: true, message: nil}, nil
}
