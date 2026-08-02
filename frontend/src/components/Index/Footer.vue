<template>
  <footer>
    <div v-if="!core.footer" class="footer-bar">
      <div class="footer-content">
        <div class="footer-left">
          <span v-if="core.environment" :class="['env-badge', 'env-' + core.environment.toLowerCase()]">{{ core.environment }}</span>
          <span class="version-text">Statping-ng {{ core.version }}<template v-if="core.commit">-{{ core.commit }}</template></span>
        </div>
        <div class="footer-right">
          <router-link class="footer-link" to="/">Monitors</router-link>
          <span class="footer-sep">|</span>
          <router-link class="footer-link" to="/help">Help</router-link>
          <span class="footer-sep">|</span>
          <router-link class="footer-link" :to="admin ? '/dashboard' : '/login'">{{ $t('dashboard') }}</router-link>
        </div>
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
.footer-bar {
  background: #2d3748;
  border-top: 1px solid #4a5568;
  padding: 0.5rem 1.5rem;
  margin-top: 3rem;
}
.footer-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  max-width: 1400px;
  margin: 0 auto;
}
.footer-left, .footer-right {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}
.version-text {
  color: #a0aec0;
  font-size: 0.8rem;
  font-family: 'Monaco', 'Menlo', monospace;
}
.footer-link {
  color: #a0aec0;
  font-size: 0.8rem;
  text-decoration: none;
  transition: color 0.15s;
}
.footer-link:hover {
  color: #fff;
}
.footer-sep {
  color: #4a5568;
  font-size: 0.8rem;
}
.env-badge {
  font-size: 0.65rem;
  font-weight: 700;
  padding: 0.2rem 0.5rem;
  border-radius: 3px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.env-prod, .env-production {
  background-color: #e53e3e;
  color: white;
}
.env-qa, .env-uat, .env-staging {
  background-color: #d69e2e;
  color: #1a202c;
}
.env-dev, .env-development {
  background-color: #38a169;
  color: white;
}
.env-test {
  background-color: #3182ce;
  color: white;
}
</style>
