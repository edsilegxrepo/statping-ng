<template>
    <div id="header">
        <h1 id="title" class="col-12 text-center pt-4 mt-4 mb-3 header-title font-6">{{core.name}}</h1>
        <h5 id="description" class="col-12 text-center mb-0 header-desc font-3 text-muted">{{core.description}}</h5>
        <div class="col-12 text-center mt-4 mb-5">
            <hr style="width: 120px; border-top: 4px solid #007bff; margin: 0 auto; border-radius: 2px;">
        </div>
    </div>
</template>

<script>
export default {
  name: 'Header',
    computed: {
      core() {
          return this.$store.getters.core
      }
    }
}
</script>
