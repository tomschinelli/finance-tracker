-- name: CreateAccount :one
insert into accounts (name, starting_balance_cents, currency)
values ($1, $2, $3)
returning *;

-- name: ListAccounts :many
select * from accounts order by id;

-- name: GetAccount :one
select * from accounts where id = $1;

-- name: SumPaidPaymentsByAccount :one
select coalesce(sum(case when c.kind = 'income' then p.amount_cents else -p.amount_cents end), 0)::bigint as total
from payments p
join categories c on c.id = p.category_id
where p.account_id = $1 and p.paid_date is not null;

-- name: SumAdjustmentsByAccount :one
select coalesce(sum(amount_cents), 0)::bigint as total
from balance_adjustments
where account_id = $1;

-- name: CreateAdjustment :one
insert into balance_adjustments (account_id, amount_cents, note, occurred_on)
values ($1, $2, $3, $4)
returning *;

-- name: ListAdjustmentsByAccount :many
select * from balance_adjustments where account_id = $1 order by occurred_on, id;

-- name: CreateCategory :one
insert into categories (name, kind)
values ($1, $2)
returning *;

-- name: ListCategories :many
select * from categories order by id;

-- name: GetCategory :one
select * from categories where id = $1;

-- name: CreateRecurringDefinition :one
insert into recurring_definitions (label, category_id, account_id, amount_cents, frequency, anchor_date, active)
values ($1, $2, $3, $4, $5, $6, $7)
returning *;

-- name: ListRecurringDefinitions :many
select * from recurring_definitions order by id;

-- name: ListActiveRecurringDefinitions :many
select * from recurring_definitions where active = true order by id;

-- name: GetRecurringDefinition :one
select * from recurring_definitions where id = $1;

-- name: UpdateRecurringDefinition :one
update recurring_definitions set
  amount_cents = coalesce(sqlc.narg('amount_cents'), amount_cents),
  account_id = coalesce(sqlc.narg('account_id'), account_id),
  active = coalesce(sqlc.narg('active'), active)
where id = sqlc.arg('id')
returning *;

-- name: CreatePayment :one
insert into payments (recurring_definition_id, label, category_id, account_id, amount_cents, due_date, paid_date)
values (null, $1, $2, $3, $4, $5, $6)
returning *;

-- name: InsertOccurrence :many
insert into payments (recurring_definition_id, label, category_id, account_id, amount_cents, due_date, paid_date)
values ($1, $2, $3, $4, $5, $6, null)
on conflict (recurring_definition_id, due_date) do nothing
returning *;

-- name: GetPayment :one
select * from payments where id = $1;

-- name: ListPaymentsByYear :many
select * from payments
where extract(year from due_date) = $1::int
order by due_date, id;

-- name: ListPaymentsByYearMonth :many
select * from payments
where extract(year from due_date) = $1::int and extract(month from due_date) = $2::int
order by due_date, id;

-- name: ListYearlyFrequencyPayments :many
select p.* from payments p
join recurring_definitions rd on rd.id = p.recurring_definition_id
where rd.frequency = 'yearly' and extract(year from p.due_date) = $1::int
order by p.due_date, p.id;

-- name: UpdatePayment :one
update payments set
  paid_date = coalesce(sqlc.narg('paid_date'), paid_date),
  amount_cents = coalesce(sqlc.narg('amount_cents'), amount_cents),
  account_id = coalesce(sqlc.narg('account_id'), account_id),
  label = coalesce(sqlc.narg('label'), label),
  category_id = coalesce(sqlc.narg('category_id'), category_id)
where id = sqlc.arg('id')
returning *;

-- name: SumActiveDefinitionsByFrequencyKind :many
select rd.frequency, c.kind, coalesce(sum(rd.amount_cents), 0)::bigint as total
from recurring_definitions rd
join categories c on c.id = rd.category_id
where rd.active = true
group by rd.frequency, c.kind;
