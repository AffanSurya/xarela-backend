CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL,
    password_hash text NOT NULL,
    full_name text,
    is_email_verified boolean NOT NULL DEFAULT false,
    base_currency char(3) NOT NULL,
    timezone text,
    is_2fa_enabled boolean NOT NULL DEFAULT false,
    last_login_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT users_email_key UNIQUE (email),
    CONSTRAINT users_base_currency_check CHECK (base_currency ~ '^[A-Z]{3}$')
);

CREATE TABLE user_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash text NOT NULL,
    ip_address text,
    user_agent text,
    device_id text,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_sessions_refresh_token_hash_key UNIQUE (refresh_token_hash)
);

CREATE TABLE connected_accounts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    account_type text NOT NULL,
    provider text NOT NULL,
    account_name text NOT NULL,
    external_account_ref text NOT NULL,
    encrypted_credentials jsonb,
    credentials_key_id text,
    credentials_rotated_at timestamptz,
    is_read_only boolean NOT NULL DEFAULT true,
    sync_status text NOT NULL,
    last_synced_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT connected_accounts_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT connected_accounts_user_key UNIQUE (id, user_id),
    CONSTRAINT connected_accounts_unique_provider UNIQUE (user_id, provider, external_account_ref)
);

CREATE TABLE assets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_type text NOT NULL,
    symbol text NOT NULL,
    name text NOT NULL,
    chain text,
    chain_id text,
    contract_address text,
    token_standard text,
    isin text,
    quote_currency char(3) NOT NULL,
    metadata jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT assets_quote_currency_check CHECK (quote_currency ~ '^[A-Z]{3}$'),
    CONSTRAINT assets_chain_contract_unique UNIQUE (chain_id, contract_address),
    CONSTRAINT assets_isin_unique UNIQUE (isin)
);

CREATE TABLE holdings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    connected_account_id uuid NOT NULL,
    asset_id uuid NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    quantity numeric(20,8) NOT NULL,
    avg_cost numeric(20,8) NOT NULL,
    cost_currency char(3) NOT NULL,
    market_value numeric(20,8) NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT holdings_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT holdings_connected_account_fkey FOREIGN KEY (connected_account_id, user_id) REFERENCES connected_accounts(id, user_id) ON DELETE CASCADE,
    CONSTRAINT holdings_user_key UNIQUE (id, user_id),
    CONSTRAINT holdings_unique_position UNIQUE (user_id, connected_account_id, asset_id),
    CONSTRAINT holdings_cost_currency_check CHECK (cost_currency ~ '^[A-Z]{3}$'),
    CONSTRAINT holdings_quantity_positive CHECK (quantity > 0),
    CONSTRAINT holdings_avg_cost_positive CHECK (avg_cost >= 0),
    CONSTRAINT holdings_market_value_positive CHECK (market_value >= 0)
);

CREATE TABLE transactions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    connected_account_id uuid NOT NULL,
    asset_id uuid NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    txn_type text NOT NULL,
    provider text NOT NULL,
    provider_txn_id text NOT NULL,
    quantity numeric(20,8) NOT NULL,
    unit_price numeric(20,8) NOT NULL,
    fee_amount numeric(20,8) NOT NULL,
    fee_currency char(3) NOT NULL,
    occurred_at timestamptz NOT NULL,
    notes text,
    raw_payload jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT transactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT transactions_connected_account_fkey FOREIGN KEY (connected_account_id, user_id) REFERENCES connected_accounts(id, user_id) ON DELETE CASCADE,
    CONSTRAINT transactions_user_key UNIQUE (id, user_id),
    CONSTRAINT transactions_idempotency_key UNIQUE (provider, provider_txn_id, connected_account_id),
    CONSTRAINT transactions_quantity_positive CHECK (quantity > 0),
    CONSTRAINT transactions_unit_price_positive CHECK (unit_price >= 0),
    CONSTRAINT transactions_fee_amount_positive CHECK (fee_amount >= 0),
    CONSTRAINT transactions_fee_currency_check CHECK (fee_currency ~ '^[A-Z]{3}$')
);

CREATE TABLE price_ticks (
    id bigserial PRIMARY KEY,
    asset_id uuid NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    price numeric(20,8) NOT NULL,
    quote_currency char(3) NOT NULL,
    source text NOT NULL,
    observed_at timestamptz NOT NULL,
    CONSTRAINT price_ticks_quote_currency_check CHECK (quote_currency ~ '^[A-Z]{3}$')
);

CREATE INDEX price_ticks_asset_observed_idx ON price_ticks (asset_id, observed_at DESC);

CREATE TABLE expense_categories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id uuid,
    name text NOT NULL,
    color text,
    is_system boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT expense_categories_user_key UNIQUE (id, user_id),
    CONSTRAINT expense_categories_parent_fkey FOREIGN KEY (parent_id, user_id) REFERENCES expense_categories(id, user_id) ON DELETE SET NULL,
    CONSTRAINT expense_categories_unique_name UNIQUE (user_id, parent_id, name)
);

CREATE TABLE expenses (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    category_id uuid NOT NULL,
    amount numeric(20,8) NOT NULL,
    currency char(3) NOT NULL,
    occurred_on date NOT NULL,
    merchant text,
    note text,
    payment_method text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT expenses_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT expenses_category_fkey FOREIGN KEY (category_id, user_id) REFERENCES expense_categories(id, user_id) ON DELETE CASCADE,
    CONSTRAINT expenses_currency_check CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT expenses_amount_positive CHECK (amount > 0)
);

