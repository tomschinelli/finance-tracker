<script setup lang="ts">
import type { Category, CategoryKind } from '~/types/api'

const api = useApi()

const { data: categories, error: categoriesError, refresh } = await useAsyncData<Category[]>(
  'categories-page',
  () => api.get('/categories'),
  { default: () => [] }
)

const error = ref('')

const editingId = ref<string | null>(null)
const editingName = ref('')

function startRename(cat: Category) {
  editingId.value = cat.id
  editingName.value = cat.name
  error.value = ''
}

function cancelRename() {
  editingId.value = null
}

async function saveRename(cat: Category) {
  if (!editingName.value || editingName.value === cat.name) {
    editingId.value = null
    return
  }
  error.value = ''
  try {
    await api.patch(`/categories/${cat.id}`, { name: editingName.value })
    editingId.value = null
    await refresh()
  } catch (err: any) {
    error.value = err.message
  }
}

const deletingId = ref<string | null>(null)

async function deleteCategory(cat: Category) {
  if (!confirm(`Delete category "${cat.name}"? This cannot be undone.`)) {
    return
  }
  deletingId.value = cat.id
  error.value = ''
  try {
    await api.delete(`/categories/${cat.id}`)
    await refresh()
  } catch (err: any) {
    error.value = err.message
  } finally {
    deletingId.value = null
  }
}

const newName = ref('')
const newKind = ref<CategoryKind>('expense')
const creating = ref(false)

async function createCategory() {
  creating.value = true
  error.value = ''
  try {
    await api.post('/categories', { name: newName.value, kind: newKind.value })
    newName.value = ''
    newKind.value = 'expense'
    await refresh()
  } catch (err: any) {
    error.value = err.message
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <div>
    <h1>Categories</h1>

    <p v-if="categoriesError" class="error">
      Could not load categories: {{ categoriesError.message }}
    </p>
    <p v-else-if="error" class="error">{{ error }}</p>

    <table v-if="!categoriesError" class="categories-table">
      <thead>
        <tr>
          <th>Name</th>
          <th>Kind</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="cat in categories" :key="cat.id">
          <td>
            <template v-if="editingId === cat.id">
              <input
                v-model="editingName" type="text" autofocus
                @keyup.enter="saveRename(cat)" @keyup.escape="cancelRename"
              >
            </template>
            <template v-else>
              {{ cat.name }}
              <span v-if="cat.is_default" class="default-badge">default</span>
            </template>
          </td>
          <td>{{ cat.kind }}</td>
          <td class="actions">
            <template v-if="editingId === cat.id">
              <button type="button" @click="saveRename(cat)">Save</button>
              <button type="button" @click="cancelRename">Cancel</button>
            </template>
            <template v-else>
              <button type="button" @click="startRename(cat)">Rename</button>
              <button
                type="button" class="danger" :disabled="cat.is_default || deletingId === cat.id"
                :title="cat.is_default ? 'The default category cannot be deleted' : ''"
                @click="deleteCategory(cat)"
              >
                {{ deletingId === cat.id ? 'Deleting...' : 'Delete' }}
              </button>
            </template>
          </td>
        </tr>
      </tbody>
    </table>

    <h2>Add category</h2>
    <form class="add-form" @submit.prevent="createCategory">
      <label>
        Name
        <input v-model="newName" type="text" required>
      </label>
      <label>
        Kind
        <select v-model="newKind">
          <option value="expense">expense</option>
          <option value="income">income</option>
        </select>
      </label>
      <button type="submit" :disabled="creating">{{ creating ? 'Saving...' : 'Add' }}</button>
    </form>
  </div>
</template>

<style scoped>
.categories-table {
  width: 100%;
  border-collapse: collapse;
  background: #fff;
  margin-bottom: 1.5rem;
}

.categories-table th,
.categories-table td {
  padding: 0.5rem 0.75rem;
  border-bottom: 1px solid #e5e7eb;
  text-align: left;
}

.default-badge {
  font-size: 0.75rem;
  color: #6b7280;
  border: 1px solid #d1d5db;
  border-radius: 999px;
  padding: 0.1rem 0.5rem;
  margin-left: 0.5rem;
}

.actions {
  display: flex;
  gap: 0.5rem;
}

.danger {
  color: #b91c1c;
  border-color: #b91c1c;
}

.danger:disabled {
  color: #9ca3af;
  border-color: #d1d5db;
  cursor: not-allowed;
}

.add-form {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-width: 280px;
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
