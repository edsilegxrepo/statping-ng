<template>
  <div class="card shadow mb-4" role="alert">
    <div class="card-body pb-2">
      <h3 class="mb-3 font-weight-bold">{{ message.title }}</h3>
      <span class="mb-2">{{ message.description }}</span>
      <div class="col-12 mb-0">
        <div class="dates">
          <div class="start">
            <strong>STARTS</strong> {{ niceDate(message.start_on) }}
            <span></span>
          </div>
          <div class="ends"><strong>ENDS</strong> {{ niceDate(message.end_on) }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  message: {
    type: Object,
    required: true,
  },
})

function niceDate(date) {
  return new Date(date).toLocaleDateString()
}
</script>

<style scoped></style>
