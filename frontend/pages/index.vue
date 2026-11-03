<script setup lang="ts">
import type { Account, Category, MonthlyOverview, Payment, RecurringDefinition } from '~/types/api'

const api = useApi()

const now = new Date()
const year = ref(now.getFullYear())
const month = ref(now.getMonth() + 1)

const { data: overview, error: overviewError, refresh } = await useAsyncData<MonthlyOverview>(
  'monthly-overview',
  () => api.get(`/overview/monthly?year=${year.value}&month=${month.value}`)
)
const { data: categories } = await useAsyncData<Category[]>('categories', () => api.get('/categories'), { default: () => [] })
const { data: accounts } = await useAsyncData<Account[]>('accounts', () => api.get('/accounts'), { default: () => [] })

const showUpcoming = ref(false)

watch([year, month], () => {
  showUpcoming.value = false
  refresh()
})

const monthLabel = computed(() => {
  const d = new Date(year.value, month.value - 1, 1)
  return d.toLocaleString('default', { month: 'long', year: 'numeric' })
})

const isCurrentMonth = computed(() => year.value === now.getFullYear() && month.value === now.getMonth() + 1)

function prevMonth() {
  if (month.value === 1) {
    month.value = 12
    year.value -= 1
  } else {
    month.value -= 1
  }
}

function nextMonth() {
  if (month.value === 12) {
    month.value = 1
    year.value += 1
  } else {
    month.value += 1
  }
}

function goToday() {
  year.value = now.getFullYear()
  month.value = now.getMonth() + 1
}

const categoryName = (id: string) => categories.value?.find(c => c.id === id)?.name ?? id
const accountName = (id: string) => accounts.value?.find(a => a.id === id)?.name ?? id
const accountNumber = (id: string) => accounts.value?.find(a => a.id === id)?.account_number ?? null

const unpaidPayments = computed(() => overview.value?.payments.filter(p => p.paid_date === null) ?? [])
const paidPayments = computed(() => overview.value?.payments.filter(p => p.paid_date !== null) ?? [])

const selectedPayment = ref<Payment | null>(null)
const showAddAdhoc = ref(false)
const showAddRecurring = ref(false)
const showAddBudget = ref(false)

function euros(cents: number) {
  return (cents / 100).toFixed(2)
}

async function onSaved() {
  selectedPayment.value = null
  await refresh()
}

async function onAdhocCreated(_payment: Payment) {
  showAddAdhoc.value = false
  await refresh()
}

async function onRecurringCreated(_definition: RecurringDefinition) {
  showAddRecurring.value = false
  await refresh()
}

async function onBudgetCreated() {
  showAddBudget.value = false
  await refresh()
}
</script>

