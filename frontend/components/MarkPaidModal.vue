<script setup lang="ts">
import type { Account, Category, Payment } from '~/types/api'

const props = defineProps<{
  payment: Payment
  categories: Category[]
  accounts: Account[]
}>()

const emit = defineEmits<{
  close: []
  saved: [payment: Payment]
}>()

function todayIso() {
  return new Date().toISOString().slice(0, 10)
}

const label = ref(props.payment.label)
const categoryId = ref(props.payment.category_id)
const accountId = ref(props.payment.account_id)
const amount = ref((props.payment.amount_cents / 100).toFixed(2))
const paidDate = ref(props.payment.paid_date ?? todayIso())
const saving = ref(false)
const error = ref('')

const api = useApi()

async function submit() {
  saving.value = true
  error.value = ''
  try {
    const updated = await api.patch<Payment>(`/payments/${props.payment.id}`, {
      label: label.value,
      category_id: categoryId.value,
      account_id: accountId.value,
      amount_cents: Math.round(parseFloat(amount.value) * 100),
      paid_date: paidDate.value
    })
    emit('saved', updated)
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
      <h2>Mark payment paid</h2>
      <label>
        Label
        <input v-model="label" type="text">
      </label>
      <label>
        Category
        <select v-model="categoryId">
          <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
      </label>
      <label>
        Account
        <select v-model="accountId">
          <option v-for="a in accounts" :key="a.id" :value="a.id">{{ a.name }}</option>
        </select>
      </label>
      <label>
        Amount
        <input v-model="amount" type="number" step="0.01">
      </label>
      <label>
        Paid date
        <input v-model="paidDate" type="date">
      </label>
      <p v-if="error" class="error">{{ error }}</p>
      <div class="actions">
        <button type="button" @click="emit('close')">Cancel</button>
        <button type="button" :disabled="saving" @click="submit">
          {{ saving ? 'Saving...' : 'Mark paid' }}
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
