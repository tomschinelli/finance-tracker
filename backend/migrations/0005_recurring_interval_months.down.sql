alter table recurring_definitions add column frequency text;

update recurring_definitions set frequency = case interval_months
  when 1 then 'monthly'
  when 12 then 'yearly'
  else 'quarterly'
end;

alter table recurring_definitions alter column frequency set not null;
alter table recurring_definitions add constraint recurring_definitions_frequency_check check (frequency in ('monthly','quarterly','yearly'));
alter table recurring_definitions drop column interval_months;
