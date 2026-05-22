<template>
    <nav class="navbar navbar-expand-lg">
        <router-link to="/" class="navbar-brand font-weight-bold">Monitors</router-link>
        <span class="d-none d-lg-inline text-primary ml-1 mr-2" style="font-size: 1.5rem; opacity: 0.8; font-weight: 600;">|</span>
        <button @click="navopen = !navopen" class="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarText" aria-controls="navbarText" aria-expanded="false" aria-label="Toggle navigation">
            <font-awesome-icon v-if="!navopen" icon="bars"/>
            <font-awesome-icon v-if="navopen" icon="times"/>
        </button>

        <div class="navbar-collapse" :class="{collapse: !navopen}" id="navbarText">
            <ul class="navbar-nav mr-auto">
                <li @click="navopen = !navopen" class="nav-item navbar-item">
                    <router-link to="/dashboard" class="nav-link" exact>{{ $t('dashboard') }}</router-link>
                </li>
                <li @click="navopen = !navopen" class="nav-item navbar-item">
                    <router-link to="/dashboard/services" class="nav-link">{{ $t('services') }}</router-link>
                </li>
                <li v-if="admin" @click="navopen = !navopen" class="nav-item navbar-item">
                    <router-link to="/dashboard/users" class="nav-link">{{ $t('users') }}</router-link>
                </li>
                <li @click="navopen = !navopen" class="nav-item navbar-item">
                    <router-link to="/dashboard/messages" class="nav-link">{{ $t('announcements') }}</router-link>
                </li>
                <li v-if="admin" @click="navopen = !navopen" class="nav-item navbar-item">
                    <router-link to="/dashboard/settings" class="nav-link">{{ $t('settings') }}</router-link>
                </li>
              <li v-if="admin" @click="navopen = !navopen" class="nav-item navbar-item">
                <router-link to="/dashboard/logs" class="nav-link">{{ $t('logs') }}</router-link>
              </li>
              <li v-if="admin" @click="navopen = !navopen" class="nav-item navbar-item">
                <router-link to="/dashboard/help" class="nav-link">{{ $t('help') }}</router-link>
              </li>
            </ul>
            <ul class="navbar-nav">
                <li @click="navopen = !navopen" class="nav-item navbar-item">
                    <a href="#" class="nav-link" @click.prevent="logout">{{ $t('logout') }}</a>
                </li>
            </ul>
        </div>
    </nav>

</template>

<script>
  import Api from "../../API"

  export default {
  name: 'TopNav',
      data () {
          return {
              navopen: false
          }
      },
    computed: {
      admin() {
        return this.$store.state.admin
      }
    },
      methods: {
        async logout () {
          try {
            await Api.logout()
          } catch (e) {
            console.error("Backend logout failed", e)
          }
          this.$store.commit('setHasAllData', false)
          this.$store.commit('setToken', null)
          this.$store.commit('setAdmin', false)
          this.$store.commit('setUser', false)
          this.$store.commit('setLoggedIn', false)
          this.$cookies.remove("statping_auth")
          await this.$router.push('/logout')
        }
    }
}
</script>

<style scoped>
.navbar-item {
    position: relative;
    padding: 0 5px;
}

.nav-link {
    transition: color 0.2s ease-in-out;
}

.nav-link::after {
    content: '';
    position: absolute;
    bottom: 0;
    left: 0;
    width: 0;
    height: 2px;
    background-color: #007bff;
    transition: width 0.2s ease-in-out;
}

.nav-link:hover::after,
.router-link-exact-active::after,
.router-link-active::after {
    width: 100%;
}

.router-link-exact-active,
.router-link-active {
    color: #007bff !important;
    font-weight: 600;
}
</style>
