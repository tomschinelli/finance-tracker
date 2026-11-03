alter table recurring_definitions add column interval_months integer;

update recurring_definitions set interval_months = case frequency
  when 'monthly' then 1
  when 'quarterly' then 3
  when 'yearly' then 12
end;

alter table recurring_definitions alter column interval_months set not null;
alter table recurring_definitions add constraint recurring_definitions_interval_months_check check (interval_months >= 1);
alter table recurring_definitions drop column frequency;
