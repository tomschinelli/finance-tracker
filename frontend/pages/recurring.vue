<script setup lang="ts">
import type { Account, Category, RecurringDefinition } from '~/types/api'

const api = useApi()

const tagFilter = ref('')

const { data: recurring, error: recurringError, refresh } = await useAsyncData<RecurringDefinition[]>(
  'recurring',
  () => api.get(tagFilter.value ? `/recurring?tag=${encodeURIComponent(tagFilter.value)}` : '/recurring'),
  { default: () => [], watch: [tagFilter] }
)
const { data: categories } = await useAsyncData<Category[]>('categories', () => api.get('/categories'), { default: () => [] })
const { data: accounts } = await useAsyncData<Account[]>('accounts', () => api.get('/accounts'), { default: () => [] })

const categoryName = (id: string) => categories.value?.find(c => c.id === id)?.name ?? id

const allTags = computed(() => {
  const set = new Set<string>()
  for (const def of recurring.value ?? []) {
    for (const tag of def.tags) set.add(tag)
  }
  return [...set].sort()
})

function euros(cents: number) {
  return (cents / 100).toFixed(2)
}

function intervalLabel(months: number) {
  if (months === 1) return 'Monthly'
  if (months === 12) return 'Yearly'
  return `Every ${months} months`
}

function tagsToText(tags: string[]) {
  return tags.join(', ')
}

function textToTags(text: string) {
  return text.split(',').map(t => t.trim()).filter(Boolean)
}

async function updateTags(def: RecurringDefinition, text: string) {
  const tags = textToTags(text)
  if (JSON.stringify(tags) === JSON.stringify(def.tags)) return
  await api.patch(`/recurring/${def.id}`, { tags })
  await refresh()
}

async function toggleActive(def: RecurringDefinition) {
  await api.patch(`/recurring/${def.id}`, { active: !def.active })
  await refresh()
}

async function updateAmount(def: RecurringDefinition, value: string) {
  const cents = Math.round(parseFloat(value) * 100)
  if (Number.isNaN(cents) || cents === def.amount_cents) return
  await api.patch(`/recurring/${def.id}`, { amount_cents: cents })
  await refresh()
}

async function updateAccount(def: RecurringDefinition, accountId: string) {
  if (accountId === def.account_id) return
  await api.patch(`/recurring/${def.id}`, { account_id: accountId })
  await refresh()
}

const newLabel = ref('')
const newCategoryId = ref('')
const newAccountId = ref('')
const newAmount = ref('')
const newIntervalKind = ref<'monthly' | 'yearly' | 'custom'>('monthly')
const newCustomInterval = ref('2')
const newAnchorDate = ref(new Date().toISOString().slice(0, 10))
const newTags = ref('')
const creating = ref(false)
const createError = ref('')

const newIntervalMonths = computed(() => {
  if (newIntervalKind.value === 'monthly') return 1
  if (newIntervalKind.value === 'yearly') return 12
  return parseInt(newCustomInterval.value, 10)
})

async function createRecurring() {
  creating.value = true
  createError.value = ''
  try {
    await api.post('/recurring', {
      label: newLabel.value,
      category_id: newCategoryId.value,
      account_id: newAccountId.value,
      amount_cents: Math.round(parseFloat(newAmount.value) * 100),
      interval_months: newIntervalMonths.value,
      anchor_date: newAnchorDate.value,
      active: true,
      tags: textToTags(newTags.value)
    })
    newLabel.value = ''
    newCategoryId.value = ''
    newAccountId.value = ''
    newAmount.value = ''
    newTags.value = ''
    await refresh()
  } catch (err: any) {
    createError.value = err.message
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <div>
    <h1>Recurring definitions</h1>

    <div class="filter-row">
      <label>
        Filter by tag
        <input v-model="tagFilter" type="text" list="known-tags" placeholder="e.g. housing">
        <datalist id="known-tags">
          <option v-for="tag in allTags" :key="tag" :value="tag" />
        </datalist>
      </label>
      <button v-if="tagFilter" type="button" @click="tagFilter = ''">Clear</button>
    </div>

    <p v-if="recurringError" class="error">
      Could not load recurring definitions: {{ recurringError.message }}
    </p>

    <table v-else class="recurring-table">
      <thead>
        <tr>
          <th>Label</th>
          <th>Category</th>
          <th>Account</th>
          <th>Amount</th>
          <th>Frequency</th>
          <th>Anchor date</th>
          <th>Tags</th>
          <th>Active</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="def in recurring" :key="def.id" :class="{ inactive: !def.active }">
          <td>{{ def.label }}</td>
          <td>{{ categoryName(def.category_id) }}</td>
          <td>
            <select :value="def.account_id" @change="updateAccount(def, ($event.target as HTMLSelectElement).value)">
              <option v-for="a in accounts" :key="a.id" :value="a.id">{{ a.name }}</option>
            </select>
          </td>
          <td>
            <input
              type="number" step="0.01" :value="euros(def.amount_cents)"
              @change="updateAmount(def, ($event.target as HTMLInputElement).value)"
            >
          </td>
          <td>{{ intervalLabel(def.interval_months) }}</td>
          <td>{{ def.anchor_date }}</td>
          <td>
            <input
              type="text" :value="tagsToText(def.tags)" placeholder="comma-separated"
              @change="updateTags(def, ($event.target as HTMLInputElement).value)"
            >
          </td>
          <td>
            <input type="checkbox" :checked="def.active" @change="toggleActive(def)">
          </td>
        </tr>
      </tbody>
    </table>

    <h2>Add recurring definition</h2>
    <form class="add-form" @submit.prevent="createRecurring">
      <label>
        Label
        <input v-model="newLabel" type="text" required>
      </label>
      <label>
        Category
        <select v-model="newCategoryId" required>
          <option value="" disabled>Select category</option>
          <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
      </label>
      <label>
        Account
        <select v-model="newAccountId" required>
          <option value="" disabled>Select account</option>
          <option v-for="a in accounts" :key="a.id" :value="a.id">{{ a.name }}</option>
        </select>
      </label>
      <label>
        Amount
        <input v-model="newAmount" type="number" step="0.01" required>
      </label>
      <label>
        Frequency
        <select v-model="newIntervalKind">
          <option value="monthly">monthly</option>
          <option value="yearly">yearly</option>
          <option value="custom">custom interval</option>
        </select>
      </label>
      <label v-if="newIntervalKind === 'custom'">
        Every how many months
        <input v-model="newCustomInterval" type="number" min="1" step="1" required>
      </label>
      <label>
        Tags
        <input v-model="newTags" type="text" placeholder="comma-separated">
      </label>
      <label>
        Anchor date
        <input v-model="newAnchorDate" type="date" required>
      </label>
      <p v-if="createError" class="error">{{ createError }}</p>
      <button type="submit" :disabled="creating">{{ creating ? 'Saving...' : 'Add' }}</button>
    </form>
  </div>
</template>

<style scoped>
.filter-row {
  display: flex;
  align-items: flex-end;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.filter-row label {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: 0.9rem;
}

.recurring-table {
  width: 100%;
  border-collapse: collapse;
  background: #fff;
  margin-bottom: 1.5rem;
}

.recurring-table th,
.recurring-table td {
  padding: 0.5rem 0.75rem;
  border-bottom: 1px solid #e5e7eb;
  text-align: left;
}

.recurring-table tr.inactive {
  color: #9ca3af;
}

.add-form {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-width: 320px;
  background: #fff;
  padding: 1rem;
  border-radius: 8px;
}

.add-form label {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: 0.9rem;
}

.error {
  color: #b91c1c;
}
</style>
