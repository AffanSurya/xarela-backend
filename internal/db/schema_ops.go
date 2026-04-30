package db

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Alert struct {
	ent.Schema
}

func (Alert) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.String("alert_type").NotEmpty(),
		field.String("severity").NotEmpty(),
		field.String("title").NotEmpty(),
		field.JSON("payload", map[string]any{}).Optional(),
		field.Bool("is_read").Default(false),
		field.Time("created_at").Default(time.Now),
		field.Time("read_at").Optional(),
	}
}

func (Alert) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("alerts").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (Alert) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "is_read", "created_at"),
	}
}

type SyncJob struct {
	ent.Schema
}

func (SyncJob) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("connected_account_id", uuid.UUID{}),
		field.String("job_type").NotEmpty(),
		field.String("status").NotEmpty(),
		field.Int("attempt").Default(0),
		field.Text("last_error").Optional(),
		field.Time("scheduled_at").Optional(),
		field.Time("started_at").Optional(),
		field.Time("finished_at").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (SyncJob) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("sync_jobs").
			Field("user_id").
			Unique().
			Required(),
		edge.From("connected_account", ConnectedAccount.Type).
			Ref("sync_jobs").
			Field("connected_account_id").
			Unique().
			Required(),
	}
}

func (SyncJob) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "status", "scheduled_at"),
	}
}

type AuditLog struct {
	ent.Schema
}

func (AuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.String("actor_type").NotEmpty(),
		field.String("action").NotEmpty(),
		field.String("resource_type").NotEmpty(),
		field.UUID("resource_id", uuid.UUID{}).Optional(),
		field.JSON("metadata", map[string]any{}).Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (AuditLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("audit_logs").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (AuditLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "created_at"),
	}
}
