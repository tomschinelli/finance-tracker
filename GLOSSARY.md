# Finance Management

A single person's household money: which purpose-based pots it sits in,
what is due when, and how much to set aside each month so every pot can
cover what's coming.

## Language

### Money containers

**Account**:
A purpose-based pot of money (e.g. Savings, Fixed Costs, Living), usually
mirroring one real bank account.
_Avoid_: Wallet, pot, bucket

**Starting balance**:
The balance an Account had when it was first recorded in the app.
_Avoid_: Opening balance, initial balance

**Current balance**:
An Account's balance today, always derived from its Starting balance and
everything recorded against it since; never entered directly.
_Avoid_: Balance (unqualified), stored balance

**Main account**:
The one Account income lands in and the usual source when money is set
aside into other Accounts. Exactly one Account is the Main account.
_Avoid_: Default account, primary account

**Account number**:
The real bank's identifier for the account an Account mirrors (e.g. an
IBAN), used to match bank statements to it.
_Avoid_: IBAN (as the general term), account id

### Classification

**Category**:
A label classifying money as either income or expense (e.g. Rent, Salary,
Groceries).
_Avoid_: Type, group

**Kind**:
Whether a Category is income or expense; decides whether its amounts add
to or subtract from a Current balance.
_Avoid_: Direction, sign, type

**Others**:
The fallback expense Category that always exists and can't be deleted.
_Avoid_: Uncategorized, misc

**Tag**:
A free-form label on a Recurring item for filtering; carries no meaning
beyond grouping.

### Recurring money

**Recurring item**:
An income or expense that repeats every fixed number of months (rent,
salary, insurance), with an amount, Account and Category. Its amount
always follows the latest Paid amount of its Payments, so it tracks
drifting figures like salary. Can be paused without being deleted.
_Avoid_: Recurring definition, recurring cost, recurring payment, subscription

**Interval**:
How many months lie between two occurrences of a Recurring item: 1 is
monthly, 12 is yearly, any other whole number is allowed. "Non-monthly"
means an Interval above 1.
_Avoid_: Frequency, period, cadence

**Occurrence**:
The Payment a Recurring item produces for one Due date.
_Avoid_: Instance, installment

### Ledger

**Payment**:
A single amount of money going into or out of an Account on a given date,
either income or expense. Either an Occurrence of a Recurring item or an
Ad hoc payment.
_Avoid_: Transaction, entry, booking

**Ad hoc payment**:
A one-off Payment that doesn't belong to any Recurring item (a repair, a
refund).
_Avoid_: Ad hoc cost, one-off, manual payment

**Due date**:
The date a Payment is supposed to happen. For an Occurrence this is an estimate; the real bill may arrive earlier or later.

**Expected**:
A Payment that hasn't happened yet (no paid date).
_Avoid_: Unpaid, pending, open

**Paid**:
A Payment that has actually happened, on its paid date.
_Avoid_: Settled, done, booked

**Expected amount**:
The amount a Payment is anticipated to be before it's Paid; for an
Occurrence, the item's amount at the time.

**Paid amount**:
The amount a Payment actually turned out to be when marked Paid; may
differ from its Expected amount.
_Avoid_: Actual amount, final amount

**Upcoming**:
An Expected Payment of a non-monthly Recurring item, due later in the same
year than the month being viewed.
_Avoid_: Future, scheduled

**Overdue**:
An Expected Payment of a non-monthly Recurring item whose Due date falls
in an earlier month than the one being viewed. Monthly Payments are never
Overdue; they stay in their own month.
_Avoid_: Late, missed

### Other balance changes

**Adjustment**:
A manual, one-sided change to an Account's balance by a known amount,
with an optional note, not tied to any Payment or Category.
_Avoid_: Balance adjustment, correction, deposit

**Reconciliation**:
Setting an Account's Current balance to the real bank balance at a point
in time; the app records the gap as an Untracked difference.
_Avoid_: Balance correction, set balance, sync

**Untracked difference**:
The amount by which an Account's real balance differed from its Current
balance at a Reconciliation, i.e. money that moved without being recorded.
_Avoid_: Discrepancy, drift, correction

**Transfer**:
Moving an amount from one Account to another; the only kind of entry that
touches two Accounts at once.
_Avoid_: Move, internal payment

### Budget

**Fixed costs**:
The total of all monthly expense Recurring items; what a month is
expected to cost.
_Avoid_: Predicted costs, monthly costs

**Buffer**:
The monthly share of non-monthly expense Recurring items (each item's
amount divided by its interval), saved up so those bills are covered when
they fall due.
_Avoid_: Reserve, smoothing, irregular costs

**Monthly budget**:
Fixed costs plus Buffer: the amount to set aside each month into the
Accounts that pay for them.
_Avoid_: Set-aside, budget (unqualified)

**Setting aside**:
Moving a month's Fixed costs and Buffer into their Accounts, as Transfers
from the Main account or as Adjustments. Done separately for each part,
at most once per month.
_Avoid_: Depositing, funding

**Budget remaining**:
Fixed costs minus the expense Payments already Paid this month.
_Avoid_: Remaining, remaining expected costs

**Excluded from budget**:
An expense Payment that doesn't count against Budget remaining (e.g. an
Ad hoc payment made from the fixed-costs Account that the budget never
planned for).

### Overviews

**Monthly overview**:
One month's Payments (Expected and Paid), its Monthly budget and Budget
remaining, plus Overdue and Upcoming Payments.

**Yearly overview**:
One year's non-monthly expense Payments, Paid versus still Expected, and
whether each Account covers what's still Expected.
_Avoid_: Annual overview, irregular costs overview

**Covers**:
An Account covers its remaining costs when its Current balance is at
least the total of its still-Expected expense Payments for the year.
_Avoid_: Sufficient, funded
