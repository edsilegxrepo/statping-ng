<template>
  <div>
    <h3>Configuration</h3>
    For security reasons, all database credentials cannot be edited from this page.

    <textarea v-if="loaded" v-model="configs" class="form-control code-editor mt-4" rows="20"></textarea>

    <button @click.prevent="save" class="btn col-12 btn-primary mt-3">Save</button>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Api from '@/API'

const loaded = ref(false)
const configs = ref(null)

onMounted(() => {
  update()
})

async function update() {
  loaded.value = false
  configs.value = await Api.configs()
  loaded.value = true
}

async function save() {
  try {
    await Api.configs_save(configs.value)
  } catch (e) {
    console.error(e)
  }
}
</script>

<style scoped>
.code-editor {
  font-family: monospace;
  font-size: 12px;
}
</style>
