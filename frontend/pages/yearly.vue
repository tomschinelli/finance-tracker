<script setup lang="ts">
import type { Account, Category, Payment, YearlyOverview } from '~/types/api'

const api = useApi()
const year = new Date().getFullYear()

const { data: overview, error: overviewError, refresh } = await useAsyncData<YearlyOverview>(
  'yearly-overview',
  () => api.get(`/overview/yearly?year=${year}`)
)
const { data: categories } = await useAsyncData<Category[]>('categories', () => api.get('/categories'), { default: () => [] })
const { data: accounts } = await useAsyncData<Account[]>('accounts', () => api.get('/accounts'), { default: () => [] })

const categoryName = (id: string) => categories.value?.find(c => c.id === id)?.name ?? id
const accountName = (id: string) => accounts.value?.find(a => a.id === id)?.name ?? id

function euros(cents: number) {
  return (cents / 100).toFixed(2)
}

const selectedPayment = ref<Payment | null>(null)

async function onSaved() {
  selectedPayment.value = null
  await refresh()
}
</script>

<template>
  <div>
    <h1>Yearly overview ({{ year }})</h1>

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
      </section>

      <section>
        <h2>Accounts vs remaining yearly costs</h2>
        <ul class="accounts">
          <li v-for="a in overview.accounts" :key="a.account_id">
            {{ a.name }}: €{{ euros(a.current_balance_cents) }} balance,
            €{{ euros(a.remaining_yearly_cost_cents) }} remaining costs
            <strong :class="a.covers_remaining ? 'covers' : 'short'">
              {{ a.covers_remaining ? 'covered' : 'short' }}
            </strong>
          </li>
        </ul>
      </section>

      <section>
        <h2>Payments</h2>
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
  </div>
</template>

<style scoped>
.summary {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
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
