<template>
  <div>
    <TopNav />
    <PageHeader />
    <div class="offset-md-3 offset-lg-4 offset-0 col-lg-4 col-md-6 mt-4">
      <div class="login_container col-12 p-4">
        <FormLogin />
      </div>
    </div>
  </div>
</template>

<script setup>
import FormLogin from '@/forms/Login.vue'
import TopNav from '@/components/Index/TopNav.vue'
import PageHeader from '@/components/Elements/PageHeader.vue'
</script>

<style scoped></style>
