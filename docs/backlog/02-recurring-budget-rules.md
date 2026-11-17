# Recurring & budget rules

All touch occurrence generation and the budget calc (`internal/finance/`).

## Recurring costs with multiple occurrences per period (FEAT-010)

Some recurring expenses happen more than once within a single period (e.g.
fuel costs several times a month) but should still be grouped under one
shared name rather than tracked as separate recurring definitions. Each
occurrence should still count toward the budget individually, so the
monthly set-aside calculation needs to sum all occurrences for the period,
not assume at most one per `RecurringDefinition`.

## Recurring item amount follows last Paid amount (FEAT-012)

When a Payment of a Recurring item is marked Paid with a Paid amount that
differs from its Expected amount, update the Recurring item's amount to
the Paid amount, so later Expected Payments and budget figures follow
(e.g. salary that varies month to month). Applies to every Recurring item,
no "variable" flag. Open: whether already-generated future Expected
Payments (e.g. projected non-monthly ones) get their amount updated too.

## Yearly overview covers all non-monthly expenses (FEAT-013)

Today the yearly overview only includes Recurring items with an interval
of exactly 12 months, doesn't filter by Kind (yearly income counts as a
cost), and computes its own "monthly buffer" (yearly total / 12) that
differs from the monthly overview's Buffer. Make it cover all non-monthly
*expense* Recurring items, so its buffer figure equals the Buffer (see
`GLOSSARY.md`: Yearly overview, Buffer, Covers).
