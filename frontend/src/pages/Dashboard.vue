<template>
  <div class="dashboard-wrapper">
    <div v-if="modal" class="modal-backdrop"></div>
    <Modal />
    <TopNav :admin="admin" />
    <div class="dashboard-body">
      <router-view :admin="admin" />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useMainStore } from '@/stores/main'
import Modal from '@/components/Elements/Modal.vue'
import TopNav from '@/components/Dashboard/TopNav.vue'

const store = useMainStore()

const modal = computed(() => store.modal.visible)
const admin = computed(() => store.admin)
</script>

<style scoped>
.dashboard-wrapper {
  min-height: 100vh;
}

.dashboard-body {
  padding: 1.5rem 3rem;
}
</style>
