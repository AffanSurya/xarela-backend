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

type RetirementPlan struct {
	ent.Schema
}

func (RetirementPlan) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.String("name").NotEmpty(),
		field.Int16("current_age"),
		field.Int16("target_retirement_age"),
		field.Other("target_annual_expense", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Other("inflation_rate", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(8,4)"}),
		field.Other("expected_return_rate", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(8,4)"}),
		field.Other("safe_withdrawal_rate", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(8,4)"}),
		field.Other("target_corpus", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (RetirementPlan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("retirement_plans").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (RetirementPlan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "name").Unique(),
	}
}

type RetirementScenario struct {
	ent.Schema
}

func (RetirementScenario) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("plan_id", uuid.UUID{}),
		field.String("scenario_name").NotEmpty(),
		field.String("risk_profile").NotEmpty(),
		field.Other("expected_return_rate", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(8,4)"}),
		field.Other("inflation_rate", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(8,4)"}),
		field.Other("monthly_contribution", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Other("projected_corpus_at_retirement", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Other("success_probability", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(5,2)"}),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (RetirementScenario) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("retirement_scenarios").
			Field("user_id").
			Unique().
			Required(),
		edge.From("plan", RetirementPlan.Type).
			Ref("scenarios").
			Field("plan_id").
			Unique().
			Required(),
	}
}

type PortfolioSnapshot struct {
	ent.Schema
}

func (PortfolioSnapshot) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.Time("snapshot_date").SchemaType(map[string]string{dialect.Postgres: "date"}),
		field.Other("total_assets", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Other("total_liabilities", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Other("net_worth", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Other("crypto_value", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Other("stock_value", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Other("cash_value", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Other("annualized_expense", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Other("savings_rate", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(8,4)"}),
		field.Time("created_at").Default(time.Now),
	}
}

func (PortfolioSnapshot) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("portfolio_snapshots").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (PortfolioSnapshot) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "snapshot_date").Unique(),
	}
}
