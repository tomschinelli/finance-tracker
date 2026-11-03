-- name: CreateAccount :one
insert into accounts (name, starting_balance_cents, currency)
values ($1, $2, $3)
returning *;

-- name: ListAccounts :many
select * from accounts order by id;

-- name: GetAccount :one
select * from accounts where id = $1;

-- name: GetDefaultAccount :one
select * from accounts where is_default = true limit 1;

-- name: UnsetDefaultAccounts :exec
update accounts set is_default = false where is_default = true;

-- name: MarkAccountDefault :one
update accounts set is_default = true where id = $1
returning *;

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

-- name: FindAdjustmentsByNote :many
select * from balance_adjustments where note = $1 order by id;

-- name: CreateCategory :one
insert into categories (name, kind)
values ($1, $2)
returning *;

-- name: ListCategories :many
select * from categories order by id;

-- name: GetCategory :one
select * from categories where id = $1;

-- name: GetDefaultCategory :one
select * from categories where is_default = true limit 1;

-- name: CreateDefaultCategory :one
insert into categories (name, kind, is_default)
values ('Others', 'expense', true)
returning *;

-- name: RenameCategory :one
update categories set name = $2 where id = $1
returning *;

-- name: DeleteCategory :exec
delete from categories where id = $1;

-- name: CreateRecurringDefinition :one
insert into recurring_definitions (label, category_id, account_id, amount_cents, interval_months, anchor_date, active, tags)
values ($1, $2, $3, $4, $5, $6, $7, $8)
returning *;

-- name: ListRecurringDefinitions :many
select * from recurring_definitions order by id;

-- name: ListRecurringDefinitionsByTag :many
select * from recurring_definitions where sqlc.arg('tag')::text = any(tags) order by id;

-- name: ListActiveRecurringDefinitions :many
select * from recurring_definitions where active = true order by id;

-- name: GetRecurringDefinition :one
select * from recurring_definitions where id = $1;

-- name: UpdateRecurringDefinition :one
update recurring_definitions set
  amount_cents = coalesce(sqlc.narg('amount_cents'), amount_cents),
  account_id = coalesce(sqlc.narg('account_id'), account_id),
  active = coalesce(sqlc.narg('active'), active),
  tags = coalesce(sqlc.narg('tags'), tags)
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

-- name: ListPaymentsByAccount :many
select * from payments
where account_id = $1
order by due_date desc, id desc;

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
where rd.interval_months = 12 and extract(year from p.due_date) = $1::int
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

-- name: ListAllPayments :many
select * from payments order by id;

-- name: ListAllBalanceAdjustments :many
select * from balance_adjustments order by id;

-- name: ClearPayments :exec
delete from payments;

-- name: ClearRecurringDefinitions :exec
delete from recurring_definitions;

-- name: ClearBalanceAdjustments :exec
delete from balance_adjustments;

-- name: ClearAccounts :exec
delete from accounts;

-- name: ClearCategories :exec
delete from categories;

-- name: InsertAccountWithID :exec
insert into accounts (id, name, starting_balance_cents, currency, is_default, created_at)
values ($1, $2, $3, $4, $5, $6);

-- name: InsertCategoryWithID :exec
insert into categories (id, name, kind, is_default)
values ($1, $2, $3, $4);

-- name: InsertRecurringDefinitionWithID :exec
insert into recurring_definitions (id, label, category_id, account_id, amount_cents, interval_months, anchor_date, active, created_at, tags)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: InsertPaymentWithID :exec
insert into payments (id, recurring_definition_id, label, category_id, account_id, amount_cents, due_date, paid_date, created_at)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: InsertBalanceAdjustmentWithID :exec
insert into balance_adjustments (id, account_id, amount_cents, note, occurred_on, created_at)
values ($1, $2, $3, $4, $5, $6);

-- name: ResetAccountsSeq :exec
select setval(pg_get_serial_sequence('accounts', 'id'), coalesce((select max(id) from accounts), 1), (select max(id) from accounts) is not null);

-- name: ResetCategoriesSeq :exec
select setval(pg_get_serial_sequence('categories', 'id'), coalesce((select max(id) from categories), 1), (select max(id) from categories) is not null);

-- name: ResetRecurringDefinitionsSeq :exec
select setval(pg_get_serial_sequence('recurring_definitions', 'id'), coalesce((select max(id) from recurring_definitions), 1), (select max(id) from recurring_definitions) is not null);

-- name: ResetPaymentsSeq :exec
select setval(pg_get_serial_sequence('payments', 'id'), coalesce((select max(id) from payments), 1), (select max(id) from payments) is not null);

-- name: ResetBalanceAdjustmentsSeq :exec
select setval(pg_get_serial_sequence('balance_adjustments', 'id'), coalesce((select max(id) from balance_adjustments), 1), (select max(id) from balance_adjustments) is not null);

-- name: SumActiveDefinitionsByIntervalKind :many
select rd.interval_months, c.kind, coalesce(sum(rd.amount_cents), 0)::bigint as total
from recurring_definitions rd
join categories c on c.id = rd.category_id
where rd.active = true
group by rd.interval_months, c.kind;