CREATE TABLE recurring_expenses (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    category_id uuid NOT NULL,
    amount numeric(20,8) NOT NULL,
    currency char(3) NOT NULL,
    cadence text NOT NULL,
    next_run_on date NOT NULL,
    end_on date,
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT recurring_expenses_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT recurring_expenses_category_fkey FOREIGN KEY (category_id, user_id) REFERENCES expense_categories(id, user_id) ON DELETE CASCADE,
    CONSTRAINT recurring_expenses_currency_check CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT recurring_expenses_amount_positive CHECK (amount > 0)
);

CREATE TABLE budgets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    category_id uuid NOT NULL,
    period_month date NOT NULL,
    amount_limit numeric(20,8) NOT NULL,
    currency char(3) NOT NULL,
    alert_threshold_percent numeric(5,2) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT budgets_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT budgets_category_fkey FOREIGN KEY (category_id, user_id) REFERENCES expense_categories(id, user_id) ON DELETE CASCADE,
    CONSTRAINT budgets_currency_check CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT budgets_amount_positive CHECK (amount_limit > 0),
    CONSTRAINT budgets_alert_threshold_check CHECK (alert_threshold_percent >= 0 AND alert_threshold_percent <= 100),
    CONSTRAINT budgets_unique_month UNIQUE (user_id, category_id, period_month)
);

CREATE TABLE retirement_plans (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name text NOT NULL,
    current_age smallint NOT NULL,
    target_retirement_age smallint NOT NULL,
    target_annual_expense numeric(20,8) NOT NULL,
    inflation_rate numeric(8,4) NOT NULL,
    expected_return_rate numeric(8,4) NOT NULL,
    safe_withdrawal_rate numeric(8,4) NOT NULL,
    target_corpus numeric(20,8) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT retirement_plans_user_key UNIQUE (id, user_id),
    CONSTRAINT retirement_plans_unique_name UNIQUE (user_id, name),
    CONSTRAINT retirement_plans_age_check CHECK (current_age >= 0 AND target_retirement_age >= current_age),
    CONSTRAINT retirement_plans_positive_amounts CHECK (target_annual_expense > 0 AND target_corpus > 0),
    CONSTRAINT retirement_plans_rate_check CHECK (
        inflation_rate >= 0 AND expected_return_rate >= 0 AND safe_withdrawal_rate > 0
    )
);

CREATE TABLE retirement_scenarios (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    plan_id uuid NOT NULL,
    scenario_name text NOT NULL,
    risk_profile text NOT NULL,
    expected_return_rate numeric(8,4) NOT NULL,
    inflation_rate numeric(8,4) NOT NULL,
    monthly_contribution numeric(20,8) NOT NULL,
    projected_corpus_at_retirement numeric(20,8) NOT NULL,
    success_probability numeric(5,2) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT retirement_scenarios_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT retirement_scenarios_plan_fkey FOREIGN KEY (plan_id, user_id) REFERENCES retirement_plans(id, user_id) ON DELETE CASCADE,
    CONSTRAINT retirement_scenarios_rates_check CHECK (expected_return_rate >= 0 AND inflation_rate >= 0),
    CONSTRAINT retirement_scenarios_probability_check CHECK (success_probability >= 0 AND success_probability <= 100),
    CONSTRAINT retirement_scenarios_contribution_positive CHECK (monthly_contribution >= 0),
    CONSTRAINT retirement_scenarios_projection_positive CHECK (projected_corpus_at_retirement >= 0)
);

CREATE TABLE portfolio_snapshots (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    snapshot_date date NOT NULL,
    total_assets numeric(20,8) NOT NULL,
    total_liabilities numeric(20,8) NOT NULL,
    net_worth numeric(20,8) NOT NULL,
    crypto_value numeric(20,8) NOT NULL,
    stock_value numeric(20,8) NOT NULL,
    cash_value numeric(20,8) NOT NULL,
    annualized_expense numeric(20,8) NOT NULL,
    savings_rate numeric(8,4) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT portfolio_snapshots_unique_day UNIQUE (user_id, snapshot_date),
    CONSTRAINT portfolio_snapshots_non_negative CHECK (
        total_assets >= 0 AND total_liabilities >= 0 AND net_worth >= 0 AND crypto_value >= 0 AND stock_value >= 0 AND cash_value >= 0 AND annualized_expense >= 0
    ),
    CONSTRAINT portfolio_snapshots_savings_rate_check CHECK (savings_rate >= 0 AND savings_rate <= 100)
);

CREATE TABLE alerts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    alert_type text NOT NULL,
    severity text NOT NULL,
    title text NOT NULL,
    payload jsonb,
    is_read boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    read_at timestamptz
);

CREATE INDEX alerts_user_read_idx ON alerts (user_id, is_read, created_at);

CREATE TABLE sync_jobs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    connected_account_id uuid NOT NULL,
    job_type text NOT NULL,
    status text NOT NULL,
    attempt integer NOT NULL DEFAULT 0,
    last_error text,
    scheduled_at timestamptz,
    started_at timestamptz,
    finished_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT sync_jobs_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT sync_jobs_connected_account_fkey FOREIGN KEY (connected_account_id, user_id) REFERENCES connected_accounts(id, user_id) ON DELETE CASCADE,
    CONSTRAINT sync_jobs_attempt_check CHECK (attempt >= 0)
);

CREATE INDEX sync_jobs_user_status_idx ON sync_jobs (user_id, status, scheduled_at);

CREATE TABLE audit_logs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_type text NOT NULL,
    action text NOT NULL,
    resource_type text NOT NULL,
    resource_id uuid,
    metadata jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX audit_logs_user_created_idx ON audit_logs (user_id, created_at);