<script setup lang="ts">
import type { Account } from '~/types/api'

const api = useApi()

const { data: accounts, error: accountsError, refresh } = await useAsyncData<Account[]>(
  'accounts-page',
  () => api.get('/accounts'),
  { default: () => [] }
)

function euros(cents: number) {
  return (cents / 100).toFixed(2)
}

function todayIso() {
  return new Date().toISOString().slice(0, 10)
}

const openAdjustmentFor = ref<string | null>(null)
const amount = ref('')
const note = ref('')
const occurredOn = ref(todayIso())
const saving = ref(false)
const error = ref('')

function openAdjustment(accountId: string) {
  openAdjustmentFor.value = accountId
  amount.value = ''
  note.value = ''
  occurredOn.value = todayIso()
  error.value = ''
}

async function submitAdjustment() {
  if (!openAdjustmentFor.value) return
  saving.value = true
  error.value = ''
  try {
    await api.post(`/accounts/${openAdjustmentFor.value}/adjustments`, {
      amount_cents: Math.round(parseFloat(amount.value) * 100),
      note: note.value,
      occurred_on: occurredOn.value
    })
    openAdjustmentFor.value = null
    await refresh()
  } catch (err: any) {
    error.value = err.message
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <h1>Accounts</h1>

    <p v-if="accountsError" class="error">
      Could not load accounts: {{ accountsError.message }}
    </p>

    <ul v-else class="accounts">
      <li v-for="a in accounts" :key="a.id" class="account">
        <div class="account-row">
          <span class="name">{{ a.name }}</span>
          <span class="balance">{{ a.currency }} {{ euros(a.current_balance_cents) }}</span>
          <button type="button" @click="openAdjustment(a.id)">Add adjustment</button>
        </div>

        <div v-if="openAdjustmentFor === a.id" class="adjustment-form">
          <label>
            Amount (+/- cents as euros)
            <input v-model="amount" type="number" step="0.01">
          </label>
          <label>
            Note
            <input v-model="note" type="text">
          </label>
          <label>
            Date
            <input v-model="occurredOn" type="date">
          </label>
          <p v-if="error" class="error">{{ error }}</p>
          <div class="actions">
            <button type="button" @click="openAdjustmentFor = null">Cancel</button>
            <button type="button" :disabled="saving" @click="submitAdjustment">
              {{ saving ? 'Saving...' : 'Save' }}
            </button>
          </div>
        </div>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.accounts {
  list-style: none;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.account {
  background: #fff;
  border-radius: 8px;
  padding: 1rem;
}

.account-row {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.name {
  font-weight: 600;
  flex: 1;
}

.adjustment-form {
  margin-top: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-width: 280px;
}

.adjustment-form label {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: 0.9rem;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.error {
  color: #b91c1c;
}
</style>
