<template>
  <footer>
    <div v-if="!core.footer" class="footer text-center mt-5 mb-5 pt-3 pb-4 footer-border">
      <div class="d-block text-muted">
        <div class="mb-3">
          <router-link class="links font-weight-bold" to="/">Monitors</router-link>
          <span class="mx-3 text-primary font-weight-bold separator">|</span>
          <router-link class="links font-weight-bold" :to="admin ? '/dashboard' : '/login'">{{
            $t('dashboard')
          }}</router-link>
        </div>
        <span class="font-1 mt-3 text-dim"> Statping {{ core.version }} </span>
      </div>
    </div>
    <div v-else class="footer text-center mb-4 p-2" v-html="sanitizedFooter"></div>
  </footer>
</template>

<script setup>
import { computed } from 'vue'
import { useMainStore } from '@/stores/main'
import DOMPurify from 'dompurify'

const store = useMainStore()

const core = computed(() => store.core)
const admin = computed(() => store.admin)
const sanitizedFooter = computed(() => DOMPurify.sanitize(core.value.footer || ''))
</script>

<style scoped>
.footer-border {
  border-top: 2px solid #dee2e6;
  border-top-color: rgba(0, 123, 255, 0.2);
}
.separator {
  font-size: 1.2rem;
  display-inline: block;
  transform: translateY(-2px);
  vertical-align: middle;
}
.hlight {
  color: #f6cbcb;
}
</style>
