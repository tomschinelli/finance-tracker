<script setup lang="ts">
import type { Account } from '~/types/api'

const props = defineProps<{
  accounts: Account[]
  costsAmountCents: number
  bufferAmountCents: number
  costsAlreadySet: boolean
  bufferAlreadySet: boolean
  costsNote: string
  bufferNote: string
}>()

const emit = defineEmits<{
  close: []
  created: []
}>()

function todayIso() {
  return new Date().toISOString().slice(0, 10)
}

function euros(cents: number) {
  return (cents / 100).toFixed(2)
}

const defaultAccount = computed(() => props.accounts.find(a => a.is_default))

const sourceType = ref<'direct' | 'transfer'>('direct')
const fromAccountId = ref(defaultAccount.value?.id ?? '')
const toAccountId = ref('')
const costsAmount = ref(euros(props.costsAmountCents))
const bufferAmount = ref(euros(props.bufferAmountCents))
const saving = ref(false)
const error = ref('')

const fromOptions = computed(() => props.accounts.filter(a => a.id !== toAccountId.value))
const toOptions = computed(() => props.accounts.filter(a => sourceType.value === 'direct' || a.id !== fromAccountId.value))

const api = useApi()

async function depositPart(amountCents: number, note: string) {
  if (sourceType.value === 'direct') {
    await api.post(`/accounts/${toAccountId.value}/adjustments`, {
      amount_cents: amountCents,
      note,
      occurred_on: todayIso()
    })
  } else {
    await api.post('/transfers', {
      from_account_id: fromAccountId.value,
      to_account_id: toAccountId.value,
      amount_cents: amountCents,
      note,
      occurred_on: todayIso()
    })
  }
}

async function submit() {
  saving.value = true
  error.value = ''
  try {
    if (!props.costsAlreadySet && parseFloat(costsAmount.value) > 0) {
      await depositPart(Math.round(parseFloat(costsAmount.value) * 100), props.costsNote)
    }
    if (!props.bufferAlreadySet && parseFloat(bufferAmount.value) > 0) {
      await depositPart(Math.round(parseFloat(bufferAmount.value) * 100), props.bufferNote)
    }
    emit('created')
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
        Deposits the current month's costs and the non-monthly buffer into the chosen account,
        as two separate adjustments, and marks each as set for this month.
      </p>
      <label class="radio-group">
        <span>
          <input v-model="sourceType" type="radio" value="direct">
          Direct deposit
        </span>
        <span>
          <input v-model="sourceType" type="radio" value="transfer">
          Transfer from another account
        </span>
      </label>
      <label v-if="sourceType === 'transfer'">
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
        Current month's costs
        <input v-model="costsAmount" type="number" step="0.01" :disabled="costsAlreadySet">
      </label>
      <p v-if="costsAlreadySet" class="hint">Already set for this month.</p>
      <label>
        Buffer for non-monthly costs
        <input v-model="bufferAmount" type="number" step="0.01" :disabled="bufferAlreadySet">
      </label>
      <p v-if="bufferAlreadySet" class="hint">Already set for this month.</p>
      <p v-if="error" class="error">{{ error }}</p>
      <div class="actions">
        <button type="button" @click="emit('close')">Cancel</button>
        <button
          type="button"
          :disabled="saving || !toAccountId || (sourceType === 'transfer' && !fromAccountId)"
          @click="submit"
        >
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

.radio-group {
  flex-direction: row;
  gap: 1rem;
}

.radio-group span {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-weight: normal;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.hint {
  color: #6b7280;
  font-style: italic;
}

.error {
  color: #b91c1c;
  font-size: 0.85rem;
}
</style>
