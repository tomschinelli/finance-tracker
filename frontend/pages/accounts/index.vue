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

function openAdjustment(account: Account) {
  openAdjustmentFor.value = account.id
  amount.value = euros(account.current_balance_cents)
  note.value = ''
  occurredOn.value = todayIso()
  error.value = ''
}

async function submitAdjustment() {
  if (!openAdjustmentFor.value) return
  const account = accounts.value?.find(a => a.id === openAdjustmentFor.value)
  if (!account) return
  saving.value = true
  error.value = ''
  try {
    const newBalanceCents = Math.round(parseFloat(amount.value) * 100)
    await api.post(`/accounts/${openAdjustmentFor.value}/adjustments`, {
      amount_cents: newBalanceCents - account.current_balance_cents,
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

const showNewAccount = ref(false)
const newAccountName = ref('')
const newAccountStartingBalance = ref('')
const newAccountCurrency = ref('EUR')
const newAccountSaving = ref(false)
const newAccountError = ref('')

function openNewAccount() {
  showNewAccount.value = true
  newAccountName.value = ''
  newAccountStartingBalance.value = ''
  newAccountCurrency.value = 'EUR'
  newAccountError.value = ''
}

async function submitNewAccount() {
  newAccountSaving.value = true
  newAccountError.value = ''
  try {
    await api.post('/accounts', {
      name: newAccountName.value,
      starting_balance_cents: Math.round(parseFloat(newAccountStartingBalance.value || '0') * 100),
      currency: newAccountCurrency.value
    })
    showNewAccount.value = false
    await refresh()
  } catch (err: any) {
    newAccountError.value = err.message
  } finally {
    newAccountSaving.value = false
  }
}
</script>

<template>
  <div>
    <div class="header-row">
      <h1>Accounts</h1>
      <button type="button" @click="openNewAccount">Add account</button>
    </div>

    <div v-if="showNewAccount" class="new-account-form">
      <label>
        Name
        <input v-model="newAccountName" type="text">
      </label>
      <label>
        Starting balance (euros)
        <input v-model="newAccountStartingBalance" type="number" step="0.01">
      </label>
      <label>
        Currency
        <input v-model="newAccountCurrency" type="text">
      </label>
      <p v-if="newAccountError" class="error">{{ newAccountError }}</p>
      <div class="actions">
        <button type="button" @click="showNewAccount = false">Cancel</button>
        <button type="button" :disabled="newAccountSaving" @click="submitNewAccount">
          {{ newAccountSaving ? 'Saving...' : 'Save' }}
        </button>
      </div>
    </div>

    <p v-if="accountsError" class="error">
      Could not load accounts: {{ accountsError.message }}
    </p>

    <ul v-else class="accounts">
      <li v-for="a in accounts" :key="a.id" class="account">
        <div class="account-row">
          <NuxtLink :to="`/accounts/${a.id}`" class="name">{{ a.name }}</NuxtLink>
          <span class="balance">{{ a.currency }} {{ euros(a.current_balance_cents) }}</span>
          <button type="button" @click="openAdjustment(a)">Add adjustment</button>
        </div>

        <div v-if="openAdjustmentFor === a.id" class="adjustment-form">
          <label>
            New balance ({{ a.currency }})
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
.header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.new-account-form {
  background: #fff;
  border-radius: 8px;
  padding: 1rem;
  margin: 0.75rem 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-width: 280px;
}

.new-account-form label {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: 0.9rem;
}

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
  color: inherit;
  text-decoration: none;
}

.name:hover {
  text-decoration: underline;
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
