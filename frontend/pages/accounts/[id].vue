<script setup lang="ts">
import type { Account, AccountAdjustment, Payment } from '~/types/api'

const route = useRoute()
const id = route.params.id as string
const api = useApi()

const { data: accounts, error: accountsError } = await useAsyncData<Account[]>(
  'accounts',
  () => api.get('/accounts'),
  { default: () => [] }
)
const account = computed(() => accounts.value?.find(a => String(a.id) === id))

const { data: adjustments, error: adjustmentsError } = await useAsyncData<AccountAdjustment[]>(
  `account-${id}-adjustments`,
  () => api.get(`/accounts/${id}/adjustments`),
  { default: () => [] }
)
const adjustmentsByRecent = computed(() => [...(adjustments.value ?? [])].reverse())

const { data: payments, error: paymentsError } = await useAsyncData<Payment[]>(
  `account-${id}-payments`,
  () => api.get(`/accounts/${id}/payments`),
  { default: () => [] }
)

function euros(cents: number) {
  return (cents / 100).toFixed(2)
}
</script>

<template>
  <div>
    <NuxtLink to="/accounts" class="back-link">&larr; Accounts</NuxtLink>

    <p v-if="accountsError" class="error">
      Could not load account: {{ accountsError.message }}
    </p>
    <p v-else-if="!account" class="error">
      Account not found.
    </p>

    <template v-else>
      <h1>{{ account.name }}</h1>

      <div class="summary">
        <div class="summary-item">
          <span class="label">Current balance</span>
          <span class="value">{{ account.currency }} {{ euros(account.current_balance_cents) }}</span>
        </div>
        <div class="summary-item">
          <span class="label">Starting balance</span>
          <span class="value">{{ account.currency }} {{ euros(account.starting_balance_cents) }}</span>
        </div>
      </div>

      <h2>Adjustment history</h2>
      <p v-if="adjustmentsError" class="error">
        Could not load adjustments: {{ adjustmentsError.message }}
      </p>
      <p v-else-if="!adjustmentsByRecent.length">No adjustments yet.</p>
      <table v-else class="adjustments">
        <thead>
          <tr>
            <th>Date</th>
            <th>Amount</th>
            <th>Note</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="adj in adjustmentsByRecent" :key="adj.id">
            <td>{{ adj.occurred_on }}</td>
            <td :class="{ negative: adj.amount_cents < 0 }">
              {{ adj.amount_cents >= 0 ? '+' : '' }}{{ euros(adj.amount_cents) }}
            </td>
            <td>{{ adj.note }}</td>
          </tr>
        </tbody>
      </table>

      <h2>Payments</h2>
      <p v-if="paymentsError" class="error">
        Could not load payments: {{ paymentsError.message }}
      </p>
      <p v-else-if="!payments?.length">No payments yet.</p>
      <table v-else class="adjustments">
        <thead>
          <tr>
            <th>Due date</th>
            <th>Label</th>
            <th>Amount</th>
            <th>Paid date</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in payments" :key="p.id">
            <td>{{ p.due_date }}</td>
            <td>{{ p.label }}</td>
            <td>{{ euros(p.amount_cents) }}</td>
            <td>{{ p.paid_date ?? 'unpaid' }}</td>
          </tr>
        </tbody>
      </table>
    </template>
  </div>
</template>

<style scoped>
.back-link {
  display: inline-block;
  margin-bottom: 1rem;
}

.summary {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.summary-item {
  background: #fff;
  border-radius: 8px;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 180px;
}

.summary-item .label {
  font-size: 0.85rem;
  color: #666;
}

.summary-item .value {
  font-size: 1.25rem;
  font-weight: 600;
}

.adjustments {
  width: 100%;
  border-collapse: collapse;
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
}

.adjustments th,
.adjustments td {
  text-align: left;
  padding: 0.6rem 1rem;
  border-bottom: 1px solid #eee;
}

.adjustments .negative {
  color: #b91c1c;
}

.error {
  color: #b91c1c;
}
</style>
