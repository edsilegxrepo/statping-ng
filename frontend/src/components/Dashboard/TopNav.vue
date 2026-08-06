<template>
  <nav class="dashboard-nav">
    <div class="dashboard-nav-content">
      <div class="dashboard-nav-left">
        <router-link to="/" class="site-name">{{ core.name || 'Statping' }}</router-link>
        <span class="nav-divider">|</span>
        <router-link to="/dashboard" class="dashboard-brand">
          <font-awesome-icon icon="chart-line" class="brand-icon" />
          <span>{{ $t('dashboard') }}</span>
        </router-link>
      </div>

      <div class="dashboard-nav-center">
        <router-link to="/dashboard/services" class="nav-item" active-class="active">
          <font-awesome-icon icon="server" class="nav-icon" />
          <span>{{ $t('services') }}</span>
        </router-link>
        <router-link v-if="admin" to="/dashboard/users" class="nav-item" active-class="active">
          <font-awesome-icon icon="users" class="nav-icon" />
          <span>{{ $t('users') }}</span>
        </router-link>
        <router-link to="/dashboard/messages" class="nav-item" active-class="active">
          <font-awesome-icon icon="bullhorn" class="nav-icon" />
          <span>{{ $t('announcements') }}</span>
        </router-link>
        <router-link v-if="admin" to="/dashboard/settings" class="nav-item" active-class="active">
          <font-awesome-icon icon="cog" class="nav-icon" />
          <span>{{ $t('settings') }}</span>
        </router-link>
        <router-link v-if="admin" to="/dashboard/logs" class="nav-item" active-class="active">
          <font-awesome-icon icon="file-alt" class="nav-icon" />
          <span>{{ $t('logs') }}</span>
        </router-link>
        <router-link v-if="admin" to="/dashboard/help" class="nav-item" active-class="active">
          <font-awesome-icon icon="question-circle" class="nav-icon" />
          <span>{{ $t('help') }}</span>
        </router-link>
      </div>

      <div class="dashboard-nav-right">
        <router-link to="/" class="nav-btn nav-btn-monitors">
          <font-awesome-icon icon="globe" class="me-1" />
          Status Page
        </router-link>
        <button class="nav-btn nav-btn-logout" @click="logout">
          <font-awesome-icon icon="sign-out-alt" class="me-1" />
          {{ $t('logout') }}
        </button>
      </div>

      <!-- Mobile toggle -->
      <button class="mobile-toggle" @click="navopen = !navopen">
        <font-awesome-icon :icon="navopen ? 'times' : 'bars'" />
      </button>
    </div>

    <!-- Mobile menu -->
    <div v-if="navopen" class="mobile-menu">
      <router-link to="/dashboard" class="mobile-item" @click="navopen = false">
        <font-awesome-icon icon="tachometer-alt" class="me-2" />{{ $t('dashboard') }}
      </router-link>
      <router-link to="/dashboard/services" class="mobile-item" @click="navopen = false">
        <font-awesome-icon icon="server" class="me-2" />{{ $t('services') }}
      </router-link>
      <router-link v-if="admin" to="/dashboard/users" class="mobile-item" @click="navopen = false">
        <font-awesome-icon icon="users" class="me-2" />{{ $t('users') }}
      </router-link>
      <router-link to="/dashboard/messages" class="mobile-item" @click="navopen = false">
        <font-awesome-icon icon="bullhorn" class="me-2" />{{ $t('announcements') }}
      </router-link>
      <router-link v-if="admin" to="/dashboard/settings" class="mobile-item" @click="navopen = false">
        <font-awesome-icon icon="cog" class="me-2" />{{ $t('settings') }}
      </router-link>
      <router-link v-if="admin" to="/dashboard/logs" class="mobile-item" @click="navopen = false">
        <font-awesome-icon icon="file-alt" class="me-2" />{{ $t('logs') }}
      </router-link>
      <router-link v-if="admin" to="/dashboard/help" class="mobile-item" @click="navopen = false">
        <font-awesome-icon icon="question-circle" class="me-2" />{{ $t('help') }}
      </router-link>
      <div class="mobile-divider"></div>
      <router-link to="/" class="mobile-item" @click="navopen = false">
        <font-awesome-icon icon="globe" class="me-2" />Status Page
      </router-link>
      <a href="#" class="mobile-item mobile-item-logout" @click.prevent="logout">
        <font-awesome-icon icon="sign-out-alt" class="me-2" />{{ $t('logout') }}
      </a>
    </div>
  </nav>
