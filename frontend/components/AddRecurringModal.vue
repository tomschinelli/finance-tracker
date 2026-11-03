<script setup lang="ts">
import type { Account, Category, RecurringDefinition } from '~/types/api'

const props = defineProps<{
  categories: Category[]
  accounts: Account[]
  fixedIntervalMonths?: number
}>()

const emit = defineEmits<{
  close: []
  created: [definition: RecurringDefinition]
}>()

function todayIso() {
  return new Date().toISOString().slice(0, 10)
}

function intervalKindFor(months?: number): 'monthly' | 'yearly' | 'custom' {
  if (months === 12) return 'yearly'
  if (months === undefined || months === 1) return 'monthly'
  return 'custom'
}

const label = ref('')
const categoryId = ref('')
const accountId = ref('')
const amount = ref('')
const intervalKind = ref<'monthly' | 'yearly' | 'custom'>(intervalKindFor(props.fixedIntervalMonths))
const customInterval = ref(String(props.fixedIntervalMonths ?? 2))
const anchorDate = ref(todayIso())
const saving = ref(false)
const error = ref('')

const intervalMonths = computed(() => {
  if (intervalKind.value === 'monthly') return 1
  if (intervalKind.value === 'yearly') return 12
  return parseInt(customInterval.value, 10)
})

const api = useApi()

async function submit() {
  if (!label.value || !categoryId.value || !accountId.value || !amount.value) {
    error.value = 'Label, category, account, and amount are required'
    return
  }
  saving.value = true
  error.value = ''
  try {
    const created = await api.post<RecurringDefinition>('/recurring', {
      label: label.value,
      category_id: categoryId.value,
      account_id: accountId.value,
      amount_cents: Math.round(parseFloat(amount.value) * 100),
      interval_months: intervalMonths.value,
      anchor_date: anchorDate.value,
      active: true
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
      <h2>Add recurring cost</h2>
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
        Frequency
        <select v-model="intervalKind" :disabled="fixedIntervalMonths !== undefined">
          <option value="monthly">monthly</option>
          <option value="yearly">yearly</option>
          <option value="custom">custom interval</option>
        </select>
      </label>
      <label v-if="intervalKind === 'custom'">
        Every how many months
        <input v-model="customInterval" type="number" min="1" step="1" :disabled="fixedIntervalMonths !== undefined">
      </label>
      <label>
        Anchor date
        <input v-model="anchorDate" type="date">
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