<template>
  <div>
    <div class="page-header">
      <h1>Monthly overview</h1>
      <div class="month-nav">
        <button type="button" @click="prevMonth">&lsaquo;</button>
        <span class="month-label">{{ monthLabel }}</span>
        <button type="button" @click="nextMonth">&rsaquo;</button>
        <button type="button" :disabled="isCurrentMonth" @click="goToday">Today</button>
      </div>
    </div>

    <p v-if="overviewError" class="error">
      Could not load overview: {{ overviewError.message }}
    </p>

    <template v-else-if="overview">
      <section class="summary">
        <div class="card">
          <h3>Predicted costs</h3>
          <p>€{{ euros(overview.budget.predicted_costs_cents) }}</p>
        </div>
        <div class="card">
          <h3>Buffer for non-monthly costs</h3>
          <p>€{{ euros(overview.budget.buffer_cents) }}</p>
        </div>
        <div class="card">
          <h3>Paid against fixed costs</h3>
          <p>€{{ euros(overview.budget.paid_against_fixed_costs_cents) }}</p>
        </div>
        <div class="card">
          <h3>Budget remaining</h3>
          <p>€{{ euros(overview.budget.remaining_cents) }}</p>
        </div>
        <div class="card">
          <h3>Remaining expected costs</h3>
          <p>€{{ euros(overview.remaining_expected_cents) }}</p>
        </div>
      </section>

      <section class="budget-action">
        <p v-if="overview.budget.costs_already_set && overview.budget.buffer_already_set">
          Budget set for this month.
        </p>
        <template v-else>
          <p>
            Costs: €{{ euros(overview.budget.predicted_costs_cents) }}
            ({{ overview.budget.costs_already_set ? 'set' : 'not set' }}),
            buffer: €{{ euros(overview.budget.buffer_cents) }}
            ({{ overview.budget.buffer_already_set ? 'set' : 'not set' }}).
          </p>
          <button type="button" @click="showAddBudget = true">Add budget to account</button>
        </template>
      </section>

      <section>
        <h2>Accounts</h2>
        <ul class="accounts">
          <li v-for="a in overview.accounts" :key="a.account_id">
            {{ a.name }}
            <span v-if="accountNumber(a.account_id)" class="account-number">({{ accountNumber(a.account_id) }})</span>
            : €{{ euros(a.current_balance_cents) }}
          </li>
        </ul>
      </section>

      <section>
        <div class="payments-header">
          <h2>Payments this month</h2>
          <div class="payments-actions">
            <button type="button" @click="showAddRecurring = true">Add recurring cost</button>
            <button type="button" @click="showAddAdhoc = true">Add ad hoc cost</button>
          </div>
        </div>

        <template v-if="overview.overdue_payments.length">
          <h3>Overdue</h3>
          <PaymentRow
            v-for="p in overview.overdue_payments"
            :key="p.id"
            :payment="p"
            :category-name="categoryName(p.category_id)"
            :account-name="accountName(p.account_id)"
            overdue
            @select="selectedPayment = $event"
          />
        </template>

        <h3>Unpaid</h3>
        <div v-if="unpaidPayments.length === 0">No unpaid payments this month.</div>
        <PaymentRow
          v-for="p in unpaidPayments"
          :key="p.id"
          :payment="p"
          :category-name="categoryName(p.category_id)"
          :account-name="accountName(p.account_id)"
          @select="selectedPayment = $event"
        />

        <div class="upcoming-toggle">
          <template v-if="overview.upcoming_payments.length === 0">
            Nothing else expected for the rest of the year.
          </template>
          <template v-else-if="!showUpcoming">
            {{ overview.upcoming_payments.length }} upcoming non-monthly cost{{ overview.upcoming_payments.length === 1 ? '' : 's' }}
            projected later this year.
            <button type="button" @click="showUpcoming = true">Show</button>
          </template>
          <template v-else>
            <button type="button" @click="showUpcoming = false">Hide upcoming</button>
          </template>
        </div>
        <template v-if="showUpcoming">
          <PaymentRow
            v-for="p in overview.upcoming_payments"
            :key="p.id"
            :payment="p"
            :category-name="categoryName(p.category_id)"
            :account-name="accountName(p.account_id)"
            @select="selectedPayment = $event"
          />
        </template>

        <h3>Paid</h3>
        <div v-if="paidPayments.length === 0">No paid payments this month.</div>
        <PaymentRow
          v-for="p in paidPayments"
          :key="p.id"
          :payment="p"
          :category-name="categoryName(p.category_id)"
          :account-name="accountName(p.account_id)"
          @select="selectedPayment = $event"
        />
      </section>
    </template>

    <MarkPaidModal
      v-if="selectedPayment"
      :payment="selectedPayment"
      :categories="categories ?? []"
      :accounts="accounts ?? []"
      @close="selectedPayment = null"
      @saved="onSaved"
    />

    <AddAdhocModal
      v-if="showAddAdhoc"
      :categories="categories ?? []"
      :accounts="accounts ?? []"
      @close="showAddAdhoc = false"
      @created="onAdhocCreated"
    />

    <AddRecurringModal
      v-if="showAddRecurring"
      :categories="categories ?? []"
      :accounts="accounts ?? []"
      @close="showAddRecurring = false"
      @created="onRecurringCreated"
    />

    <AddBudgetModal
      v-if="showAddBudget && overview"
      :accounts="accounts ?? []"
      :costs-amount-cents="overview.budget.predicted_costs_cents"
      :buffer-amount-cents="overview.budget.buffer_cents"
      :costs-already-set="overview.budget.costs_already_set"
      :buffer-already-set="overview.budget.buffer_already_set"
      :costs-note="overview.budget.costs_marker_note"
      :buffer-note="overview.budget.buffer_marker_note"
      @close="showAddBudget = false"
      @created="onBudgetCreated"
    />
  </div>
</template>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.month-nav {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.month-label {
  min-width: 10rem;
  text-align: center;
  font-weight: 600;
}

.summary {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.card {
  background: #fff;
  border-radius: 8px;
  padding: 1rem;
  text-align: center;
}

.card h3 {
  margin: 0 0 0.5rem;
  font-size: 0.85rem;
  color: #6b7280;
}

.card p {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
}

.budget-action {
  background: #fff;
  border-radius: 8px;
  padding: 1rem;
  margin-bottom: 1.5rem;
}

.budget-action p {
  margin: 0 0 0.5rem;
}

.accounts {
  list-style: none;
  padding: 0;
}

.account-number {
  font-size: 0.85rem;
  color: #6b7280;
}

.payments-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.payments-actions {
  display: flex;
  gap: 0.5rem;
}

.upcoming-toggle {
  padding: 0.6rem 0.75rem;
  font-size: 0.85rem;
  color: #6b7280;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.error {
  color: #b91c1c;
}
</style>
