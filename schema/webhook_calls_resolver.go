package schema

import (
	"context"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/webhookcall"
	"github.com/PromptPal/PromptPal/service"
)

type webhookCallArgs struct {
	Pagination paginationInput
}

// Calls method for webhook response - returns paginated webhook calls
func (w webhookResponse) Calls(ctx context.Context, args webhookCallArgs) (webhookCallsResponse, error) {
	stat := service.EntClient.WebhookCall.Query().
		Where(webhookcall.WebhookID(w.w.ID)).
		Order(ent.Desc(webhookcall.FieldID))

	return webhookCallsResponse{
		stat:       stat,
		pagination: args.Pagination,
	}, nil
}
