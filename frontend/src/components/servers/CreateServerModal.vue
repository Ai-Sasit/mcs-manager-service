<template>
  <teleport to="body">
    <transition name="fade">
      <div class="modal-overlay" @click.self="handleClose">
        <div class="modal">
          <div class="modal-header">
            <h2>Create New Server</h2>
            <button class="close-btn" :disabled="creating" @click="handleClose">
              <PhX :size="18" weight="regular" />
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
                  @click="setEdition('java')"
                >
                  ☕ <span>Java</span>
                </button>
                <button
                  type="button"
                  class="edition-btn"
                  :class="{ active: form.edition === 'bedrock' }"
                  @click="setEdition('bedrock')"
                >
                  💎 <span>Bedrock</span>
                </button>
              </div>
            </div>

            <div class="form-group" v-if="form.edition === 'java'">
              <label>Server Type</label>
              <div class="type-selector">
                <button
                  type="button"
                  class="edition-btn"
                  :class="{ active: form.server_type === 'vanilla' }"
                  @click="setServerType('vanilla')"
                >
                  🟢 <span>Vanilla</span>
                </button>
                <button
                  type="button"
                  class="edition-btn"
                  :class="{ active: form.server_type === 'paper' }"
                  @click="setServerType('paper')"
                >
                  📄 <span>Paper</span>
                </button>
                <button
                  type="button"
                  class="edition-btn"
                  :class="{ active: form.server_type === 'spigot' }"
                  @click="setServerType('spigot')"
                >
                  🔧 <span>Spigot</span>
                </button>
                <button
                  type="button"
                  class="edition-btn"
                  :class="{ active: form.server_type === 'forge' }"
                  @click="setServerType('forge')"
                >
                  ⚒️ <span>Forge</span>
                </button>
                <button
                  type="button"
                  class="edition-btn"
                  :class="{ active: form.server_type === 'fabric' }"
                  @click="setServerType('fabric')"
                >
                  🧵 <span>Fabric</span>
                </button>
              </div>
              <p class="type-hint" v-if="form.server_type === 'paper'">
                Recommended for plugins. Best performance.
              </p>
              <p class="type-hint" v-else-if="form.server_type === 'spigot'">
                Supports Bukkit/Spigot plugins.
              </p>
              <p class="type-hint" v-else-if="form.server_type === 'forge'">
                Modded server with Forge mod loader. Upload mods via Mods page.
              </p>
              <p class="type-hint" v-else-if="form.server_type === 'fabric'">
                Lightweight modded server with Fabric loader. Upload mods via
                Mods page.
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
                style="width: 100%"
              >
                <el-option
                  v-for="v in versions"
                  :key="v.id"
                  :label="
                    form.edition === 'bedrock' ? `Latest (${v.id})` : v.id
                  "
                  :value="v.id"
                />
              </el-select>
            </div>

            <div class="form-group">
              <label>Server Name</label>
              <el-input
                v-model="form.name"
                placeholder="production-server-1"
                maxlength="50"
                size="large"
              />
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
                  style="width: 100%"
                />
              </div>
              <div class="form-group">
                <label>Max Players</label>
                <el-input-number
                  v-model="form.max_players"
                  :min="1"
                  :max="1000"
                  controls-position="right"
                  size="large"
                  style="width: 100%"
                />
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
                style="width: 100%"
              />
            </div>

            <el-alert
              v-if="error"
              :title="error"
              type="error"
              show-icon
              :closable="false"
            />

            <div v-if="setupEvents.length" class="setup-progress">
              <div class="progress-head">
                <span>Installation progress</span>
                <strong>{{ setupPercent }}%</strong>
              </div>
              <el-progress :percentage="setupPercent" :stroke-width="8" />
              <div
                v-if="setupConnectionState !== 'open'"
                class="connection-note"
                :class="setupConnectionState"
              >
                Progress stream: {{ setupConnectionState }}
              </div>
              <div class="setup-steps">
                <div
                  v-for="event in setupEvents"
                  :key="`${event.step}-${event.status}-${event.percent}`"
                  class="setup-step"
                  :class="event.status"
                >
                  <span class="step-dot"></span>
                  <span>{{ event.message }}</span>
                </div>
              </div>
            </div>

            <el-button
              type="primary"
              size="large"
              class="submit-btn"
              :loading="creating"
              native-type="submit"
            >
              {{ creating ? "Deploying..." : "Deploy Server" }}
            </el-button>
          </form>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from "vue";
