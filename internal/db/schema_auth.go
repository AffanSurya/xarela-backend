package db

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.String("email").NotEmpty(),
		field.String("password_hash").NotEmpty(),
		field.String("full_name").Optional(),
		field.Bool("is_email_verified").Default(false),
		field.String("base_currency").MaxLen(3),
		field.String("timezone").Optional(),
		field.Bool("is_2fa_enabled").Default(false),
		field.Time("last_login_at").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("sessions", UserSession.Type),
		edge.To("connected_accounts", ConnectedAccount.Type),
		edge.To("holdings", Holding.Type),
		edge.To("transactions", Transaction.Type),
		edge.To("expense_categories", ExpenseCategory.Type),
		edge.To("expenses", Expense.Type),
		edge.To("recurring_expenses", RecurringExpense.Type),
		edge.To("budgets", Budget.Type),
		edge.To("retirement_plans", RetirementPlan.Type),
		edge.To("portfolio_snapshots", PortfolioSnapshot.Type),
		edge.To("alerts", Alert.Type),
		edge.To("sync_jobs", SyncJob.Type),
		edge.To("audit_logs", AuditLog.Type),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email").Unique(),
	}
}

type UserSession struct {
	ent.Schema
}

func (UserSession) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.String("refresh_token_hash").NotEmpty(),
		field.String("ip_address").Optional(),
		field.Text("user_agent").Optional(),
		field.String("device_id").Optional(),
		field.Time("expires_at"),
		field.Time("revoked_at").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (UserSession) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("sessions").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (UserSession) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("refresh_token_hash").Unique(),
		index.Fields("user_id", "expires_at"),
	}
}
