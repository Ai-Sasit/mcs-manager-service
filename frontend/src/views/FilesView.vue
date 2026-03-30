<template>
  <div class="files-view page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">File Explorer</h2>
        <p class="page-subtitle">Browse and edit server files directly from the web.</p>
      </div>
      <div class="header-actions">
        <el-select v-model="selectedServerId" placeholder="Select Server" @change="fetchFiles" style="width: 250px">
          <el-option
            v-for="server in servers"
            :key="server.id"
            :label="server.name"
            :value="server.id" />
        </el-select>
        <el-button :icon="Refresh" @click="fetchFiles" :loading="loading" :disabled="!selectedServerId">
          Refresh
        </el-button>
      </div>
    </div>

    <div v-if="selectedServerId" class="files-layout">
      <!-- File Tree -->
      <el-card class="file-tree-card card">
        <div class="breadcrumb-row">
          <el-button size="small" :icon="ArrowLeft" v-if="currentPath" @click="goUp">Back</el-button>
          <span class="path-text">{{ currentPath || '/' }}</span>
        </div>
        <el-table :data="files" stripe style="width: 100%" @row-click="handleFileClick" row-class-name="file-row">
           <el-table-column label="Name" min-width="180">
            <template #default="{ row }">
              <div class="file-name">
                <el-icon v-if="row.is_dir"><Folder /></el-icon>
                <el-icon v-else><Document /></el-icon>
                <span>{{ row.name }}</span>
              </div>
            </template>
           </el-table-column>
           <el-table-column prop="size" label="Size" width="100">
             <template #default="{ row }">
               {{ row.is_dir ? '-' : formatSize(row.size) }}
             </template>
           </el-table-column>
        </el-table>
      </el-card>

      <!-- Editor -->
      <el-card class="editor-card card" v-loading="editorLoading">
        <div v-if="editingFile" class="editor-header">
          <span class="editing-title">Editing: {{ editingFile }}</span>
          <el-button type="primary" size="small" :icon="Check" @click="saveFile" :loading="saving">Save Changes</el-button>
        </div>
        <div v-if="editingFile" class="editor-content">
           <el-input 
              v-model="fileContent" 
              type="textarea" 
              autosize 
              :rows="15" 
              spellcheck="false" 
              class="custom-editor" />
        </div>
        <el-empty v-else description="Select a file to edit" />
      </el-card>
    </div>

    <div v-else class="center-placeholder card">
      <el-empty description="Please select a server to browse files" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from "vue";
import { 
  Folder, Document, Refresh, ArrowLeft, Check 
} from "@element-plus/icons-vue";
import { ElMessage } from "element-plus";
import api from "../api";
import { useServersStore } from "../stores/servers";

const store = useServersStore();
const selectedServerId = ref("");
const files = ref([]);
const currentPath = ref("");
const loading = ref(false);
const editorLoading = ref(false);
const saving = ref(false);
const editingFile = ref("");
const fileContent = ref("");

const servers = computed(() => store.servers);

async function fetchFiles() {
  if (!selectedServerId.value) return;
  loading.value = true;
  try {
    const { data } = await api.listFiles(selectedServerId.value, currentPath.value);
    files.value = data.data || [];
  } catch (e) {
    ElMessage.error("Failed to list files: " + (e.response?.data?.message || e.message));
  } finally {
    loading.value = false;
  }
}

async function handleFileClick(row) {
  if (row.is_dir) {
    currentPath.value = row.path;
    fetchFiles();
  } else {
    loadFile(row.path);
  }
}

async function loadFile(path) {
  editingFile.value = path;
  editorLoading.value = true;
  try {
    const { data } = await api.readFile(selectedServerId.value, path);
    fileContent.value = data.data;
  } catch (e) {
    ElMessage.error("Failed to load file: " + (e.response?.data?.message || e.message));
  } finally {
    editorLoading.value = false;
  }
}

async function saveFile() {
  saving.value = true;
  try {
    await api.writeFile(selectedServerId.value, editingFile.value, fileContent.value);
    ElMessage.success("File saved successfully");
  } catch (e) {
    ElMessage.error("Failed to save file: " + (e.response?.data?.message || e.message));
  } finally {
    saving.value = false;
  }
}

function goUp() {
  const parts = currentPath.value.split('/');
  parts.pop();
  currentPath.value = parts.join('/');
  fetchFiles();
}

function formatSize(bytes) {
  if (bytes < 1024) return bytes + " B";
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + " KB";
  return (bytes / 1048576).toFixed(1) + " MB";
}

onMounted(async () => {
  if (store.servers.length === 0) {
    await store.fetchServers();
  }
});
</script>

<style scoped>
.files-layout {
  display: grid;
  grid-template-columns: 400px 1fr;
  gap: 20px;
  height: calc(100vh - 200px);
}
.file-tree-card {
  overflow-y: auto;
}
.file-row {
  cursor: pointer;
}
.file-name {
  display: flex;
  align-items: center;
  gap: 8px;
}
.editing-title {
  font-family: monospace;
  font-size: 13px;
}
.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.editor-card {
  display: flex;
  flex-direction: column;
}
.editor-content {
  flex: 1;
  overflow-y: auto;
}
:deep(.custom-editor textarea) {
  font-family: 'JetBrains Mono', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.6;
}
.breadcrumb-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  padding: 8px;
  background: var(--color-bg);
  border-radius: 4px;
}
.center-placeholder {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}
</style>
