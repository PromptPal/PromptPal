package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"time"
)

// WebhookCall holds the schema definition for the WebhookCall entity.
type WebhookCall struct {
	ent.Schema
}

// Fields of the WebhookCall.
func (WebhookCall) Fields() []ent.Field {
	return []ent.Field{
		field.Int("webhook_id").StorageKey("webhook_call_webhook"),
		field.String("trace_id").Comment("Trace ID to bind the prompt call"),
		field.String("url").Comment("The URL that was called"),
		field.JSON("request_headers", map[string]string{}).Optional().Comment("HTTP request headers sent"),
		field.Text("request_body").Comment("Request payload sent to webhook"),
		field.Int("status_code").Optional().Comment("HTTP response status code"),
		field.JSON("response_headers", map[string]string{}).Optional().Comment("HTTP response headers received"),
		field.Text("response_body").Optional().Comment("Response body received from webhook"),
		field.Time("start_time").Default(time.Now).Comment("When the webhook call started"),
		field.Time("end_time").Optional().Comment("When the webhook call completed"),
		field.Bool("is_timeout").Default(false).Comment("Whether the call timed out"),
		field.Bool("is_success").Default(false).Comment("Whether the call was successful (status 2xx)"),
		field.String("error_message").Optional().Comment("Error message if call failed"),
		field.String("user_agent").Optional().Comment("User-Agent header sent with request"),
	}
}

// Edges of the WebhookCall.
func (WebhookCall) Edges() []ent.Edge {
	return []ent.Edge{
		edge.
			From("webhook", Webhook.Type).
			Ref("calls").
			Unique().
			Field("webhook_id").
			Required(),
	}
}

func (WebhookCall) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}