<script setup>
import { useInvoiceStore } from '@/stores/invoice'
import { useClientStore } from '@/stores/clients'
import { PrinterIcon } from '@heroicons/vue/24/outline'
import DatePick from '@/components/invoice/DatePick.vue'

const clients = useClientStore()

const invoice = useInvoiceStore()
</script>
<template>
  <div class="grid grid-cols-8 grid-rows-4">
    <!-- Invoice Date & Due  -->
    <div class="pointer-events-none col-start-1 row-start-1 text-3xl font-bold">Invoice:</div>
    <div class="pointer-events-none col-span-1 col-start-1 row-start-2"><p>Invoice No:</p></div>
    <div class="pointer-events-none col-span-1 col-start-1 row-start-3"><p>Date Created:</p></div>
    <div class="pointer-events-none col-span-1 col-start-1 row-start-4"><p>Date Due:</p></div>
    <!-- Values -->
    <div class="col-span-2 col-start-2 row-start-2" v-text="invoice.data.invoiceNumber"></div>
    <div class="col-span-2 col-start-2 row-start-3">
      <!-- Date Created -->
      <DatePick
        @update-date="
          (formatedDate) => {
            console.log(formatedDate)
            invoice.data.createdDate = formatedDate
            console.log('Invo-store=>date-created:', invoice.data.createdDate)
          }
        "
      ></DatePick>
    </div>
    <div class="col-span-2 col-start-2 row-start-4">
      <!-- Due by Date -->
      <DatePick
        @update-date="
          (formatedDate) => {
            console.log(formatedDate)
            invoice.data.dueDate = formatedDate
            console.log('Invo-store=>due-date:', invoice.data.createdDate)
          }
        "
      ></DatePick>
    </div>
    <!-- Center Icons -->
    <div class="col-span-2 col-start-4 flex items-center justify-center">
      <button class="hover:text-acc cursor-pointer" title="create PDF">
        <PrinterIcon class="size-7"></PrinterIcon>
      </button>
    </div>
    <!-- Client Fields -->
    <div class="col-span-1 col-start-6 row-start-1"><div class="text-3xl font-bold">To:</div></div>
    <div class="col-span-1 col-start-6 row-start-2"><p>Client Name:</p></div>
    <div class="col-span-1 col-start-6 row-start-3"><p>Company:</p></div>
    <div class="col-span-1 col-start-6 row-start-4"><p>Address:</p></div>
    <!-- Values -->
    <div class="col-span-2 col-start-7 row-start-2">{{ clients.selectedClient.name }}</div>
    <div class="col-span-2 col-start-7 row-start-3">{{ clients.selectedClient.company }}</div>
    <div class="col-span-2 col-start-7 row-start-4 line-clamp-2">
      {{ clients.selectedClient.address }}
    </div>
  </div>
</template>
