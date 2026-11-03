export interface Account {
  id: string
  name: string
  currency: string
  starting_balance_cents: number
  current_balance_cents: number
}

export interface AccountAdjustment {
  id: string
  amount_cents: number
  note: string
  occurred_on: string
  created_at: string
}

export type CategoryKind = 'income' | 'expense'

export interface Category {
  id: string
  name: string
  kind: CategoryKind
}

export type RecurringFrequency = 'monthly' | 'quarterly' | 'yearly'

export interface RecurringDefinition {
  id: string
  label: string
  category_id: string
  account_id: string
  amount_cents: number
  frequency: RecurringFrequency
  anchor_date: string
  active: boolean
}

export interface Payment {
  id: string
  recurring_definition_id: string | null
  label: string
  category_id: string
  account_id: string
  amount_cents: number
  due_date: string
  paid_date: string | null
}

export interface YearlyOverview {
  year: number
  payments: Payment[]
  totals: {
    paid_cents: number
    remaining_cents: number
  }
  accounts: {
    account_id: string
    name: string
    current_balance_cents: number
    remaining_yearly_cost_cents: number
    covers_remaining: boolean
  }[]
}

export interface MonthlyOverview {
  year: number
  month: number
  payments: Payment[]
  budget: {
    set_aside_cents: number
    paid_against_fixed_costs_cents: number
    remaining_cents: number
  }
  accounts: {
    account_id: string
    name: string
    current_balance_cents: number
  }[]
  remaining_expected_cents: number
}
