<script setup lang="ts">
import type { Account, ExportData } from '~/types/api'

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
const otherAccountId = ref('')
const saving = ref(false)
const error = ref('')

const otherAccountOptions = computed(() =>
  accounts.value?.filter(a => a.id !== openAdjustmentFor.value) ?? []
)

function openAdjustment(account: Account) {
  openAdjustmentFor.value = account.id
  amount.value = euros(account.current_balance_cents)
  note.value = ''
  occurredOn.value = todayIso()
  otherAccountId.value = ''
  error.value = ''
}

async function submitAdjustment() {
  if (!openAdjustmentFor.value) return
  const account = accounts.value?.find(a => a.id === openAdjustmentFor.value)
  if (!account) return

  const deltaCents = Math.round(parseFloat(amount.value) * 100) - account.current_balance_cents
  if (!account.is_default) {
    if (!otherAccountId.value) {
      error.value = 'Select the account this amount moves with'
      return
    }
    if (deltaCents === 0) {
      error.value = 'No change to apply'
      return
    }
  }

  saving.value = true
  error.value = ''
  try {
    if (account.is_default) {
      await api.post(`/accounts/${openAdjustmentFor.value}/adjustments`, {
        amount_cents: deltaCents,
        note: note.value,
        occurred_on: occurredOn.value
      })
    } else if (deltaCents > 0) {
      await api.post('/transfers', {
        from_account_id: otherAccountId.value,
        to_account_id: account.id,
        amount_cents: deltaCents,
        note: note.value,
        occurred_on: occurredOn.value
      })
    } else {
      await api.post('/transfers', {
        from_account_id: account.id,
        to_account_id: otherAccountId.value,
        amount_cents: -deltaCents,
        note: note.value,
        occurred_on: occurredOn.value
      })
    }
    openAdjustmentFor.value = null
    await refresh()
  } catch (err: any) {
    error.value = err.message
  } finally {
    saving.value = false
  }
}

const settingDefaultFor = ref<string | null>(null)

async function setDefault(account: Account) {
  settingDefaultFor.value = account.id
  try {
    await api.post(`/accounts/${account.id}/default`, {})
    await refresh()
  } catch (err: any) {
    error.value = err.message
  } finally {
    settingDefaultFor.value = null
  }
}

const clearing = ref(false)

async function clearData() {
  if (!confirm('Clear ALL data, including accounts and categories? The database will be completely empty afterwards. This cannot be undone.')) {
    return
  }
  clearing.value = true
  error.value = ''
  try {
    await api.post('/data/clear', {})
    await refresh()
  } catch (err: any) {
    error.value = err.message
  } finally {
    clearing.value = false
  }
}

const exporting = ref(false)

async function exportData() {
  exporting.value = true
  error.value = ''
  try {
    const data = await api.get<ExportData>('/data/export')
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `finance-export-${todayIso()}.json`
    a.click()
    URL.revokeObjectURL(url)
  } catch (err: any) {
    error.value = err.message
  } finally {
    exporting.value = false
  }
}

const importing = ref(false)
const importInput = ref<HTMLInputElement | null>(null)

function triggerImport() {
  importInput.value?.click()
}

async function handleImportFile(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return

  if (!confirm('Import this file? It will replace ALL data, including accounts and categories, and cannot be undone.')) {
    return
  }

  importing.value = true
  error.value = ''
  try {
    const text = await file.text()
    const data = JSON.parse(text)
    await api.post('/data/import', data)
    await refresh()
  } catch (err: any) {
    error.value = err.message
  } finally {
    importing.value = false
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
      <div class="header-actions">
        <input ref="importInput" type="file" accept="application/json" class="hidden-input" @change="handleImportFile">
        <button type="button" :disabled="exporting" @click="exportData">
          {{ exporting ? 'Exporting...' : 'Export data' }}
        </button>
        <button type="button" :disabled="importing" @click="triggerImport">
          {{ importing ? 'Importing...' : 'Import data' }}
        </button>
        <button type="button" class="danger" :disabled="clearing" @click="clearData">
          {{ clearing ? 'Clearing...' : 'Clear data' }}
        </button>
        <button type="button" @click="openNewAccount">Add account</button>
      </div>
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
          <span v-if="a.is_default" class="default-badge">default</span>
          <span class="balance">{{ a.currency }} {{ euros(a.current_balance_cents) }}</span>
          <button
            v-if="!a.is_default"
            type="button"
            :disabled="settingDefaultFor === a.id"
            @click="setDefault(a)"
          >
            {{ settingDefaultFor === a.id ? 'Setting...' : 'Set as default' }}
          </button>
          <button type="button" @click="openAdjustment(a)">Add adjustment</button>
        </div>

        <div v-if="openAdjustmentFor === a.id" class="adjustment-form">
          <label>
            New balance ({{ a.currency }})
            <input v-model="amount" type="number" step="0.01">
          </label>
          <label v-if="!a.is_default">
            Account this moves with
            <select v-model="otherAccountId">
              <option value="" disabled>Select account</option>
              <option v-for="o in otherAccountOptions" :key="o.id" :value="o.id">{{ o.name }}</option>
            </select>
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
            <button type="button" :disabled="saving || (!a.is_default && !otherAccountId)" @click="submitAdjustment">
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

.header-actions {
  display: flex;
  gap: 0.5rem;
}

.danger {
  color: #b91c1c;
  border-color: #b91c1c;
}

.hidden-input {
  display: none;
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

.default-badge {
  font-size: 0.75rem;
  color: #6b7280;
  border: 1px solid #d1d5db;
  border-radius: 999px;
  padding: 0.1rem 0.5rem;
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
