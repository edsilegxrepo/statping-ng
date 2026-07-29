<template>
  <div id="app">
    <router-view />
    <Footer v-if="$route.path !== '/setup'" />
  </div>
</template>

<script setup>
import { ref, computed, onBeforeMount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useMainStore } from './stores/main'
import Footer from './components/Index/Footer.vue'

const router = useRouter()
const route = useRoute()
const store = useMainStore()
const { locale } = useI18n()

const loaded = ref(false)

const core = computed(() => store.core)

onBeforeMount(async () => {
  await store.loadCore()

  locale.value = core.value.language || 'en'

  if (!core.value.setup) {
    router.push('/setup')
  }

  if (route.path !== '/setup') {
    if (store.admin) {
      await store.loadAdmin()
    } else {
      await store.loadRequired()
    }
    loaded.value = true
  }
})
</script>

<style lang="scss">
@use './assets/css/bootstrap.min.css';
@use './assets/scss/index';
</style>
