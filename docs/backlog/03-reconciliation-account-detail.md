# Reconciliation & account detail

FEAT-009 is loosely related; split it out during grilling if it doesn't fit.

## Reconciliation (FEAT-011)

Let the user set an Account's Current balance to the real bank balance at
any time (see `GLOSSARY.md`: Reconciliation, Untracked difference). The gap
is recorded as a relative entry, labeled distinctly from a normal
Adjustment, and the UI hints that money moved without being tracked. The
plain Adjustment stays alongside it.

**To be discussed:** turning part of an Untracked difference into normal
Payments later, once its source is known (e.g. a €300 bill found inside a
+€1,000 difference). Leaning towards: the reconciled balance never changes;
the remaining difference shifts by the added Payment's signed amount. Name
for the action ("Explain"?) and the sign handling not yet agreed.

## Account type: debit, prepaid credit, credit card (FEAT-009)

Add a `type` field to accounts (e.g. debit card, prepaid credit card,
credit card) to capture more detail about how each account is held.
Independent of the `account_number` field already added for bank import
matching (FEAT-005).
