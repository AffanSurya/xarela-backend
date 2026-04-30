package db

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ExpenseCategory struct {
	ent.Schema
}

func (ExpenseCategory) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("parent_id", uuid.UUID{}).
			Optional(),
		field.String("name").NotEmpty(),
		field.String("color").Optional(),
		field.Bool("is_system").Default(false),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (ExpenseCategory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("expense_categories").
			Field("user_id").
			Unique().
			Required(),
		edge.From("parent", ExpenseCategory.Type).
			Ref("children").
			Field("parent_id").
			Unique(),
	}
}

func (ExpenseCategory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "parent_id", "name").Unique(),
	}
}

type Expense struct {
	ent.Schema
}

func (Expense) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("category_id", uuid.UUID{}),
		field.Other("amount", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.String("currency").MaxLen(3),
		field.Time("occurred_on").SchemaType(map[string]string{dialect.Postgres: "date"}),
		field.String("merchant").Optional(),
		field.Text("note").Optional(),
		field.String("payment_method").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Expense) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("expenses").
			Field("user_id").
			Unique().
			Required(),
		edge.From("category", ExpenseCategory.Type).
			Ref("expenses").
			Field("category_id").
			Unique().
			Required(),
	}
}

func (Expense) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "occurred_on"),
	}
}

type RecurringExpense struct {
	ent.Schema
}

func (RecurringExpense) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("category_id", uuid.UUID{}),
		field.Other("amount", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.String("currency").MaxLen(3),
		field.String("cadence").NotEmpty(),
		field.Time("next_run_on").SchemaType(map[string]string{dialect.Postgres: "date"}),
		field.Time("end_on").Optional().SchemaType(map[string]string{dialect.Postgres: "date"}),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (RecurringExpense) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("recurring_expenses").
			Field("user_id").
			Unique().
			Required(),
		edge.From("category", ExpenseCategory.Type).
			Ref("recurring_expenses").
			Field("category_id").
			Unique().
			Required(),
	}
}

func (RecurringExpense) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "next_run_on"),
	}
}

type Budget struct {
	ent.Schema
}

func (Budget) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("category_id", uuid.UUID{}),
		field.Time("period_month").SchemaType(map[string]string{dialect.Postgres: "date"}),
		field.Other("amount_limit", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.String("currency").MaxLen(3),
		field.Other("alert_threshold_percent", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(5,2)"}),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Budget) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("budgets").
			Field("user_id").
			Unique().
			Required(),
		edge.From("category", ExpenseCategory.Type).
			Ref("budgets").
			Field("category_id").
			Unique().
			Required(),
	}
}

func (Budget) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "category_id", "period_month").Unique(),
	}
}
