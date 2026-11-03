<script setup lang="ts">
import type { Account, AccountAdjustment, Category, Payment } from '~/types/api'

const route = useRoute()
const id = route.params.id as string
const api = useApi()

const { data: accounts, error: accountsError, refresh: refreshAccounts } = await useAsyncData<Account[]>(
  'accounts',
  () => api.get('/accounts'),
  { default: () => [] }
)
const account = computed(() => accounts.value?.find(a => String(a.id) === id))

const editingAccountNumber = ref(false)
const accountNumberInput = ref('')
const accountNumberSaving = ref(false)
const accountNumberError = ref('')

function openAccountNumberEdit() {
  accountNumberInput.value = account.value?.account_number ?? ''
  accountNumberError.value = ''
  editingAccountNumber.value = true
}

async function saveAccountNumber() {
  if (!account.value) return
  accountNumberSaving.value = true
  accountNumberError.value = ''
  try {
    await api.patch(`/accounts/${account.value.id}`, {
      account_number: accountNumberInput.value.trim() || null
    })
    editingAccountNumber.value = false
    await refreshAccounts()
  } catch (err: any) {
    accountNumberError.value = err.message
  } finally {
    accountNumberSaving.value = false
  }
}

const { data: adjustments, error: adjustmentsError } = await useAsyncData<AccountAdjustment[]>(
  `account-${id}-adjustments`,
  () => api.get(`/accounts/${id}/adjustments`),
  { default: () => [] }
)

const { data: payments, error: paymentsError } = await useAsyncData<Payment[]>(
  `account-${id}-payments`,
  () => api.get(`/accounts/${id}/payments`),
  { default: () => [] }
)

const { data: categories } = await useAsyncData<Category[]>(
  'categories',
  () => api.get('/categories'),
  { default: () => [] }
)
const categoryKind = (categoryId: string) => categories.value?.find(c => c.id === categoryId)?.kind

interface HistoryEntry {
  key: string
  date: string
  amountCents: number
  note: string
}

const historyByRecent = computed<HistoryEntry[]>(() => {
  const fromAdjustments: HistoryEntry[] = (adjustments.value ?? []).map(adj => ({
    key: `adj-${adj.id}`,
    date: adj.occurred_on,
    amountCents: adj.amount_cents,
    note: adj.note,
  }))
  const fromPayments: HistoryEntry[] = (payments.value ?? [])
    .filter(p => p.paid_date)
    .map(p => ({
      key: `pay-${p.id}`,
      date: p.paid_date as string,
      amountCents: categoryKind(p.category_id) === 'income' ? p.amount_cents : -p.amount_cents,
      note: p.label,
    }))
  return [...fromAdjustments, ...fromPayments].sort((a, b) => b.date.localeCompare(a.date))
})

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
        <div class="summary-item">
          <span class="label">Account number</span>
          <template v-if="!editingAccountNumber">
            <span class="value account-number-row">
              {{ account.account_number ?? 'Not set' }}
              <button type="button" @click="openAccountNumberEdit">Edit</button>
            </span>
          </template>
          <div v-else class="account-number-edit">
            <input v-model="accountNumberInput" type="text">
            <p v-if="accountNumberError" class="error">{{ accountNumberError }}</p>
            <div class="actions">
              <button type="button" @click="editingAccountNumber = false">Cancel</button>
              <button type="button" :disabled="accountNumberSaving" @click="saveAccountNumber">
                {{ accountNumberSaving ? 'Saving...' : 'Save' }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <h2>Adjustment history</h2>
      <p v-if="adjustmentsError" class="error">
        Could not load adjustments: {{ adjustmentsError.message }}
      </p>
      <p v-else-if="paymentsError" class="error">
        Could not load payments: {{ paymentsError.message }}
      </p>
      <p v-else-if="!historyByRecent.length">No adjustments yet.</p>
      <table v-else class="adjustments">
        <thead>
          <tr>
            <th>Date</th>
            <th>Amount</th>
            <th>Note</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="entry in historyByRecent" :key="entry.key">
            <td>{{ entry.date }}</td>
            <td :class="{ negative: entry.amountCents < 0 }">
              {{ entry.amountCents >= 0 ? '+' : '' }}{{ euros(entry.amountCents) }}
            </td>
            <td>{{ entry.note }}</td>
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

.account-number-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.account-number-edit {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.account-number-edit .actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
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
