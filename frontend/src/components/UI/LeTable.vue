<script>
export default {
  props: {
    columns: {
      type: Array,
      required: true, // [{ label: "Name", key: "name" }, { label: "Email", key: "email" }]
    },
    data: {
      type: Array,
      required: true, // [{ name: "John Doe", email: "john@example.com" }]
    },
  },
}
</script>

<template>
  <div class="px-4 sm:px-6 lg:px-8">
    <div class="sm:flex sm:items-center"></div>
    <div class="mt-8 flow-root">
      <div class="-mx-4 -my-2 sm:-mx-6 lg:-mx-8">
        <div class="inline-block min-w-full py-2 align-middle">
          <table class="min-w-full border-separate border-spacing-0">
            <thead>
              <tr>
                <th
                  v-for="(column, index) in columns"
                  :key="index"
                  scope="col"
                  class="border-vborder text-fg bg-vbg-head sticky top-0 z-10 border-b px-3 py-3.5 text-left text-sm font-semibold backdrop-blur-sm backdrop-filter"
                >
                  {{ column.label }}
                </th>
                <th
                  class="border-vborder bg-vbg-head text-fg sticky top-0 z-10 border-b px-3 py-3.5 text-left text-sm font-semibold backdrop-blur-sm backdrop-filter"
                ></th>
              </tr>
            </thead>
            <tbody>
              <!-- Allows top row insertion -->
              <slot></slot>
              <tr v-for="(row, rowIndex) in data" :key="rowIndex">
                <td
                  v-for="(column, colIndex) in columns"
                  :key="colIndex"
                  :class="[
                    rowIndex !== data.length - 1 ? 'border-vborder border-b' : '',
                    'text-fg px-3 py-4 text-sm whitespace-nowrap',
                  ]"
                >
                  {{ row[column.key] }}
                </td>
                <td
                  :class="[
                    rowIndex !== data.length - 1 ? 'border-vborder border-b' : '',
                    'relative py-4 pr-4 pl-3 text-right text-sm font-medium whitespace-nowrap',
                  ]"
                >
                  <div class="flex gap-4">
                    <a
                      href="#"
                      class="text-success hover:text-indigo-900"
                      @click="$emit('edit', row)"
                    >
                      Edit<span class="sr-only">, {{ row.name }}</span>
                    </a>

                    <a
                      href="#"
                      class="text-red-600 hover:text-red-900"
                      @click="$emit('delete', row)"
                    >
                      Delete<span class="sr-only">, {{ row.name }}</span>
                    </a>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
