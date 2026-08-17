<template>
  <div class="login-page">
    <button
      class="theme-toggle"
      type="button"
      :aria-label="isDark ? 'Switch to light theme' : 'Switch to dark theme'"
      :title="isDark ? 'Light theme' : 'Dark theme'"
      @click="toggleTheme"
    >
      <PhSun v-if="isDark" :size="19" />
      <PhMoon v-else :size="19" />
    </button>

    <div class="login-card">
      <div class="login-logo">
        <div class="brand-line">
          <span class="logo-mark"><PhDesktop :size="21" weight="bold" /></span>
          <span>MC Manage</span>
        </div>
        <h1>Welcome back</h1>
        <p>Sign in to manage your Minecraft servers.</p>
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
          Sign in
        </el-button>
      </el-form>

      <div class="login-footer">
        <span>Secure access only</span>
        <span>Version 1.0.0</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { PhDesktop, PhMoon, PhSun } from "@phosphor-icons/vue";
import { getApiErrorMessage } from "@/utils/apiError";
import { useTheme } from "@/composables/useTheme";

const router = useRouter();
const auth = useAuthStore();
const formRef = ref(null);
const loading = ref(false);
const error = ref("");
const { isDark, toggleTheme } = useTheme();

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
  display: flex;
  min-height: 100vh;
  align-items: center;
  justify-content: center;
  padding: 32px;
  background:
    linear-gradient(120deg, color-mix(in srgb, var(--color-bg) 82%, transparent), color-mix(in srgb, var(--color-primary-bg) 58%, transparent)),
    url("/login-bg-light.webp") center / cover no-repeat,
    var(--color-bg);
}

.login-card {
  width: min(100%, 420px);
  padding: 32px;
  background: var(--color-layer);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-flyout);
}

.theme-toggle {
  position: fixed;
  top: 20px;
  right: 20px;
  display: inline-flex;
  width: 36px;
  height: 36px;
  align-items: center;
  justify-content: center;
  padding: 0;
  color: var(--color-text-secondary);
  cursor: pointer;
  background: var(--color-layer);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  box-shadow: var(--shadow-sm);
}

.theme-toggle:hover {
  color: var(--color-text);
  background: var(--color-layer-alt);
}

.login-logo {
  margin-bottom: 28px;
}

.brand-line {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 24px;
  color: var(--color-text);
  font-size: 15px;
  font-weight: 600;
}

.logo-mark {
  display: inline-flex;
  width: 30px;
  height: 30px;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  background: var(--color-primary);
  border-radius: var(--radius-sm);
}

.login-logo h1 {
  margin: 0;
  color: var(--color-text);
  font-size: 26px;
  font-weight: 600;
  letter-spacing: -0.02em;
}

.login-logo p {
  margin: 7px 0 0;
  color: var(--color-text-secondary);
}

.login-alert {
  margin-bottom: 16px;
}

.login-btn {
  width: 100%;
  min-height: 40px;
  margin-top: 4px;
}

.login-footer {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-top: 24px;
  color: var(--color-text-muted);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

:deep(.el-form-item__label) {
  color: var(--color-text) !important;
  font-weight: 600;
}

:deep(.el-input__wrapper) {
  min-height: 40px;
}

@media (max-width: 560px) {
  .login-page {
    padding: 16px;
  }

  .login-card {
    padding: 24px 20px;
  }

  .theme-toggle {
    top: 12px;
    right: 12px;
  }

  .login-footer {
    flex-direction: column;
  }
}
</style>
