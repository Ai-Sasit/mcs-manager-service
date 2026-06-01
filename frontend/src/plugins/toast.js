import { ElNotification, ElMessage } from "element-plus";

class Toast {
  success(title, message, duration = 3000, position = "bottom-right") {
    ElNotification({
      title,
      message,
      type: "success",
      duration,
      position,
    });
  }

  error(title, message, duration = 3000, position = "bottom-right") {
    ElNotification({
      title,
      message,
      type: "error",
      duration,
      position,
    });
  }

  warning(title, message, duration = 3000, position = "bottom-right") {
    ElNotification({
      title,
      message,
      type: "warning",
      duration,
      position,
    });
  }

  info(title, message, duration = 3000, position = "bottom-right") {
    ElNotification({
      title,
      message,
      type: "info",
      duration,
      position,
    });
  }

  clear() {
    ElNotification.closeAll();
  }

  boxSuccess(message, duration = 3000) {
    ElMessage({ message, type: "success", duration });
  }

  boxError(message, duration = 3000) {
    ElMessage({ message, type: "error", duration });
  }

  boxWarning(message, duration = 3000) {
    ElMessage({ message, type: "warning", duration });
  }

  boxInfo(message, duration = 3000) {
    ElMessage({ message, type: "info", duration });
  }
}

export default {
  install: (app) => {
    app.config.globalProperties.$toast = new Toast();
  },
};