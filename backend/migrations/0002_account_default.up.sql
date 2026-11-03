alter table accounts add column is_default boolean not null default false;
create unique index accounts_one_default_idx on accounts (is_default) where is_default;
