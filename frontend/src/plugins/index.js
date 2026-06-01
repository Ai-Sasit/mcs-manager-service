import router from "@/router";
import apiService from "@/api";
import toast from "./toast";
import loadingScreen from "./loadingScreen";
import ElementPlus from "element-plus";
import { createPinia } from "pinia";
import * as ElementPlusIconsVue from "@element-plus/icons-vue";

export function registerPlugins(app) {
  for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component);
  }

  app.use(createPinia());
  app.use(ElementPlus);
  app.use(router);
  app.use(apiService);
  app.use(toast);
  app.use(loadingScreen);
}