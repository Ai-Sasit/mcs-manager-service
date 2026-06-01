<template>
  <div class="console-page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">Global Console</h2>
        <p class="page-subtitle">Access raw server terminals and send commands in real-time.</p>
      </div>
      <div class="header-actions">
        <el-select 
          v-model="selectedServerId" 
          placeholder="Select a running server" 
          style="width: 250px"
          @change="onServerChange"
        >
          <el-option
            v-for="server in runningServers"
            :key="server.id"
            :label="server.name"
            :value="server.id"
          >
            <span style="float: left">{{ server.name }}</span>
            <span style="float: right; color: var(--color-success); font-size: 12px">●</span>
          </el-option>
        </el-select>
        <el-button :icon="Refresh" @click="store.fetchServers" :loading="store.loading">
          Refresh List
        </el-button>
      </div>
    </div>

    <el-card class="console-card" v-if="selectedServerId">
      <div class="terminal-container">
        <div class="terminal-header">
          <span>Terminal: {{ selectedServerName }}</span>
          <el-tag size="small" :type="terminalConnectionState === 'open' ? 'success' : 'warning'">
            {{ terminalConnectionState }}
          </el-tag>
          <el-button size="small" type="danger" plain @click="disconnectTerminal">Disconnect</el-button>
        </div>
        <div class="log-box" ref="termBox">
          <div v-for="(line, i) in termLines" :key="i" class="log-line">{{ line }}</div>
          <div v-if="termLines.length === 0" class="log-empty">
            Establishing connection to {{ selectedServerName }}...
          </div>
        </div>
        <div class="cmd-row">
          <el-input
            v-model="cmdInput"
            placeholder="Enter server command... (e.g. /say hello)"
            @keyup.enter="sendCommand"
          >
            <template #prepend><span class="cmd-prompt">/</span></template>
          </el-input>
          <el-button 
            type="primary" 
            @click="sendCommand"
            :disabled="terminalConnectionState !== 'open'"
          >
            Execute
          </el-button>
        </div>
      </div>
    </el-card>

    <div v-else class="center-state card">
      <el-icon size="48" color="var(--color-text-muted)"><Cpu /></el-icon>
      <h3>No Server Selected</h3>
      <p>Please select a running server from the dropdown above to access its console.</p>
      <p v-if="runningServers.length === 0" style="color: var(--color-danger)">
        There are currently no running servers. Start a server first.
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from "vue";
import { Refresh, Cpu } from "@element-plus/icons-vue";
import { ElMessage } from "element-plus";
import { useServersStore } from "@/stores/servers";
import { useServerTerminal } from "@/composables/useServerTerminal";

const store = useServersStore();
const selectedServerId = ref(null);

const runningServers = computed(() => {
  return store.servers.filter(s => s.status === 'running');
});

const selectedServerName = computed(() => {
  const s = store.servers.find(s => s.id === selectedServerId.value);
  return s ? s.name : "UnknownServer";
});

const termLines = ref([]);
const cmdInput = ref("");
const termBox = ref(null);
const terminalConnectionState = ref("idle");
let termComp = null;

onMounted(() => {
  store.fetchServers();
});

onUnmounted(() => {
  disconnectTerminal();
});

function scrollToBottom(el) {
  nextTick(() => {
    if (el) el.scrollTop = el.scrollHeight;
  });
}

function onServerChange() {
  disconnectTerminal();
  termLines.value = [];
  
  if (!selectedServerId.value) return;

  termComp = useServerTerminal(selectedServerId.value);
  termComp.onStateChange((state) => {
    terminalConnectionState.value = state;
  });
  termComp.onLine((line) => {
    termLines.value.push(line);
    if (termLines.value.length > 3000) termLines.value.shift();
    scrollToBottom(termBox.value);
  });
  termComp.connect();
  terminalConnectionState.value = termComp.state.value;
}

function disconnectTerminal() {
  if (termComp) {
    termComp.disconnect();
    termComp = null;
  }
  terminalConnectionState.value = "closed";
}

function sendCommand() {
  if (!cmdInput.value.trim() || !termComp) return;
  if (!termComp.sendCommand(cmdInput.value.trim())) {
    ElMessage.warning("Terminal is not connected.");
    return;
  }
  cmdInput.value = "";
}
</script>

<style scoped>
.console-page {
  animation: fadeIn 0.3s ease-out;
  display: flex;
  flex-direction: column;
  height: calc(100vh - 120px);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-title {
  margin: 0 0 4px;
  font-size: 24px;
  font-weight: 600;
}

.page-subtitle {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 14px;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.console-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}

:deep(.console-card .el-card__body) {
  padding: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.terminal-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.terminal-header {
  display: flex;
  gap: 12px;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  background: var(--color-bg);
  border-bottom: 1px solid var(--color-border);
  font-weight: 600;
  font-size: 14px;
}

.log-box {
  flex: 1;
  background: var(--color-white);
  padding: 20px;
  overflow-y: auto;
  font-family: 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.7;
  min-height: 300px;
}

.log-line {
  color: var(--color-text);
  margin-bottom: 2px;
  word-break: break-all;
}

.log-empty {
  color: var(--color-text-muted);
  font-style: italic;
}

.cmd-row {
  display: flex;
  gap: 12px;
  padding: 16px 20px;
  background: var(--color-bg);
  border-top: 1px solid var(--color-border);
}

.cmd-prompt {
  color: var(--color-primary);
  font-weight: 700;
  font-size: 16px;
}

.center-state {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  padding: 80px 20px;
  text-align: center;
  gap: 16px;
}

.center-state h3 {
  margin: 0;
  font-size: 20px;
}

.center-state p {
  margin: 0;
  color: var(--color-text-secondary);
  max-width: 400px;
}
</style>
