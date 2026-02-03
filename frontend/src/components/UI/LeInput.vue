<template>
  <div class="relative" :class="props.classNames">
    <label v-if="!labelHidden" :for="props.id" class="input-label">
      {{ props.label }}
    </label>
    <input
      v-bind="$attrs"
      :id="id"
      :name="name"
      :value="modelValue"
      @input="emit('update:modelValue', $event.target.value)"
      class="input"
      :type="type"
      @focus="$event.target.select()"
    />
  </div>
</template>

<script setup>
const props = defineProps({
  label: String,
  id: String,
  name: String,
  modelValue: [String, Number],
  labelHidden: Boolean,
  classNames: String,
  type: String,
})

const emit = defineEmits(['update:modelValue'])
</script>