</template>

<script setup>
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import { useCookies } from "vue3-cookies";
import Api from "@/API";
import { useMainStore } from "@/stores/main";

defineProps({
	admin: {
		type: Boolean,
		default: false,
	},
});

const router = useRouter();
const store = useMainStore();
const { cookies } = useCookies();

const navopen = ref(false);
const core = computed(() => store.core);

async function logout() {
	let redirectUrl = null;
	try {
		const response = await Api.logout();
		if (response.redirect) {
			redirectUrl = response.redirect;
		}
	} catch (e) {
		console.error("Backend logout failed", e);
	}
	store.hasAllData = false;
	store.token = null;
	store.admin = false;
	store.user = false;
	store.loggedIn = false;
	cookies.remove("statping_auth");

	if (redirectUrl) {
		window.location.href = redirectUrl;
	} else {
		await router.push("/logout");
	}
}
</script>

<style scoped>
.dashboard-nav {
  position: sticky;
  top: 0;
  z-index: 1000;
  background: linear-gradient(135deg, var(--color-dark-bg) 0%, var(--color-dark-bg-secondary) 100%);
  box-shadow: var(--shadow-lg);
}

.dashboard-nav-content {
  max-width: 1400px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  padding: 0 var(--space-4);
}

.dashboard-nav-left {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.site-name {
  color: #fff;
  font-weight: 600;
  font-size: 1rem;
  text-decoration: none;
  max-width: 200px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.site-name:hover {
  color: #fff;
  opacity: 0.9;
}

.nav-divider {
  color: var(--color-primary-light);
  font-size: 1.25rem;
  font-weight: 600;
  opacity: 0.6;
}

.dashboard-brand {
  display: flex;
  align-items: center;
  text-decoration: none;
  color: var(--color-dark-text-muted);
  font-weight: 500;
  font-size: 0.95rem;
  transition: color var(--transition-fast);
}

.dashboard-brand:hover {
  color: #fff;
}

.brand-icon {
  margin-right: var(--space-2);
  color: var(--color-primary-light);
}

.dashboard-nav-center {
  display: flex;
  align-items: center;
  gap: var(--space-1);
}

.nav-item {
  display: flex;
  align-items: center;
  padding: var(--space-2) var(--space-3);
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-dark-text-muted);
  text-decoration: none;
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

.nav-item:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
}

.nav-item.active {
  color: #fff;
  background: rgba(255, 255, 255, 0.15);
}

.nav-icon {
  margin-right: var(--space-2);
  font-size: 0.875rem;
}

.dashboard-nav-right {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.nav-btn {
  display: flex;
  align-items: center;
  padding: var(--space-2) var(--space-3);
  font-size: 0.875rem;
  font-weight: 500;
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-md);
  text-decoration: none;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.nav-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  border-color: rgba(255, 255, 255, 0.2);
}

.nav-btn-monitors {
  background: var(--color-primary-bg);
  border-color: rgba(59, 130, 246, 0.3);
  color: var(--color-primary-light);
}

.nav-btn-monitors:hover {
  background: rgba(59, 130, 246, 0.2);
}

.nav-btn-logout {
  background: var(--color-danger-bg);
  border-color: rgba(239, 68, 68, 0.3);
  color: var(--color-danger-light);
}

.nav-btn-logout:hover {
  background: rgba(239, 68, 68, 0.2);
}

.mobile-toggle {
  display: none;
  padding: var(--space-2);
  background: transparent;
  border: none;
  color: #fff;
  font-size: 1.25rem;
  cursor: pointer;
}

.mobile-menu {
  display: none;
  padding: var(--space-2);
  background: var(--color-dark-bg-secondary);
  border-top: 1px solid var(--color-dark-border);
}

.mobile-item {
  display: block;
  padding: var(--space-3) var(--space-4);
  color: var(--color-dark-text-muted);
  text-decoration: none;
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

.mobile-item:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
}

.mobile-item-logout {
  color: var(--color-danger-light);
}

.mobile-divider {
  height: 1px;
  background: var(--color-dark-border);
  margin: var(--space-2) 0;
}

/* Responsive */
@media (max-width: 1024px) {
  .dashboard-nav-center {
    display: none;
  }

  .dashboard-nav-right {
    display: none;
  }

  .mobile-toggle {
    display: block;
  }

  .mobile-menu {
    display: block;
  }
}

@media (max-width: 576px) {
  .brand-text {
    display: none;
  }
}
</style>
