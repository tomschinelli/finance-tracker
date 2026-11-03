alter table recurring_definitions add column tags text[] not null default '{}';
