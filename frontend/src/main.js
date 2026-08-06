import { createPinia } from "pinia";
import { createApp } from "vue";
import { createI18n } from "vue-i18n";
import VueApexCharts from "vue3-apexcharts";
import VueCookies from "vue3-cookies";

import App from "./App.vue";
import {
	FontAwesomeIcon,
	FontAwesomeLayers,
	FontAwesomeLayersText,
} from "./icons";
import language from "./languages";
import router from "./router";

const app = createApp(App);

app.component("font-awesome-icon", FontAwesomeIcon);
app.component("font-awesome-layers", FontAwesomeLayers);
app.component("font-awesome-layers-text", FontAwesomeLayersText);

const i18n = createI18n({
	legacy: false,
	locale: "en",
	fallbackLocale: "en",
	messages: language,
});

app.use(createPinia());
app.use(router);
app.use(i18n);
app.use(VueApexCharts);
app.use(VueCookies);

app.mount("#app");
