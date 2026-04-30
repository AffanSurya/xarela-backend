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

type ConnectedAccount struct {
	ent.Schema
}

func (ConnectedAccount) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.String("account_type").NotEmpty(),
		field.String("provider").NotEmpty(),
		field.String("account_name").NotEmpty(),
		field.String("external_account_ref").NotEmpty(),
		field.JSON("encrypted_credentials", map[string]any{}).Optional(),
		field.String("credentials_key_id").Optional(),
		field.Time("credentials_rotated_at").Optional(),
		field.Bool("is_read_only").Default(true),
		field.String("sync_status").NotEmpty(),
		field.Time("last_synced_at").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (ConnectedAccount) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("connected_accounts").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (ConnectedAccount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "provider", "external_account_ref").Unique(),
	}
}

type Asset struct {
	ent.Schema
}

func (Asset) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.String("asset_type").NotEmpty(),
		field.String("symbol").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("chain").Optional(),
		field.String("chain_id").Optional(),
		field.String("contract_address").Optional(),
		field.String("token_standard").Optional(),
		field.String("isin").Optional(),
		field.String("quote_currency").MaxLen(3),
		field.JSON("metadata", map[string]any{}).Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Asset) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("chain_id", "contract_address").Unique(),
		index.Fields("isin").Unique(),
		index.Fields("symbol"),
	}
}

type Holding struct {
	ent.Schema
}

func (Holding) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("connected_account_id", uuid.UUID{}),
		field.UUID("asset_id", uuid.UUID{}),
		field.Other("quantity", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Other("avg_cost", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.String("cost_currency").MaxLen(3),
		field.Other("market_value", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Holding) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("holdings").
			Field("user_id").
			Unique().
			Required(),
		edge.From("connected_account", ConnectedAccount.Type).
			Ref("holdings").
			Field("connected_account_id").
			Unique().
			Required(),
		edge.From("asset", Asset.Type).
			Ref("holdings").
			Field("asset_id").
			Unique().
			Required(),
	}
}

func (Holding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "connected_account_id", "asset_id").Unique(),
	}
}

type Transaction struct {
	ent.Schema
}

func (Transaction) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("connected_account_id", uuid.UUID{}),
		field.UUID("asset_id", uuid.UUID{}),
		field.String("txn_type").NotEmpty(),
		field.String("provider").NotEmpty(),
		field.String("provider_txn_id").NotEmpty(),
		field.Other("quantity", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Other("unit_price", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Other("fee_amount", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.String("fee_currency").MaxLen(3),
		field.Time("occurred_at"),
		field.Text("notes").Optional(),
		field.JSON("raw_payload", map[string]any{}).Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (Transaction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("transactions").
			Field("user_id").
			Unique().
			Required(),
		edge.From("connected_account", ConnectedAccount.Type).
			Ref("transactions").
			Field("connected_account_id").
			Unique().
			Required(),
		edge.From("asset", Asset.Type).
			Ref("transactions").
			Field("asset_id").
			Unique().
			Required(),
	}
}

func (Transaction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("provider", "provider_txn_id", "connected_account_id").Unique(),
		index.Fields("user_id", "occurred_at"),
	}
}

type PriceTick struct {
	ent.Schema
}

func (PriceTick) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("asset_id", uuid.UUID{}),
		field.Other("price", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.String("quote_currency").MaxLen(3),
		field.String("source").NotEmpty(),
		field.Time("observed_at"),
	}
}

func (PriceTick) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("asset", Asset.Type).
			Ref("price_ticks").
			Field("asset_id").
			Unique().
			Required(),
	}
}

func (PriceTick) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("asset_id", "observed_at"),
	}
}
