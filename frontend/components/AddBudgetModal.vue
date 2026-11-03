<script setup lang="ts">
import type { Account, Transfer } from '~/types/api'

const props = defineProps<{
  accounts: Account[]
  amountCents: number
  note: string
}>()

const emit = defineEmits<{
  close: []
  created: [transfer: Transfer]
}>()

function todayIso() {
  return new Date().toISOString().slice(0, 10)
}

const defaultAccount = computed(() => props.accounts.find(a => a.is_default))

const fromAccountId = ref(defaultAccount.value?.id ?? '')
const toAccountId = ref('')
const amount = ref(euros(props.amountCents))
const saving = ref(false)
const error = ref('')

const fromOptions = computed(() => props.accounts.filter(a => a.id !== toAccountId.value))
const toOptions = computed(() => props.accounts.filter(a => a.id !== fromAccountId.value))

const api = useApi()

function euros(cents: number) {
  return (cents / 100).toFixed(2)
}

async function submit() {
  saving.value = true
  error.value = ''
  try {
    const created = await api.post<Transfer>('/transfers', {
      from_account_id: fromAccountId.value,
      to_account_id: toAccountId.value,
      amount_cents: Math.round(parseFloat(amount.value) * 100),
      note: props.note,
      occurred_on: todayIso()
    })
    emit('created', created)
  } catch (err: any) {
    error.value = err.message
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="overlay" @click.self="emit('close')">
    <div class="modal">
      <h2>Add budget to account</h2>
      <p>
        This moves the amount below from the account you pick into the chosen account, and
        marks the budget as set for this month. It defaults to this month's estimated costs
        (€{{ euros(amountCents) }}), but can be overridden.
      </p>
      <label>
        From account
        <select v-model="fromAccountId">
          <option value="" disabled>Select account</option>
          <option v-for="a in fromOptions" :key="a.id" :value="a.id">{{ a.name }}</option>
        </select>
      </label>
      <label>
        To account
        <select v-model="toAccountId">
          <option value="" disabled>Select account</option>
          <option v-for="a in toOptions" :key="a.id" :value="a.id">{{ a.name }}</option>
        </select>
      </label>
      <label>
        Amount
        <input v-model="amount" type="number" step="0.01">
      </label>
      <p v-if="error" class="error">{{ error }}</p>
      <div class="actions">
        <button type="button" @click="emit('close')">Cancel</button>
        <button type="button" :disabled="saving || !fromAccountId || !toAccountId || !amount" @click="submit">
          {{ saving ? 'Saving...' : 'Add budget to account' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal {
  background: #fff;
  border-radius: 8px;
  padding: 1.5rem;
  width: 320px;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.modal p {
  margin: 0;
  font-size: 0.85rem;
  color: #4b5563;
}

label {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: 0.9rem;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.error {
  color: #b91c1c;
  font-size: 0.85rem;
}
</style>
