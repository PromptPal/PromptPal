package schema

import (
	"context"
	"net/http"
	"testing"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/enttest"
	"github.com/PromptPal/PromptPal/service"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"
)

func TestWebhookCRUD(t *testing.T) {
	// Create test database
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	// Setup services
	service.EntClient = client
	rbacService = service.NewRBACService(client)
	
	// Initialize RBAC data
	err := rbacService.InitializeRBACData(context.Background())
	require.NoError(t, err)

	// Create test user
	user, err := client.User.Create().
		SetName("Test User").
		SetAddr("test@example.com").
		SetEmail("test@example.com").
		SetPhone("").
		SetLang("en").
		SetLevel(255). // Admin level
		Save(context.Background())
	require.NoError(t, err)
	require.NotNil(t, user)

	// Create test project
	project, err := client.Project.Create().
		SetName("Test Project").
		SetCreatorID(user.ID).
		SetProviderId(1).
		Save(context.Background())
	require.NoError(t, err)
	require.NotNil(t, project)

	// Create context with user
	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: user.ID,
	})

	t.Run("CreateWebhook", func(t *testing.T) {
		args := createWebhookArgs{
			Data: createWebhookData{
				Name:        "Test Webhook",
				Description: stringPtr("Test webhook description"),
				URL:         "https://example.com/webhook",
				Event:       "onPromptFinished",
				Enabled:     boolPtr(true),
				ProjectID:   int32(project.ID),
			},
		}

		resolver := QueryResolver{}
		resp, err := resolver.CreateWebhook(ctx, args)
		require.NoError(t, err)
		require.Equal(t, "Test Webhook", resp.Name())
		require.Equal(t, "Test webhook description", resp.Description())
		require.Equal(t, "https://example.com/webhook", resp.URL())
		require.Equal(t, "onPromptFinished", resp.Event())
		require.True(t, resp.Enabled())
	})

	t.Run("CreateWebhookInvalidEvent", func(t *testing.T) {
		args := createWebhookArgs{
			Data: createWebhookData{
				Name:      "Test Webhook",
				URL:       "https://example.com/webhook",
				Event:     "invalidEvent",
				ProjectID: int32(project.ID),
			},
		}

		resolver := QueryResolver{}
		_, err := resolver.CreateWebhook(ctx, args)
		require.Error(t, err)
		require.Contains(t, err.Error(), "only onPromptFinished event is supported")
	})

	t.Run("ListWebhooks", func(t *testing.T) {
		// Create a webhook first
		webhook, err := client.Webhook.Create().
			SetName("List Test Webhook").
			SetURL("https://example.com/webhook").
			SetEvent("onPromptFinished").
			SetCreatorID(user.ID).
			SetProjectID(project.ID).
			Save(context.Background())
		require.NoError(t, err)
		require.NotNil(t, webhook)

		args := webhooksArgs{
			ProjectID: int32(project.ID),
			Pagination: paginationInput{
				Limit:  10,
				Offset: 0,
			},
		}

		resolver := QueryResolver{}
		resp := resolver.Webhooks(ctx, args)
		
		count, err := resp.Count(ctx)
		require.NoError(t, err)
		require.Greater(t, count, int32(0))

		edges, err := resp.Edges(ctx)
		require.NoError(t, err)
		require.Greater(t, len(edges), 0)

		found := false
		for _, edge := range edges {
			if edge.Name() == "List Test Webhook" {
				found = true
				break
			}
		}
		require.True(t, found)
	})

	t.Run("UpdateWebhook", func(t *testing.T) {
		// Create a webhook first
		webhook, err := client.Webhook.Create().
			SetName("Update Test Webhook").
			SetURL("https://example.com/webhook").
			SetEvent("onPromptFinished").
			SetCreatorID(user.ID).
			SetProjectID(project.ID).
			Save(context.Background())
		require.NoError(t, err)
		require.NotNil(t, webhook)

		args := updateWebhookArgs{
			ID: int32(webhook.ID),
			Data: updateWebhookData{
				Name:        stringPtr("Updated Webhook"),
				Description: stringPtr("Updated description"),
				URL:         stringPtr("https://updated.example.com/webhook"),
				Enabled:     boolPtr(false),
			},
		}

		resolver := QueryResolver{}
		resp, err := resolver.UpdateWebhook(ctx, args)
		require.NoError(t, err)
		require.Equal(t, "Updated Webhook", resp.Name())
		require.Equal(t, "Updated description", resp.Description())
		require.Equal(t, "https://updated.example.com/webhook", resp.URL())
		require.False(t, resp.Enabled())
	})

	t.Run("UpdateWebhookInvalidEvent", func(t *testing.T) {
		// Create a webhook first
		webhook, err := client.Webhook.Create().
			SetName("Update Test Webhook").
			SetURL("https://example.com/webhook").
			SetEvent("onPromptFinished").
			SetCreatorID(user.ID).
			SetProjectID(project.ID).
			Save(context.Background())
		require.NoError(t, err)
		require.NotNil(t, webhook)

		args := updateWebhookArgs{
			ID: int32(webhook.ID),
			Data: updateWebhookData{
				Event: stringPtr("invalidEvent"),
			},
		}

		resolver := QueryResolver{}
		_, err := resolver.UpdateWebhook(ctx, args)
		require.Error(t, err)
		require.Contains(t, err.Error(), "only onPromptFinished event is supported")
	})

	t.Run("DeleteWebhook", func(t *testing.T) {
		// Create a webhook first
		webhook, err := client.Webhook.Create().
			SetName("Delete Test Webhook").
			SetURL("https://example.com/webhook").
			SetEvent("onPromptFinished").
			SetCreatorID(user.ID).
			SetProjectID(project.ID).
			Save(context.Background())
		require.NoError(t, err)
		require.NotNil(t, webhook)

		args := deleteWebhookArgs{
			ID: int32(webhook.ID),
		}

		resolver := QueryResolver{}
		result, err := resolver.DeleteWebhook(ctx, args)
		require.NoError(t, err)
		require.True(t, result)

		// Verify webhook is deleted
		_, err = client.Webhook.Get(context.Background(), webhook.ID)
		require.Error(t, err)
		require.True(t, ent.IsNotFound(err))
	})

	t.Run("UnauthorizedAccess", func(t *testing.T) {
		// Create another user without permissions
		unauthorizedUser, err := client.User.Create().
			SetName("Unauthorized User").
			SetAddr("unauthorized@example.com").
			SetEmail("unauthorized@example.com").
			SetPhone("").
			SetLang("en").
			SetLevel(0). // No admin level
			Save(context.Background())
		require.NoError(t, err)
		require.NotNil(t, unauthorizedUser)

		unauthorizedCtx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
			UserID: unauthorizedUser.ID,
		})

		// Try to create webhook without permissions
		args := createWebhookArgs{
			Data: createWebhookData{
				Name:      "Unauthorized Webhook",
				URL:       "https://example.com/webhook",
				Event:     "onPromptFinished",
				ProjectID: int32(project.ID),
			},
		}

		resolver := QueryResolver{}
		_, err := resolver.CreateWebhook(unauthorizedCtx, args)
		require.Error(t, err)
		require.Contains(t, err.Error(), "insufficient permissions")
	})
}

