// composables/useFormHandler.js
import { reactive } from 'vue'

export function useFormHandler(initialState) {
  const form = reactive({ ...initialState, errors: {} })

  const updateField = (field, value) => {
    form[field] = value
    if (form.errors[field]) {
      delete form.errors[field] // Clear error on change
    }
  }

  const validate = (rules) => {
    form.errors = {}
    for (const field in rules) {
      if (!form[field]) {
        form.errors[field] = `${field} is required`
      }
    }
    return Object.keys(form.errors).length === 0
  }

  const reset = () => {
    Object.assign(form, { ...initialState, errors: {} })
  }

  return { form, updateField, validate, reset }
}
// use example
/**
 import useFormHandler from '...' 
 const {}
 */
