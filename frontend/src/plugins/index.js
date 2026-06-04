import router from "@/router";
import apiService from "@/api";
import toast from "./toast";
import loadingScreen from "./loadingScreen";
import ElementPlus from "element-plus";
import { createPinia } from "pinia";
export function registerPlugins(app) {
  app.use(createPinia());
  app.use(ElementPlus);
  app.use(router);
  app.use(apiService);
  app.use(toast);
  app.use(loadingScreen);
}