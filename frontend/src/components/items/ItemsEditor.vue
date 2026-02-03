<script setup>
import { ref, computed, onMounted, reactive } from 'vue'
import { BriefcaseIcon } from '@heroicons/vue/24/outline'
import { useItemsStore } from '@/stores/items'
import { TabGroup, TabList, Tab, TabPanels, TabPanel } from '@headlessui/vue'
import { TrashIcon, PencilIcon, XCircleIcon, CheckCircleIcon } from '@heroicons/vue/24/outline'
import LeBtn from '../UI/LeBtn.vue'
import LeInput from '../UI/LeInput.vue'
import LeFlexTd from '../UI/table/LeFlexTd.vue'
import LeFlexTh from '../UI/table/LeFlexTh.vue'
import { MagnifyingGlassIcon } from '@heroicons/vue/24/outline'

const isOpen = ref(false)
const itemStore = useItemsStore()

onMounted(() => {
  itemStore.fetchItems('style', 'https://dummyjson.com/products?limit=20')
  itemStore.fetchItems('sample', 'https://dummyjson.com/products?limit=5&skip=5')
})

const searchQueries = reactive({
  style: '',
  sample: '',
})

const filteredItems = computed(() => {
  return itemStore.items[selectedEl.value].filter((item) =>
    item.name.toLowerCase().includes(searchQueries[selectedEl.value].toLowerCase()),
  )
})

const selectedEl = ref('style')

const sliderTransform = computed(() =>
  selectedEl.value === 'style' ? 'translate-x-[-51%] ' : 'translate-x-[51%]',
)

function moveSlider(tabName) {
  selectedEl.value = tabName
  console.log(selectedEl.value)
}
// Crud
const newItemForm = reactive({
  id: null,
  name: '',
  time: null,
  price: null,
})
const editForm = reactive({ id: null, name: '', time: null, price: null })
const editing = ref(false)
const editedRowId = ref(null)

function addNewItem(type, item) {
  console.log('New item sent to shop: ', type, item)

  itemStore.createItem(type, item)
  resetInputs(newItemForm)
}

function removeItem(item) {
  itemStore.deleteItem(selectedEl.value, item)
}

function editSave() {
  itemStore.updateItem(selectedEl.value, editForm)
  editing.value = false
  resetInputs(editForm)
}

function editTrue(item) {
  editedRowId.value = item.id
  Object.assign(editForm, item)
  editing.value = true
}

function cancelEdit() {
  // Reset edit state
  editing.value = false
  editedRowId.value = null
  resetInputs(editForm)
}

