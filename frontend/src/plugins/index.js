import { createPinia } from "pinia";
import ElementPlus from "element-plus";
import router from "@/router";

export function registerPlugins(app) {
  app.use(createPinia());
  app.use(ElementPlus);
  app.use(router);
}
