alter table categories add column is_default boolean not null default false;
create unique index categories_one_default_idx on categories (is_default) where is_default;
