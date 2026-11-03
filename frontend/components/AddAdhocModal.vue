<script setup lang="ts">
import type { Account, Category, Payment } from '~/types/api'

defineProps<{
  categories: Category[]
  accounts: Account[]
}>()

const emit = defineEmits<{
  close: []
  created: [payment: Payment]
}>()

function todayIso() {
  return new Date().toISOString().slice(0, 10)
}

const label = ref('')
const categoryId = ref('')
const accountId = ref('')
const amount = ref('')
const dueDate = ref(todayIso())
const alreadyPaid = ref(false)
const excludedFromBudget = ref(false)
const saving = ref(false)
const error = ref('')

const api = useApi()

async function submit() {
  saving.value = true
  error.value = ''
  try {
    const created = await api.post<Payment>('/payments', {
      label: label.value,
      category_id: categoryId.value,
      account_id: accountId.value,
      amount_cents: Math.round(parseFloat(amount.value) * 100),
      due_date: dueDate.value,
      paid_date: alreadyPaid.value ? dueDate.value : null,
      excluded_from_budget: excludedFromBudget.value
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
      <h2>Add ad hoc cost</h2>
      <label>
        Label
        <input v-model="label" type="text">
      </label>
      <label>
        Category
        <select v-model="categoryId">
          <option value="" disabled>Select category</option>
          <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
      </label>
      <label>
        Account
        <select v-model="accountId">
          <option value="" disabled>Select account</option>
          <option v-for="a in accounts" :key="a.id" :value="a.id">{{ a.name }}</option>
        </select>
      </label>
      <label>
        Amount
        <input v-model="amount" type="number" step="0.01">
      </label>
      <label>
        Due date
        <input v-model="dueDate" type="date">
      </label>
      <label class="checkbox">
        <input v-model="alreadyPaid" type="checkbox">
        Already paid
      </label>
      <label class="checkbox">
        <input v-model="excludedFromBudget" type="checkbox">
        Exclude from monthly budget
      </label>
      <p v-if="error" class="error">{{ error }}</p>
      <div class="actions">
        <button type="button" @click="emit('close')">Cancel</button>
        <button type="button" :disabled="saving" @click="submit">
          {{ saving ? 'Saving...' : 'Add' }}
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

label.checkbox {
  flex-direction: row;
  align-items: center;
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
