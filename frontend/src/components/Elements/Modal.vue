<template>
  <div v-if="modal.visible" class="modal d-block mt-5 pt-5" tabindex="-1">
    <div class="modal-dialog">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">{{ modal.title }}</h5>
        </div>
        <div class="modal-body">
          <p class="modal-body-text">{{ modal.body }}</p>
        </div>
        <div class="modal-footer">
          <button v-if="!modal.hideCancel" @click.prevent="close" type="button" class="btn btn-secondary" data-dismiss="modal">Close</button>
          <button @click.prevent="runFunc" type="button" :class="`btn ${modal.btnColor}`">{{ modal.btnText }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useMainStore } from '@/stores/main'

const store = useMainStore()

const modal = computed(() => store.modal)

function runFunc() {
  if (store.modal.func) {
    store.modal.func()
  }
  close()
}

function close() {
  store.setModal({ visible: false })
}
</script>

<style scoped>
.modal-body-text {
  white-space: pre-line;
}
</style>
