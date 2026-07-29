<template>
  <div class="container-fluid px-md-5 mt-md-5">
    <div v-if="modal" class="modal-backdrop"></div>
    <Modal />
    <TopNav :admin="admin" />
    <router-view :admin="admin" />
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
