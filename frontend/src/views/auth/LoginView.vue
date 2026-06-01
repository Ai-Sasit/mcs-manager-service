<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-logo">
        <div class="brand-line">
          <div class="logo-mark">
            <el-icon size="22" color="#ffffff"><Monitor /></el-icon>
          </div>
          <span>MC Manage</span>
        </div>
        <h1>Welcome back</h1>
      </div>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        @submit.prevent="handleLogin"
      >
        <el-form-item label="Username" prop="username">
          <el-input
            v-model="form.username"
            placeholder="Enter username"
            size="large"
            clearable
          />
        </el-form-item>
        <el-form-item label="Password" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="Enter password"
            size="large"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <el-alert
          v-if="error"
          :title="error"
          type="error"
          show-icon
          :closable="false"
          class="login-alert"
        />
        <el-button
          type="primary"
          size="large"
          class="login-btn"
          :loading="loading"
          @click="handleLogin"
        >
          Sign In
        </el-button>
      </el-form>

      <div class="login-footer">
        <span>Secure Access Only</span>
        <span>Version 1.0.0</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { Monitor } from "@element-plus/icons-vue";
import { getApiErrorMessage } from "@/utils/apiError";

const router = useRouter();
const auth = useAuthStore();
const formRef = ref(null);
const loading = ref(false);
const error = ref("");

const form = ref({ username: "", password: "" });

const rules = {
  username: [
    { required: true, message: "Username is required", trigger: "blur" },
  ],
  password: [
    { required: true, message: "Password is required", trigger: "blur" },
  ],
};

async function handleLogin() {
  if (!formRef.value) return;
  const valid = await formRef.value.validate().catch(() => false);
  if (!valid) return;
  loading.value = true;
  error.value = "";
  try {
    await auth.login(form.value.username, form.value.password);
    router.push("/");
  } catch (e) {
    error.value = getApiErrorMessage(e, "Login failed");
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding: 48px 7vw;
  background:
    linear-gradient(90deg, rgba(244, 252, 248, 0.18), rgba(240, 253, 250, 0.44) 52%, rgba(255, 255, 255, 0.28)),
    url("/login-bg-light.webp") center / cover no-repeat,
    #eaf7ef;
}

.login-page::before {
  content: "";
  position: absolute;
  inset: 0;
  background:
    radial-gradient(circle at 22% 18%, rgba(255, 255, 255, 0.62), transparent 34%),
    linear-gradient(90deg, rgba(255, 255, 255, 0.08), rgba(236, 253, 245, 0.28) 56%, rgba(255, 255, 255, 0.62));
  pointer-events: none;
}

.login-card {
  position: relative;
  z-index: 1;
  width: min(100%, 420px);
  padding: 34px;
  border: 1px solid rgba(255, 255, 255, 0.68);
  border-radius: var(--radius);
  background: rgba(255, 255, 255, 0.58);
  box-shadow: 0 24px 80px rgba(15, 23, 42, 0.18);
  backdrop-filter: blur(20px) saturate(145%);
  -webkit-backdrop-filter: blur(20px) saturate(145%);
}

.login-logo {
  margin-bottom: 28px;
}

.brand-line {
  display: flex;
  align-items: center;
  gap: 12px;
  color: var(--color-text);
  font-size: 17px;
  font-weight: 800;
  margin-bottom: 18px;
}

.logo-mark {
  width: 44px;
  height: 44px;
  border-radius: var(--radius);
  background: linear-gradient(135deg, var(--color-primary), var(--color-accent));
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 12px 28px rgba(16, 185, 129, 0.32);
}

.login-logo h1 {
  font-size: 26px;
  font-weight: 800;
  color: var(--color-text);
  margin: 0;
}

.login-alert {
  margin-bottom: 20px;
}

.login-btn {
  width: 100%;
  height: 44px;
  font-weight: 700 !important;
}

.login-footer {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-top: 24px;
  font-size: 12px;
  color: rgba(31, 41, 55, 0.58);
  text-transform: uppercase;
  letter-spacing: 0;
}

:deep(.el-form-item__label) {
  font-weight: 600;
  color: rgba(31, 41, 55, 0.82);
  font-size: 14px;
  margin-bottom: 4px;
}

:deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.88) !important;
  border-color: rgba(148, 163, 184, 0.42) !important;
}

:deep(.el-input__wrapper.is-focus) {
  border-color: var(--color-primary) !important;
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.2) !important;
}

@media (max-width: 560px) {
  .login-page {
    padding: 18px;
    justify-content: center;
  }

  .login-card {
    padding: 28px 20px;
  }

  .login-logo h1 {
    font-size: 23px;
  }

  .login-footer {
    flex-direction: column;
  }
}
</style>
