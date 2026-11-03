create table accounts (
  id bigserial primary key,
  name text not null,
  starting_balance_cents bigint not null default 0,
  currency text not null default 'EUR',
  created_at timestamptz not null default now()
);

create table categories (
  id bigserial primary key,
  name text not null,
  kind text not null check (kind in ('income','expense'))
);

create table recurring_definitions (
  id bigserial primary key,
  label text not null,
  category_id bigint not null references categories(id),
  account_id bigint not null references accounts(id),
  amount_cents bigint not null,
  frequency text not null check (frequency in ('monthly','quarterly','yearly')),
  anchor_date date not null,
  active boolean not null default true,
  created_at timestamptz not null default now()
);

create table payments (
  id bigserial primary key,
  recurring_definition_id bigint references recurring_definitions(id),
  label text not null,
  category_id bigint not null references categories(id),
  account_id bigint not null references accounts(id),
  amount_cents bigint not null,
  due_date date not null,
  paid_date date,
  created_at timestamptz not null default now(),
  unique (recurring_definition_id, due_date)
);

create table balance_adjustments (
  id bigserial primary key,
  account_id bigint not null references accounts(id),
  amount_cents bigint not null,
  note text,
  occurred_on date not null,
  created_at timestamptz not null default now()
);
