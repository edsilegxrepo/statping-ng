<template>
  <div>
    <TopNav />
    <div class="offset-md-3 offset-lg-4 offset-0 col-lg-4 col-md-6 mt-5">
        <PageHeader />
        <div class="col-12 mt-5 mb-5 text-center">
            <div class="mb-4">
                <font-awesome-icon icon="exclamation-triangle" class="text-muted" size="4x"/>
            </div>
            <h2 class="font-weight-bold">404 - Page Not Found</h2>
            <h5 class="text-muted">The requested resource could not be found.</h5>
            <router-link class="btn btn-primary px-5 shadow-sm mt-4" to="/">Back To Homepage</router-link>
        </div>
    </div>
  </div>
</template>

<script setup>
import PageHeader from "@/components/Elements/PageHeader.vue";
import TopNav from "@/components/Index/TopNav.vue";
</script>

<style scoped>
</style>
