<script setup lang="ts">
import type { Account, Category, MonthlyOverview, Payment } from '~/types/api'

const api = useApi()

const now = new Date()
const year = now.getFullYear()
const month = now.getMonth() + 1

const { data: overview, error: overviewError, refresh } = await useAsyncData<MonthlyOverview>(
  'monthly-overview',
  () => api.get(`/overview/monthly?year=${year}&month=${month}`)
)
const { data: categories } = await useAsyncData<Category[]>('categories', () => api.get('/categories'), { default: () => [] })
const { data: accounts } = await useAsyncData<Account[]>('accounts', () => api.get('/accounts'), { default: () => [] })

const categoryName = (id: string) => categories.value?.find(c => c.id === id)?.name ?? id
const accountName = (id: string) => accounts.value?.find(a => a.id === id)?.name ?? id

const selectedPayment = ref<Payment | null>(null)
const showAddAdhoc = ref(false)

function euros(cents: number) {
  return (cents / 100).toFixed(2)
}

async function onSaved() {
  selectedPayment.value = null
  await refresh()
}

async function onCreated() {
  showAddAdhoc.value = false
  await refresh()
}
</script>

<template>
  <div>
    <h1>Monthly overview</h1>

    <p v-if="overviewError" class="error">
      Could not load overview: {{ overviewError.message }}
    </p>

    <template v-else-if="overview">
      <section class="summary">
        <div class="card">
          <h3>Set-aside budget</h3>
          <p>€{{ euros(overview.budget.set_aside_cents) }}</p>
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

      <section>
        <h2>Accounts</h2>
        <ul class="accounts">
          <li v-for="a in overview.accounts" :key="a.account_id">
            {{ a.name }}: €{{ euros(a.current_balance_cents) }}
          </li>
        </ul>
      </section>

      <section>
        <div class="payments-header">
          <h2>Payments this month</h2>
          <button type="button" @click="showAddAdhoc = true">Add ad hoc cost</button>
        </div>
        <div v-if="overview.payments.length === 0">No payments this month.</div>
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

    <AddAdhocModal
      v-if="showAddAdhoc"
      :categories="categories ?? []"
      :accounts="accounts ?? []"
      @close="showAddAdhoc = false"
      @created="onCreated"
    />
  </div>
</template>

<style scoped>
.summary {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
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

.payments-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.error {
  color: #b91c1c;
}
</style>
