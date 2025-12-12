<script setup>
import Card from './Card.vue'

defineProps({
  columns: {
    type: Array,
    required: true
    // Each column: { key: string, label: string, align?: 'left'|'center'|'right', class?: string, headerClass?: string }
  },
  items: {
    type: Array,
    default: () => []
  },
  emptyIcon: {
    type: String,
    default: 'fas fa-inbox'
  },
  emptyMessage: {
    type: String,
    default: 'No data found'
  },
  rowClass: {
    type: String,
    default: 'hover:bg-gray-50 dark:hover:bg-gray-800/30 transition'
  },
  clickableRows: {
    type: Boolean,
    default: false
  }
})

defineEmits(['row-click'])

const getAlignClass = (align) => {
  switch (align) {
    case 'center': return 'text-center'
    case 'right': return 'text-right'
    default: return 'text-left'
  }
}
</script>

<template>
  <Card :padding="false" class="overflow-hidden">
    <table class="w-full">
      <thead class="bg-gray-50 dark:bg-gray-800/50">
        <tr>
          <th
            v-for="col in columns"
            :key="col.key"
            :class="[
              'px-3 sm:px-6 py-3 sm:py-4 text-xs font-semibold text-gray-700 dark:text-gray-400 uppercase tracking-wider',
              getAlignClass(col.align),
              col.headerClass
            ]"
          >
            {{ col.label }}
          </th>
        </tr>
      </thead>
      <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
        <tr
          v-for="(item, index) in items"
          :key="item.id || item.login || index"
          :class="[rowClass, { 'cursor-pointer': clickableRows }]"
          @click="clickableRows && $emit('row-click', item)"
        >
          <td
            v-for="col in columns"
            :key="col.key"
            :class="['px-3 sm:px-6 py-3 sm:py-4', getAlignClass(col.align), col.class]"
          >
            <slot :name="col.key" :item="item" :index="index">
              {{ item[col.key] }}
            </slot>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Empty State -->
    <div v-if="!items.length" class="text-center py-12">
      <i :class="emptyIcon" class="text-4xl text-gray-400 dark:text-gray-600 mb-4"></i>
      <p class="text-gray-600 dark:text-gray-400">{{ emptyMessage }}</p>
    </div>
  </Card>
</template>
