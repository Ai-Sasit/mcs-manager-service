import { ElLoading } from "element-plus";

class LoadingScreen {
  constructor() {
    this.loadingInstance = null;
  }

  on() {
    if (!this.loadingInstance) {
      this.loadingInstance = ElLoading.service({
        lock: true,
        text: "Loading",
        background: "rgba(0, 0, 0, 0.7)",
      });
    }
  }

  off() {
    if (this.loadingInstance) {
      this.loadingInstance.close();
      this.loadingInstance = null;
    }
  }
}

export default {
  install(app) {
    app.config.globalProperties.$loading = new LoadingScreen();
  },
};