func TestWebhookResponseMethods(t *testing.T) {
	// Create test database
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	// Create test user
	user, err := client.User.Create().
		SetName("Test User").
		SetAddr("test@example.com").
		SetEmail("test@example.com").
		SetPhone("").
		SetLang("en").
		SetLevel(255).
		Save(context.Background())
	require.NoError(t, err)
	require.NotNil(t, user)

	// Create test project
	project, err := client.Project.Create().
		SetName("Test Project").
		SetCreatorID(user.ID).
		SetProviderId(1).
		Save(context.Background())
	require.NoError(t, err)
	require.NotNil(t, project)

	// Create test webhook
	webhook, err := client.Webhook.Create().
		SetName("Test Webhook").
		SetDescription("Test Description").
		SetURL("https://example.com/webhook").
		SetEvent("onPromptFinished").
		SetEnabled(true).
		SetCreatorID(user.ID).
		SetProjectID(project.ID).
		Save(context.Background())
	require.NoError(t, err)
	require.NotNil(t, webhook)

	resp := webhookResponse{w: webhook}

	t.Run("TestResponseMethods", func(t *testing.T) {
		require.Equal(t, int32(webhook.ID), resp.ID())
		require.Equal(t, "Test Webhook", resp.Name())
		require.Equal(t, "Test Description", resp.Description())
		require.Equal(t, "https://example.com/webhook", resp.URL())
		require.Equal(t, "onPromptFinished", resp.Event())
		require.True(t, resp.Enabled())
		require.NotEmpty(t, resp.CreatedAt())
		require.NotEmpty(t, resp.UpdatedAt())
	})

	t.Run("TestCreatorRelation", func(t *testing.T) {
		creator, err := resp.Creator(context.Background())
		require.NoError(t, err)
		require.Equal(t, user.ID, creator.u.ID)
	})

	t.Run("TestProjectRelation", func(t *testing.T) {
		proj, err := resp.Project(context.Background())
		require.NoError(t, err)
		require.Equal(t, project.ID, proj.p.ID)
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}