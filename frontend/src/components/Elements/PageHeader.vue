<template>
  <div class="col-12 mb-5 text-center">
    <h1 class="page-header-title">{{ core.name }}</h1>
    <p class="text-muted mt-2">{{ core.description }}</p>
    <hr class="page-header-hr" />
  </div>
</template>

<script setup>
import { useMainStore } from '@/stores/main'
import { storeToRefs } from 'pinia'

const store = useMainStore()
const { core } = storeToRefs(store)
</script>

<style scoped>
.page-header-title {
  font-weight: bold;
  color: #343a40;
  margin-bottom: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
  font-size: 2.2rem;
}

.page-header-hr {
  width: 100px;
  border-top: 3px solid #007bff;
  margin: 0.5rem auto 0;
}
</style>
