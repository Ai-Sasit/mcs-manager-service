import { createApp } from "vue";
import App from "@/App.vue";
import { registerPlugins } from "@/plugins";

import "element-plus/dist/index.css";
import "@/assets/themes/tokens.scss";
import "@/assets/themes/global.scss";
import "@/assets/themes/element-overrides.scss";
import { initTheme } from "@/composables/useTheme";

// Resolve the user preference before mounting to avoid a light-theme flash.
initTheme();

const app = createApp(App);

registerPlugins(app);

app.mount("#app");
