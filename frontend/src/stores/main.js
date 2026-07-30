import { defineStore } from 'pinia'
import Api from '../API'

export const useMainStore = defineStore('main', {
  state: () => ({
    hasAllData: false,
    hasPublicData: false,
    core: {},
    oauth: {},
    token: null,
    services: [],
    service: null,
    groups: [],
    messages: [],
    users: [],
    notifiers: [],
    checkins: [],
    admin: false,
    user: false,
    loggedIn: false,
    modal: {
      visible: false,
      title: 'Modal Header',
      body: 'This is the content for the modal body',
      btnText: 'Save Changes',
      btnColor: 'btn-primary',
      func: null,
    },
  }),

  getters: {
    isAdmin: (state) => state.admin,
    isUser: (state) => state.user,

    globalMessages: (state) =>
      state.messages.filter((s) => !s.service || s.service === 0),

    servicesInOrder: (state) =>
      [...state.services].sort((a, b) => a.order_id - b.order_id),

    servicesNoGroup: (state) =>
      state.services
        .filter((g) => g.group_id === 0)
        .sort((a, b) => a.order_id - b.order_id),

    groupsInOrder: (state) =>
      [...state.groups].sort((a, b) => a.order_id - b.order_id),

    groupsClean: (state) =>
      state.groups
        .filter((g) => g.name !== '')
        .sort((a, b) => a.order_id - b.order_id),

    groupsCleanInOrder: (state) =>
      state.groups
        .filter((g) => g.name !== '')
        .sort((a, b) => a.order_id - b.order_id),

    serviceCheckins: (state) => (id) => {
      return state.checkins.filter((c) => c.service_id === id)
    },

    serviceByAll: (state) => (element) => {
      const num = parseInt(element, 10)
      if (!Number.isNaN(num)) {
        return state.services.find((s) => s.id === num)
      }
      return state.services.find((s) => s.permalink === element)
    },

    serviceById: (state) => (id) => {
      return state.services.find((s) => s.permalink === id || s.id === id)
    },

    servicesInGroup: (state) => (id) => {
      return state.services
        .filter((s) => s.group_id === id)
        .sort((a, b) => a.order_id - b.order_id)
    },

    serviceMessages: (state) => (id) => {
      return state.messages.filter((s) => s.service === id)
    },

    onlineServices: (state) => (online) => {
      return state.services.filter((s) => s.online === online)
    },

    groupById: (state) => (id) => {
      return state.groups.find((g) => g.id === id)
    },

    cleanGroups: (state) => () => {
      return state.groups
        .filter((g) => g.name !== '')
        .sort((a, b) => a.order_id - b.order_id)
    },

    userById: (state) => (id) => {
      return state.users.find((u) => u.id === id)
    },

    messageById: (state) => (id) => {
      return state.messages.find((m) => m.id === id)
    },
  },

  actions: {
    setCore(core) {
      this.core = core
    },
    setToken(token) {
      this.token = token
    },
    setService(service) {
      this.service = service
    },
    setServices(services) {
      this.services = services
    },
    setCheckins(checkins) {
      this.checkins = checkins
    },
    setGroups(groups) {
      this.groups = groups
    },
    setMessages(messages) {
      this.messages = messages
    },
    setUsers(users) {
      this.users = users
    },
    setNotifiers(notifiers) {
      this.notifiers = notifiers
    },
    setAdmin(admin) {
      this.admin = admin
    },
    setLoggedIn(loggedIn) {
      this.loggedIn = loggedIn
    },
    setUser(user) {
      this.user = user
    },
    setOAuth(oauth) {
      this.oauth = oauth
    },
    setModal(modal) {
      this.modal = modal
    },

    async getAllServices() {
      const services = await Api.services()
      this.services = services
    },

    async loadCore() {
      const [core, token] = await Promise.all([Api.core(), Api.token()])
      this.core = core
      this.admin = token
      this.user = token !== undefined
    },

    async loadRequired() {
      const [groups, services, messages, oauth] = await Promise.all([
        Api.groups(),
        Api.services(),
        Api.messages(),
        Api.oauth(),
      ])
      this.groups = groups
      this.services = services
      this.messages = messages
      this.oauth = oauth
      this.hasPublicData = true
    },

    async loadAdmin() {
      const [groups, services, messages, checkins, notifiers, users, oauth] =
        await Promise.all([
          Api.groups(),
          Api.services(),
          Api.messages(),
          Api.checkins(),
          Api.notifiers(),
          Api.users(),
          Api.oauth(),
        ])
      this.groups = groups
      this.services = services
      this.messages = messages
      this.hasPublicData = true
      this.checkins = checkins
      this.notifiers = notifiers
      this.users = users
      this.oauth = oauth
    },
  },
})
