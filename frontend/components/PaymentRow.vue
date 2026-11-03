<script setup lang="ts">
import type { Payment } from '~/types/api'

const props = defineProps<{
  payment: Payment
  categoryName?: string
  accountName?: string
  currencySymbol?: string
}>()

const emit = defineEmits<{
  select: [payment: Payment]
}>()

const isPaid = computed(() => props.payment.paid_date !== null)
const amount = computed(() => (props.payment.amount_cents / 100).toFixed(2))
</script>

<template>
  <div
    class="payment-row"
    :class="{ unpaid: !isPaid }"
    @click="emit('select', payment)"
  >
    <span class="label">{{ payment.label }}</span>
    <span class="category">{{ categoryName }}</span>
    <span class="account">{{ accountName }}</span>
    <span class="due">{{ payment.due_date }}</span>
    <span class="amount">{{ currencySymbol ?? '€' }}{{ amount }}</span>
    <span class="status">{{ isPaid ? 'paid' : 'expected' }}</span>
  </div>
</template>

<style scoped>
.payment-row {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr 1fr 80px;
  gap: 0.5rem;
  padding: 0.6rem 0.75rem;
  background: #fff;
  border-bottom: 1px solid #e5e7eb;
  cursor: pointer;
  align-items: center;
}

.payment-row:hover {
  background: #f0f4ff;
}

.payment-row.unpaid {
  color: #9ca3af;
}

.amount {
  text-align: right;
}

.status {
  text-align: right;
  font-size: 0.8rem;
}
</style>
