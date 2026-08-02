import { createRouter, createWebHistory } from 'vue-router'
import Api from '../API'
import { useMainStore } from '../stores/main'

const Index = () => import('@/pages/Index.vue')
const Dashboard = () => import('@/pages/Dashboard.vue')
const DashboardIndex = () => import('@/components/Dashboard/DashboardIndex.vue')
const DashboardUsers = () => import('@/components/Dashboard/DashboardUsers.vue')
const DashboardServices = () => import('@/components/Dashboard/DashboardServices.vue')
const DashboardMessages = () => import('@/components/Dashboard/DashboardMessages.vue')
const EditService = () => import('@/components/Dashboard/EditService.vue')
const Logs = () => import('@/pages/Logs.vue')
const Help = () => import('@/pages/Help.vue')
const AdminHelp = () => import('@/components/Dashboard/AdminHelp.vue')
const Settings = () => import('@/pages/Settings.vue')
const Login = () => import('@/pages/Login.vue')
const Service = () => import('@/pages/Service.vue')
const Setup = () => import('@/forms/Setup.vue')
const Incidents = () => import('@/components/Dashboard/Incidents.vue')
const Checkins = () => import('@/components/Dashboard/Checkins.vue')
const Failures = () => import('@/components/Dashboard/Failures.vue')
const NotFound = () => import('@/pages/NotFound.vue')
const Importer = () => import('@/components/Dashboard/Importer.vue')

const routes = [
  {
    path: '/setup',
    name: 'Setup',
    component: Setup,
    meta: {
      title: 'Statping Setup',
    },
    beforeEnter: async (_to, _from, next) => {
      try {
        const core = await Api.core()
        if (core.setup) {
          next('/')
        } else {
          next()
        }
      } catch {
        next()
      }
    },
  },
  {
    path: '/',
    name: 'Index',
    component: Index,
  },
  {
    path: '/dashboard',
    component: Dashboard,
    meta: {
      requiresAuth: true,
      title: 'Statping - Dashboard',
    },
    children: [
      {
        path: '',
        component: DashboardIndex,
        meta: {
          requiresAuth: true,
          title: 'Statping - Dashboard',
        },
      },
      {
        path: 'users',
        component: DashboardUsers,
        meta: {
          requiresAuth: true,
          title: 'Statping - Users',
        },
      },
      {
        path: 'services',
        component: DashboardServices,
        meta: {
          requiresAuth: true,
          title: 'Statping - Services',
        },
      },
      {
        path: 'create_service',
        component: EditService,
        meta: {
          requiresAuth: true,
          title: 'Statping - Create Service',
        },
      },
      {
        path: 'edit_service/:id',
        component: EditService,
        meta: {
          requiresAuth: true,
          title: 'Statping - Edit Service',
        },
      },
      {
        path: 'service/:id/incidents',
        component: Incidents,
        meta: {
          requiresAuth: true,
          title: 'Statping - Incidents',
        },
      },
      {
        path: 'service/:id/checkins',
        component: Checkins,
        meta: {
          requiresAuth: true,
          title: 'Statping - Checkins',
        },
      },
      {
        path: 'service/:id/failures',
        component: Failures,
        meta: {
          requiresAuth: true,
          title: 'Statping - Service Failures',
        },
      },
      {
        path: 'messages',
        component: DashboardMessages,
        meta: {
          requiresAuth: true,
          title: 'Statping - Messages',
        },
      },
      {
        path: 'settings',
        component: Settings,
        meta: {
          requiresAuth: true,
          title: 'Statping - Settings',
        },
      },
      {
        path: 'logs',
        component: Logs,
        meta: {
          requiresAuth: true,
          title: 'Statping - Logs',
        },
      },
      {
        path: 'help',
        component: AdminHelp,
        meta: {
          requiresAuth: true,
          title: 'Statping - Admin Help',
        },
      },
      {
        path: 'import',
        component: Importer,
        meta: {
          requiresAuth: true,
          title: 'Statping - Import',
        },
      },
    ],
  },
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: {
      title: 'Statping - Login',
    },
  },
  {
    path: '/help',
    name: 'Help',
    component: Help,
    meta: {
      title: 'Statping - Help',
    },
  },
  { path: '/logout', redirect: '/' },
  {
    path: '/service/:id',
    name: 'Service',
    component: Service,
    props: true,
  },
  {
    path: '/:pathMatch(.*)*',
    component: NotFound,
    name: 'NotFound',
  },
]

const router = createRouter({
  history: createWebHistory(),
  scrollBehavior(_to, _from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    } else {
      return { top: 0, left: 0 }
    }
  },
  routes,
})

router.beforeEach(async (to, _from, next) => {
  const nearestWithTitle = to.matched
    .slice()
    .reverse()
    .find((r) => r.meta?.title)

  if (nearestWithTitle) {
    document.title = nearestWithTitle.meta.title
  }

  if (to.matched.some((record) => record.meta.requiresAuth)) {
    const store = useMainStore()

    if (store.loggedIn) {
      next()
      return
    }

    try {
      const jwt = await Api.check_token()
      if (jwt?.admin) {
        store.setAdmin(true)
        store.setLoggedIn(true)
        store.setUser(true)
        next()
      } else {
        store.setLoggedIn(false)
        next('/login')
      }
    } catch (e) {
      if (import.meta.env.DEV) console.error('Auth check failed:', e)
      store.setLoggedIn(false)
      next('/login')
    }
  } else {
    next()
  }
})

export default router
