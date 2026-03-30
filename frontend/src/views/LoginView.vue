<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-logo">
        <div class="logo-mark">
          <el-icon size="24" color="#ffffff"><Monitor /></el-icon>
        </div>
        <h2>MC Manage</h2>
        <p>Server Administration Panel</p>
      </div>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        @submit.prevent="handleLogin">
        <el-form-item label="Username" prop="username">
          <el-input
            v-model="form.username"
            placeholder="Enter username"
            size="large"
            clearable />
        </el-form-item>
        <el-form-item label="Password" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="Enter password"
            size="large"
            show-password
            @keyup.enter="handleLogin" />
        </el-form-item>
        <el-alert
          v-if="error"
          :title="error"
          type="error"
          show-icon
          :closable="false"
          style="margin-bottom: 20px" />
        <el-button
          type="primary"
          size="large"
          class="login-btn"
          :loading="loading"
          @click="handleLogin">
          Sign In
        </el-button>
      </el-form>
      
      <div class="login-footer">
        Secure Access Only
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "../stores/auth";
import { Monitor } from "@element-plus/icons-vue";

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
    error.value = e.response?.data?.message || "Login failed";
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg);
}

.login-card {
  width: 420px;
  background: var(--color-white);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.06);
  padding: 48px 40px;
}

.login-logo {
  text-align: center;
  margin-bottom: 36px;
}

.logo-mark {
  width: 52px;
  height: 52px;
  border-radius: 12px;
  background: var(--color-primary);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16px;
}

.login-logo h2 {
  font-size: 22px;
  font-weight: 700;
  color: var(--color-text);
  margin: 0;
}

.login-logo p {
  color: var(--color-text-secondary);
  font-size: 14px;
  margin-top: 4px;
}

.login-btn {
  width: 100%;
  height: 44px;
  font-weight: 600 !important;
}

.login-footer {
  text-align: center;
  margin-top: 28px;
  font-size: 12px;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

:deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--color-text);
  font-size: 14px;
  margin-bottom: 4px;
}
</style>