const resetInputs = (formName) => {
  formName.id = null
  formName.name = ''
  formName.type = ''
  formName.time = null
  formName.price = null
}
</script>
<template>
  <div class="flex flex-col items-center" title="client items">
    <!-- Text that always stays gray-900 -->

    <BriefcaseIcon
      @click="isOpen = true"
      class="hover:text-acc size-8 cursor-pointer stroke-1"
    ></BriefcaseIcon>
  </div>

  <Teleport to="body">
    <transition name="fadeItIn">
      <div
        v-if="isOpen"
        class="fixed inset-0 top-0 right-0 z-[900] h-[100vh] w-[100vw] bg-black/50 select-none"
      ></div>
    </transition>

    <transition mode="out-in" name="slideItIn">
      <div
        class="bg-primary fixed top-0 right-0 z-[901] h-[100vh] w-[50vw] 2xl:w-1/3 dark:bg-neutral-800"
        v-if="isOpen"
      >
        <div class="">
          <div class="relative top-2 flex w-full">
            <!-- Search -->
            <div class="ml-2 w-1/3">
              <div class="flex items-center">
                <MagnifyingGlassIcon class="text-fg fixed ml-1.5 size-4.5"></MagnifyingGlassIcon>
                <input
                  v-model="searchQueries[selectedEl]"
                  :id="`${selectedEl}-search-1`"
                  :placeholder="`Search ${selectedEl}s...`"
                  type="text"
                  class="input pl-7.5"
                />
              </div>
            </div>
            <LeBtn @click="isOpen = false" class="absolute right-4">Close</LeBtn>
          </div>
          <div class="w-full px-2 py-12 sm:px-0">
            <TabGroup>
              <TabList class="relative flex w-full items-center justify-around place-self-center">
                <!-- Slider-->
                <span
                  class="dark:bg-sec bg-sec absolute -bottom-1 z-10 h-[130%] w-[48%] transform rounded-md transition-transform duration-300 ease-in-out"
                  :class="sliderTransform"
                />
                <!-- Tabs -->
                <Tab
                  class="z-20 w-1/2 cursor-pointer focus:outline-none"
                  @click="moveSlider('style')"
                  :class="
                    selectedEl === 'style'
                      ? 'text-fg dark:text-acc transition duration-500 ease-out'
                      : 'text-fg hover:text-acc hover:cursor-pointer'
                  "
                >
                  Styles
                </Tab>
                <Tab
                  class="z-20 w-1/2 cursor-pointer focus:outline-none"
                  @click="moveSlider('sample')"
                  :class="
                    selectedEl === 'sample'
                      ? 'text-fg dark:text-acc transition duration-500 ease-out'
                      : 'text-fg hover:text-acc hover:cursor-pointer'
                  "
                >
                  Samples
                </Tab>
              </TabList>
              <TabPanels class="p-2">
                <LeFlexTh class="rounded-t-md">
                  <div scope="col" class="text-fg w-[50%] py-3 pl-2">Name</div>
                  <div scope="col" class="text-fg w-[20%] py-3">Price</div>
                  <div scope="col" class="text-fg w-[10%] py-3">Time</div>
                  <div scope="col" class="text-fg w-[20%] py-3 text-center">Action</div>
                </LeFlexTh>
                <div class="bg-sec disabled: flex w-full items-center gap-4 pt-1 pb-2 pl-2">
                  <LeInput
                    label=""
                    :id="`new-${selectedEl}-name-1`"
                    v-model="newItemForm.name"
                    :placeholder="selectedEl + ' name'"
                    class-names="w-[50%]"
                  ></LeInput>
                  <LeInput
                    label=""
                    :id="`new-${selectedEl}-price-1`"
                    v-model="newItemForm.price"
                    class-names="w-[20%]"
                    placeholder="price"
                  ></LeInput>
                  <LeInput
                    label=""
                    :id="`new-${selectedEl}-time-1`"
                    v-model="newItemForm.time"
                    class-names="w-[15%]"
                    :placeholder="selectedEl === 'style' ? 'N/A' : 'minutes'"
                    :title="selectedEl === 'style' ? 'Styles do not contain time' : ''"
                    :disabled="selectedEl === 'style'"
                    :class="
                      selectedEl === 'style'
                        ? 'disabled:hover:border-warn disabled:cursor-not-allowed'
                        : ''
                    "
                  ></LeInput>
                  <div class="flex w-[15%] justify-center pr-2">
                    <LeBtn
                      @click="addNewItem(selectedEl, newItemForm)"
                      buttonText=""
                      class="flex"
                      :title="`add new ${selectedEl}`"
                    >
                      Add
                    </LeBtn>
                  </div>
                </div>

                <TabPanel
                  class="border-brdr max-h-[70vh] overflow-y-auto rounded-b border-x border-b"
                >
                  <LeFlexTd
                    v-for="(item, index) in filteredItems"
                    :key="index"
                    class="flex w-full items-center"
                  >
                    <!-- Editable Row -->
                    <template v-if="editing && editedRowId === item.id">
                      <LeInput
                        :id="`edit-sample-name-${item.id}`"
                        :placeholder="item.name"
                        v-model="editForm.name"
                        class-names="w-[50%] px-1"
                      />
                      <LeInput
                        :id="`edit-sample-company-${item.id}`"
                        :placeholder="item.price"
                        v-model="editForm.price"
                        class-names="w-[20%] px-1"
                      />
                      <LeInput
                        :id="`edit-sample-email-${item.id}`"
                        :placeholder="item.time"
                        v-model="editForm.time"
                        class-names="w-[10%] px-1"
                      />
                      <div
                        class="flex w-[20%] items-center justify-center gap-5 py-4 pl-4 text-center"
                      >
                        <a href="#" class="text-success hover:brightness-110" @click="editSave()"
                          ><CheckCircleIcon class="size-5"></CheckCircleIcon
                        ></a>
                        <a href="#" class="text-danger hover:brightness-110" @click="cancelEdit()"
                          ><XCircleIcon class="size-5"></XCircleIcon
                        ></a>
                      </div>
                    </template>
                    <!-- Read-Only Row -->
                    <template v-else>
                      <!-- name -->
                      <div class="line-clamp-1 w-[50%] pl-2">{{ item.name }}</div>
                      <!-- time -->
                      <div class="w-[20%] pl-2">{{ item.price }}</div>
                      <!-- price -->
                      <div class="w-[10%] pl-2">{{ item.price }}</div>
                      <!-- action -->
                      <div class="w-[20%]">
                        <!-- Controls pading for entire read only row -->
                        <div
                          class="flex w-full items-center justify-center gap-5 py-3 pl-4 text-center"
                        >
                          <a href="#" @click="editTrue(item)">
                            <PencilIcon class="hover:text-success size-5"></PencilIcon>
                          </a>
                          <a href="#" class="hover:text-danger" @click="removeItem(item)">
                            <TrashIcon class="size-5"></TrashIcon>
                          </a>
                        </div>
                      </div>
                    </template>
                  </LeFlexTd>
                </TabPanel>
                <TabPanel
                  class="border-brdr max-h-[70vh] overflow-y-auto rounded-b border-x border-b"
                >
                  <LeFlexTd
                    v-for="(item, index) in filteredItems"
                    :key="index"
                    class="flex w-full items-center"
                  >
                    <!-- Editable Row -->
                    <template v-if="editing && editedRowId === item.id">
                      <LeInput
                        :id="`edit-sample-name-${item.id}`"
                        :placeholder="item.name"
                        v-model="editForm.name"
                        class-names="w-[50%] px-1"
                      />
                      <LeInput
                        :id="`edit-sample-company-${item.id}`"
                        :placeholder="item.price"
                        v-model="editForm.price"
                        class-names="w-[20%] px-1"
                      />
                      <LeInput
                        :id="`edit-sample-email-${item.id}`"
                        :placeholder="item.time"
                        v-model="editForm.time"
                        class-names="w-[10%] px-1"
                      />
                      <div
                        class="flex w-[20%] items-center justify-center gap-5 py-4 pl-4 text-center"
                      >
                        <a href="#" class="text-success hover:brightness-110" @click="editSave()"
                          ><CheckCircleIcon class="size-5"></CheckCircleIcon
                        ></a>
                        <a href="#" class="text-danger hover:brightness-110" @click="cancelEdit()"
                          ><XCircleIcon class="size-5"></XCircleIcon
                        ></a>
                      </div>
                    </template>
                    <!-- Read-Only Row -->
                    <template v-else>
                      <!-- name -->
                      <div class="line-clamp-1 w-[50%] pl-2">{{ item.name }}</div>
                      <!-- time -->
                      <div class="w-[20%] pl-2">{{ item.price }}</div>
                      <!-- price -->
                      <div class="w-[10%] pl-2">{{ item.price }}</div>
                      <!-- action -->
                      <div class="w-[20%]">
                        <!-- Controls pading for entire read only row -->
                        <div
                          class="flex w-full items-center justify-center gap-5 py-3 pl-4 text-center"
                        >
                          <a href="#" @click="editTrue(item)">
                            <PencilIcon class="hover:text-success size-5"></PencilIcon>
                          </a>
                          <a href="#" class="hover:text-danger" @click="removeItem(item)">
                            <TrashIcon class="size-5"></TrashIcon>
                          </a>
                        </div>
                      </div>
                    </template>
                  </LeFlexTd>
                </TabPanel>
              </TabPanels>
            </TabGroup>
          </div>
        </div>
      </div>
    </transition>
  </Teleport>
</template>
<style scoped>
.fadeItIn-enter-active,
.fadeItIn-leave-active {
  transition: opacity 0.3s ease;
}

.fadeItIn-enter-from,
.fadeItIn-leave-to {
  opacity: 0;
}
.slideItIn-enter-to,
.slideItIn-leave-from {
  opacity: 1;
}

.slideItIn-enter-active,
.slideItIn-leave-active {
  transition: transform 0.5s ease;
}
.slideItIn-enter-from,
.slideItIn-leave-to {
  transform: translateX(100%);
}
.slideItIn-enter-to,
.slideItIn-leave-from {
  transform: translateX(0);
}
</style>