import { PhX } from "@phosphor-icons/vue";
import apiClient from "@/api/client";
import { useServerSetupStream } from "@/composables/useServerSetupStream";
import {
  DEFAULT_JAVA_PORT,
  DEFAULT_BEDROCK_PORT,
  DEFAULT_MAX_PLAYERS,
  DEFAULT_MEMORY_MB,
} from "@/constants";
import { getApiErrorMessage } from "@/utils/apiError";

// Note: isApiSuccess is no longer needed since we switched to SSE streaming

const emit = defineEmits(["close", "created"]);

const form = ref({
  name: "",
  edition: "java",
  server_type: "paper",
  mod_loader: "",
  version: "",
  port: DEFAULT_JAVA_PORT,
  max_players: DEFAULT_MAX_PLAYERS,
  memory_mb: DEFAULT_MEMORY_MB,
});

const versions = ref([]);
const loadingVersions = ref(false);
const creating = ref(false);
let error = ref(null);
let setupEvents = ref([]);
let setupPercent = ref(0);
let setupConnectionState = ref("idle");
let setupStream = null;

async function fetchVersions() {
  loadingVersions.value = true;
  form.value.version = "";
  try {
    const { data } =
      form.value.edition === "java"
        ? await apiClient.get("/versions/java")
        : await apiClient.get("/versions/bedrock");
    versions.value = data.data || [];
    if (form.value.edition === "bedrock" && versions.value.length > 0) {
      form.value.version = versions.value[0].id;
    }
  } catch (e) {
    error.value = getApiErrorMessage(e, "Failed to fetch versions");
  } finally {
    loadingVersions.value = false;
  }
}

function setEdition(edition) {
  form.value.edition = edition;
  form.value.port =
    edition === "java" ? DEFAULT_JAVA_PORT : DEFAULT_BEDROCK_PORT;
  form.value.server_type = edition === "java" ? "paper" : "";
  form.value.mod_loader = "";
  fetchVersions();
}

function setServerType(type) {
  form.value.server_type = type;
  form.value.mod_loader = type === "forge" || type === "fabric" ? type : "";
}

function handleClose() {
  if (creating.value) return;
  emit("close");
}

async function handleCreate() {
  creating.value = true;
  error.value = null;
  setupEvents.value = [];
  setupPercent.value = 0;
  if (setupStream) {
    setupStream.disconnect();
    setupStream = null;
  }

  try {
    setupStream = useServerSetupStream();
    // Reassign reactive refs so template directly binds to composable state
    setupEvents = setupStream.setupEvents;
    setupPercent = setupStream.setupPercent;
    setupConnectionState = setupStream.state;
    error = setupStream.error;

    const result = await setupStream.start({
      name: form.value.name,
      edition: form.value.edition,
      server_type: form.value.server_type,
      mod_loader: form.value.mod_loader,
      version: form.value.version,
      port: form.value.port,
      max_players: form.value.max_players,
      memory_mb: form.value.memory_mb,
    });

    if (result && result.status === "failed") {
      creating.value = false;
      return;
    }

    if (result && result.status === "success" && result.step === "complete") {
      creating.value = false;
      emit("created", { id: result.server_id });
      window.setTimeout(() => emit("close"), 650);
      return;
    }

    // Stream ended without a terminal event
    creating.value = false;
  } catch (e) {
    error.value = getApiErrorMessage(e, "Failed to create server");
    creating.value = false;
  }
}

onMounted(fetchVersions);
onUnmounted(() => {
  setupStream?.disconnect();
});
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
.close-btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
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
.type-selector {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
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
.setup-progress {
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  background: var(--color-bg);
  padding: 14px;
}
.progress-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--color-text);
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 8px;
}
.setup-steps {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 12px;
  max-height: 150px;
  overflow-y: auto;
}
.connection-note {
  margin-top: 10px;
  font-size: 12px;
  color: var(--color-text-muted);
}
.connection-note.reconnecting,
.connection-note.connecting {
  color: var(--color-warning);
}
.connection-note.error {
  color: var(--color-danger);
}
.setup-step {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--color-text-secondary);
  font-size: 12px;
}
.step-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--color-text-muted);
  flex: 0 0 auto;
}
.setup-step.active .step-dot {
  background: var(--color-warning);
  animation: pulse 1s infinite alternate;
}
.setup-step.success .step-dot {
  background: var(--color-success);
}
.setup-step.failed {
  color: var(--color-danger);
}
.setup-step.failed .step-dot {
  background: var(--color-danger);
}
:deep(.el-input-number .el-input__wrapper) {
  padding-left: 12px;
}
@keyframes pulse {
  from {
    opacity: 0.45;
  }
  to {
    opacity: 1;
  }
}
</style>
