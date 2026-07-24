import Vue from "vue";
import VueApexCharts from "vue-apexcharts";
import VueClipboard from "vue-clipboard2";
import VueCookies from "vue-cookies";
import VueI18n from "vue-i18n";
import VueObserveVisibility from "vue-observe-visibility";
import VueRouter from "vue-router";
import router from "./routes";
import "./mixin";
import "./icons";
import language from "./languages";
import store from "./store";

const App = () => import(/* webpackChunkName: "index" */ "@/App.vue");

Vue.component("apexchart", VueApexCharts);

Vue.use(VueClipboard);
Vue.use(VueRouter);
Vue.use(VueObserveVisibility);
Vue.use(VueCookies);
Vue.use(VueI18n);

const i18n = new VueI18n({
	fallbackLocale: "en",
	messages: language,
});

Vue.$cookies.config("3d");

Vue.config.productionTip = process.env.NODE_ENV !== "production";
Vue.config.devtools = process.env.NODE_ENV !== "production";
Vue.config.performance = process.env.NODE_ENV !== "production";

new Vue({
	router,
	store,
	i18n,
	render: (h) => h(App),
}).$mount("#app");
