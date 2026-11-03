<script setup lang="ts">
import type { Account, Category, Payment, RecurringDefinition, YearlyOverview } from '~/types/api'

const api = useApi()
const currentYear = new Date().getFullYear()
const year = ref(currentYear)

const { data: overview, error: overviewError, refresh } = await useAsyncData<YearlyOverview>(
  'yearly-overview',
  () => api.get(`/overview/yearly?year=${year.value}`)
)
const { data: categories } = await useAsyncData<Category[]>('categories', () => api.get('/categories'), { default: () => [] })
const { data: accounts } = await useAsyncData<Account[]>('accounts', () => api.get('/accounts'), { default: () => [] })

watch(year, () => refresh())

const isCurrentYear = computed(() => year.value === currentYear)

function prevYear() {
  year.value -= 1
}

function nextYear() {
  year.value += 1
}

function goToday() {
  year.value = currentYear
}

const categoryName = (id: string) => categories.value?.find(c => c.id === id)?.name ?? id
const accountName = (id: string) => accounts.value?.find(a => a.id === id)?.name ?? id
const accountNumber = (id: string) => accounts.value?.find(a => a.id === id)?.account_number ?? null

function euros(cents: number) {
  return (cents / 100).toFixed(2)
}

const selectedPayment = ref<Payment | null>(null)
const showAddRecurring = ref(false)

async function onSaved() {
  selectedPayment.value = null
  await refresh()
}

async function onRecurringCreated(_definition: RecurringDefinition) {
  showAddRecurring.value = false
  await refresh()
}
</script>

<template>
  <div>
    <div class="page-header">
      <h1>Yearly overview</h1>
      <div class="year-nav">
        <button type="button" @click="prevYear">&lsaquo;</button>
        <span class="year-label">{{ year }}</span>
        <button type="button" @click="nextYear">&rsaquo;</button>
        <button type="button" :disabled="isCurrentYear" @click="goToday">Today</button>
      </div>
    </div>

    <p v-if="overviewError" class="error">
      Could not load overview: {{ overviewError.message }}
    </p>

    <template v-else-if="overview">
      <section class="summary">
        <div class="card">
          <h3>Paid</h3>
          <p>€{{ euros(overview.totals.paid_cents) }}</p>
        </div>
        <div class="card">
          <h3>Remaining</h3>
          <p>€{{ euros(overview.totals.remaining_cents) }}</p>
        </div>
        <div class="card">
          <h3>Monthly buffer needed</h3>
          <p>€{{ euros(overview.totals.monthly_buffer_cents) }}</p>
        </div>
      </section>

      <section>
        <h2>Accounts vs remaining yearly costs</h2>
        <ul class="accounts">
          <li v-for="a in overview.accounts" :key="a.account_id">
            {{ a.name }}
            <span v-if="accountNumber(a.account_id)" class="account-number">({{ accountNumber(a.account_id) }})</span>
            : €{{ euros(a.current_balance_cents) }} balance,
            €{{ euros(a.remaining_yearly_cost_cents) }} remaining costs
            <strong :class="a.covers_remaining ? 'covers' : 'short'">
              {{ a.covers_remaining ? 'covered' : 'short' }}
            </strong>
          </li>
        </ul>
      </section>

      <section>
        <div class="payments-header">
          <h2>Payments</h2>
          <button v-if="isCurrentYear" type="button" @click="showAddRecurring = true">Add yearly cost</button>
        </div>
        <div v-if="overview.payments.length === 0">No yearly payments.</div>
        <PaymentRow
          v-for="p in overview.payments"
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

    <AddRecurringModal
      v-if="showAddRecurring"
      :categories="categories ?? []"
      :accounts="accounts ?? []"
      :fixed-interval-months="12"
      @close="showAddRecurring = false"
      @created="onRecurringCreated"
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

.year-nav {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.year-label {
  min-width: 4rem;
  text-align: center;
  font-weight: 600;
}

.summary {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
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

.covers {
  color: #15803d;
}

.short {
  color: #b91c1c;
}

.error {
  color: #b91c1c;
}
</style>
