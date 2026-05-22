<template>
    <footer>
        <div v-if="!core.footer" class="footer text-center mt-5 mb-5 pt-3 pb-4 footer-border">
          <div class="d-block text-muted">
          <div class="mb-3">
            <router-link class="links font-weight-bold" to="/">Monitors</router-link>
            <span class="mx-3 text-primary font-weight-bold separator">|</span>
            <router-link class="links font-weight-bold" :to="admin ? '/dashboard' : '/login'">{{$t('dashboard')}}</router-link>
          </div>
            <span class="font-1 mt-3 text-dim">
              Statping {{core.version}}
            </span>
          </div>
        </div>
        <div v-else class="footer text-center mb-4 p-2" v-html="core.footer"></div>
    </footer>
</template>

<script>
  import Dashboard from "../../pages/Dashboard";

  export default {
  name: 'Footer',
  components: {
    Dashboard
  },
  computed: {
      core() {
        return this.$store.getters.core
      },
    commit() {
      return this.$store.getters.core.commit.slice(0,8)
    },
    admin() {
      return this.$store.getters.admin
    },
  }
}
</script>

<style scoped>
.footer-border {
    border-top: 2px solid #dee2e6;
    border-top-color: rgba(0, 123, 255, 0.2);
}
.separator {
    font-size: 1.2rem;
    display: inline-block;
    transform: translateY(-2px);
    vertical-align: middle;
}
.hlight {
  color: #f6cbcb;
}
</style>
