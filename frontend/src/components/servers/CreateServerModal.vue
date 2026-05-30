<template>
  <teleport to="body">
    <transition name="fade">
      <div class="modal-overlay" @click.self="$emit('close')">
        <div class="modal">
          <div class="modal-header">
            <h2>Create New Server</h2>
            <button class="close-btn" @click="$emit('close')">
              <el-icon size="18"><Close /></el-icon>
            </button>
          </div>

          <form @submit.prevent="handleCreate" class="modal-form">
            <div class="form-group">
              <label>Server Edition</label>
              <div class="edition-selector">
                <button
                  type="button"
                  class="edition-btn"
                  :class="{ active: form.edition === 'java' }"
                  @click="setEdition('java')">
                  ☕
                  <span>Java</span>
                </button>
                <button
                  type="button"
                  class="edition-btn"
                  :class="{ active: form.edition === 'bedrock' }"
                  @click="setEdition('bedrock')">
                  💎
                  <span>Bedrock</span>
                </button>
              </div>
            </div>

            <div class="form-group" v-if="form.edition === 'java'">
              <label>Server Type</label>
              <div class="edition-selector">
                <button
                  type="button"
                  class="edition-btn"
                  :class="{ active: form.server_type === 'vanilla' }"
                  @click="form.server_type = 'vanilla'">
                  🟢
                  <span>Vanilla</span>
                </button>
                <button
                  type="button"
                  class="edition-btn"
                  :class="{ active: form.server_type === 'paper' }"
                  @click="form.server_type = 'paper'">
                  📄
                  <span>Paper</span>
                </button>
                <button
                  type="button"
                  class="edition-btn"
                  :class="{ active: form.server_type === 'spigot' }"
                  @click="form.server_type = 'spigot'">
                  🔧
                  <span>Spigot</span>
                </button>
              </div>
              <p class="type-hint" v-if="form.server_type === 'paper'">
                Recommended for plugins. Best performance.
              </p>
              <p class="type-hint" v-else-if="form.server_type === 'spigot'">
                Supports Bukkit/Spigot plugins.
              </p>
              <p class="type-hint" v-else>
                Official Mojang server. No plugin support.
              </p>
            </div>

            <div class="form-group">
              <label>Version</label>
              <el-select
                v-model="form.version"
                placeholder="Select version..."
                :loading="loadingVersions"
                size="large"
                style="width: 100%">
                <el-option
                  v-for="v in versions"
                  :key="v.id"
                  :label="
                    form.edition === 'bedrock' ? `Latest (${v.id})` : v.id
                  "
                  :value="v.id" />
              </el-select>
            </div>

            <div class="form-group">
              <label>Server Name</label>
              <el-input
                v-model="form.name"
                placeholder="production-server-1"
                maxlength="50"
                size="large" />
            </div>

            <div class="form-row">
              <div class="form-group">
                <label>Port</label>
                <el-input-number
                  v-model="form.port"
                  :min="1024"
                  :max="65535"
                  controls-position="right"
                  size="large"
                  style="width: 100%" />
              </div>
              <div class="form-group">
                <label>Max Players</label>
                <el-input-number
                  v-model="form.max_players"
                  :min="1"
                  :max="1000"
                  controls-position="right"
                  size="large"
                  style="width: 100%" />
              </div>
            </div>

            <div class="form-group" v-if="form.edition === 'java'">
              <label>Allocated Memory (MB)</label>
              <el-input-number
                v-model="form.memory_mb"
                :min="512"
                :step="256"
                controls-position="right"
                size="large"
                style="width: 100%" />
            </div>

            <el-alert
              v-if="error"
              :title="error"
              type="error"
              show-icon
              :closable="false" />

            <el-button
              type="primary"
              size="large"
              class="submit-btn"
              :loading="creating"
              native-type="submit">
              {{ creating ? "Deploying..." : "Deploy Server" }}
            </el-button>
          </form>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { Close } from "@element-plus/icons-vue";
import api from "@/api";
import {
  DEFAULT_JAVA_PORT,
  DEFAULT_BEDROCK_PORT,
  DEFAULT_MAX_PLAYERS,
  DEFAULT_MEMORY_MB,
} from "@/constants";

const emit = defineEmits(["close", "created"]);

const form = ref({
  name: "",
  edition: "java",
  server_type: "paper",
  version: "",
  port: DEFAULT_JAVA_PORT,
  max_players: DEFAULT_MAX_PLAYERS,
  memory_mb: DEFAULT_MEMORY_MB,
});

const versions = ref([]);
const loadingVersions = ref(false);
const creating = ref(false);
const error = ref(null);

async function fetchVersions() {
  loadingVersions.value = true;
  form.value.version = "";
  try {
    const { data } =
      form.value.edition === "java"
        ? await api.getJavaVersions()
        : await api.getBedrockVersions();
    versions.value = data.data || [];

    // Auto-select Bedrock version
    if (form.value.edition === "bedrock" && versions.value.length > 0) {
      form.value.version = versions.value[0].id;
    }
  } catch (e) {
    console.error("Failed to fetch versions:", e);
  } finally {
    loadingVersions.value = false;
  }
}

function setEdition(edition) {
  form.value.edition = edition;
  form.value.port =
    edition === "java" ? DEFAULT_JAVA_PORT : DEFAULT_BEDROCK_PORT;
  form.value.server_type = edition === "java" ? "paper" : "";
  fetchVersions();
}

async function handleCreate() {
  creating.value = true;
  error.value = null;
  try {
    const { data } = await api.createServer(form.value);
    if (data.success) {
      emit("created", data.data);
      emit("close");
    } else {
      error.value = data.message;
    }
  } catch (e) {
    error.value = e.response?.data?.message || e.message;
  } finally {
    creating.value = false;
  }
}

onMounted(fetchVersions);
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 24px;
}

.modal {
  width: 100%;
  max-width: 520px;
  max-height: 90vh;
  overflow-y: auto;
  background: var(--color-white);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
  padding: 32px;
  animation: modalIn 0.25s ease-out;
}

@keyframes modalIn {
  from {
    opacity: 0;
    transform: scale(0.96) translateY(8px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 28px;
}

.modal-header h2 {
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text);
  margin: 0;
}

.close-btn {
  width: 32px;
  height: 32px;
  border-radius: var(--radius);
  border: 1px solid var(--color-border);
  background: var(--color-white);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary);
  transition: all 0.15s ease;
}

.close-btn:hover {
  background: var(--color-bg);
  color: var(--color-text);
}

.modal-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

label {
  font-weight: 500;
  color: var(--color-text);
  font-size: 14px;
}

.edition-selector {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.edition-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 20px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-white);
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: inherit;
  font-weight: 500;
  font-size: 14px;
  color: var(--color-text-secondary);
}

.edition-btn:hover {
  border-color: var(--color-text-muted);
  background: var(--color-bg);
}

.edition-btn.active {
  border-color: var(--color-primary);
  background: rgba(16, 185, 129, 0.04);
  color: var(--color-primary);
}

.type-hint {
  font-size: 12px;
  color: var(--color-text-muted);
  margin: 2px 0 0;
}

.submit-btn {
  width: 100%;
  height: 44px;
  font-weight: 600 !important;
  margin-top: 4px;
}

:deep(.el-input-number .el-input__wrapper) {
  padding-left: 12px;
}
</style>
