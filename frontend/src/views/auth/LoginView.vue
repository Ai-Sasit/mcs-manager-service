<template>
  <div class="login-page">
    <section class="hero-panel" aria-label="MC Manage overview">
      <div class="hero-art">
        <div class="sun"></div>
        <div class="cloud cloud-one"></div>
        <div class="cloud cloud-two"></div>
        <div class="mountain mountain-back"></div>
        <div class="mountain mountain-front"></div>
        <div class="terrain">
          <div class="server-rack">
            <div class="rack-light active"></div>
            <div class="rack-light"></div>
            <div class="rack-line"></div>
            <div class="rack-line short"></div>
          </div>
          <div class="console-window">
            <span>$ java -jar server.jar</span>
            <span class="ok">world ready</span>
          </div>
        </div>
      </div>
      <div class="hero-copy">
        <div class="brand-row">
          <div class="brand-mark">
            <el-icon size="22"><Monitor /></el-icon>
          </div>
          <span>MC Manage</span>
        </div>
        <h1>Minecraft operations, clean and controlled.</h1>
        <p>Monitor servers, manage players, edit configs, and keep setup work visible from the first download.</p>
      </div>
    </section>

    <section class="form-panel">
      <div class="login-card">
        <div class="login-logo">
          <div class="logo-mark">
            <el-icon size="24" color="#ffffff"><Monitor /></el-icon>
          </div>
          <h2>Welcome back</h2>
          <p>Sign in to continue to MC Manage.</p>
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
            style="margin-bottom: 20px"
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
    </section>
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
  display: grid;
  grid-template-columns: minmax(0, 1.05fr) minmax(420px, 0.95fr);
  background: #f5f7fa;
}

.hero-panel {
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 48px;
  background: linear-gradient(180deg, #dff7ff 0%, #e9fbf2 56%, #9ad17b 56%, #6aa84f 100%);
}

.hero-art {
  position: absolute;
  inset: 0;
}

.sun {
  position: absolute;
  top: 58px;
  right: 80px;
  width: 86px;
  height: 86px;
  background: #ffd166;
  box-shadow: inset -10px -10px 0 rgba(219, 136, 22, 0.18);
}

.cloud {
  position: absolute;
  background: #ffffff;
  box-shadow: 32px 0 0 #ffffff, 64px 0 0 #ffffff, 32px -24px 0 #ffffff;
  height: 28px;
  width: 34px;
  opacity: 0.92;
}

.cloud-one {
  top: 112px;
  left: 18%;
}

.cloud-two {
  top: 190px;
  right: 22%;
  transform: scale(0.72);
}

.mountain {
  position: absolute;
  bottom: 36%;
  width: 34vw;
  height: 22vw;
  background: #6b7280;
  transform: rotate(45deg);
}

.mountain-back {
  left: 4%;
  opacity: 0.42;
}

.mountain-front {
  right: 10%;
  background: #52616b;
  opacity: 0.62;
}

.terrain {
  position: absolute;
  inset-inline: 0;
  bottom: 0;
  height: 44%;
  background:
    linear-gradient(90deg, rgba(0, 0, 0, 0.08) 1px, transparent 1px) 0 0 / 42px 42px,
    linear-gradient(#6aa84f 0 26%, #8b5a2b 26% 100%);
  border-top: 10px solid #4f8f3b;
}

.server-rack,
.console-window {
  position: absolute;
  border: 3px solid rgba(31, 41, 55, 0.88);
  box-shadow: 10px 10px 0 rgba(31, 41, 55, 0.16);
}

.server-rack {
  left: 16%;
  bottom: 33%;
  width: 120px;
  height: 150px;
  background: #263238;
  padding: 18px;
}

.rack-light {
  width: 14px;
  height: 14px;
  background: #64748b;
  margin-bottom: 16px;
}

.rack-light.active {
  background: #22c55e;
  box-shadow: 0 0 18px rgba(34, 197, 94, 0.65);
}

.rack-line {
  width: 68px;
  height: 8px;
  background: #94a3b8;
  margin-bottom: 14px;
}

.rack-line.short {
  width: 44px;
}

.console-window {
  right: 13%;
  bottom: 42%;
  width: min(320px, 36vw);
  min-height: 128px;
  background: #111827;
  color: #d1fae5;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 22px;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 13px;
}

.console-window .ok {
  color: #6ee7b7;
}

.hero-copy {
  position: relative;
  z-index: 1;
  max-width: 560px;
  margin-top: auto;
  color: #17301f;
}

.brand-row {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  font-weight: 800;
  font-size: 18px;
  margin-bottom: 22px;
}

.brand-mark,
.logo-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.brand-mark {
  width: 42px;
  height: 42px;
  background: #111827;
  color: #ffffff;
  border: 3px solid rgba(255, 255, 255, 0.7);
}

.hero-copy h1 {
  max-width: 620px;
  color: #17301f;
  font-size: 60px;
  line-height: 1;
  font-weight: 900;
  margin: 0 0 22px;
}

.hero-copy p {
  max-width: 520px;
  color: rgba(23, 48, 31, 0.78);
  font-size: 17px;
  line-height: 1.7;
}

.form-panel {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 42px;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.92), rgba(245, 247, 250, 0.96)),
    var(--color-bg);
}

.login-card {
  width: min(100%, 430px);
  background: var(--color-white);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: 0 16px 44px rgba(31, 41, 55, 0.08);
  padding: 42px 38px;
}

.login-logo {
  margin-bottom: 34px;
}

.logo-mark {
  width: 52px;
  height: 52px;
  border-radius: var(--radius);
  background: var(--color-primary);
  margin-bottom: 18px;
}

.login-logo h2 {
  font-size: 26px;
  font-weight: 800;
  color: var(--color-text);
  margin: 0;
}

.login-logo p {
  color: var(--color-text-secondary);
  font-size: 14px;
  margin-top: 6px;
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
  margin-top: 28px;
  font-size: 12px;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0;
}

:deep(.el-form-item__label) {
  font-weight: 600;
  color: var(--color-text);
  font-size: 14px;
  margin-bottom: 4px;
}

@media (max-width: 960px) {
  .login-page {
    grid-template-columns: 1fr;
  }

  .hero-panel {
    min-height: 420px;
    padding: 32px 24px;
  }

  .hero-copy h1 {
    font-size: 38px;
  }

  .form-panel {
    padding: 28px 18px;
  }
}

@media (max-width: 560px) {
  .hero-panel {
    min-height: 360px;
  }

  .console-window {
    right: 24px;
    width: calc(100% - 150px);
    min-height: 104px;
    padding: 16px;
    font-size: 11px;
  }

  .server-rack {
    left: 24px;
    width: 92px;
    height: 122px;
  }

  .hero-copy p {
    font-size: 14px;
  }

  .login-card {
    padding: 32px 22px;
  }

  .login-footer {
    flex-direction: column;
  }
}
</style>
