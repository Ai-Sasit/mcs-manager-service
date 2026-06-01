<template>
  <div class="upload-area">
    <div
      class="dropzone"
      :class="{ dragging }"
      @dragover.prevent="dragging = true"
      @dragleave="dragging = false"
      @drop.prevent="handleDrop"
    >
      <input
        type="file"
        ref="fileInput"
        @change="handleFile"
        :accept="edition === 'java' ? '.jar' : '.mcpack,.mcaddon,.zip'"
        style="display: none"
      />
      <div class="dropzone-content">
        <el-icon size="36" class="drop-icon"><Upload /></el-icon>
        <p class="drop-text">
          Drop {{ edition === "java" ? "plugin (.jar)" : "addon" }} files here
        </p>
        <el-button size="small" @click="$refs.fileInput.click()">
          Browse Files
        </el-button>
      </div>
    </div>
    <div v-if="uploading" class="upload-status">
      <el-icon class="spin" size="16"><Loading /></el-icon>
      Uploading...
    </div>
    <div v-if="message" class="upload-message" :class="messageType">
      {{ message }}
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { Upload, Loading } from "@element-plus/icons-vue";
import apiClient from "@/api/client";
import { getApiErrorMessage } from "@/utils/apiError";

const props = defineProps({
  serverId: String,
  edition: String,
});

const emit = defineEmits(["uploaded"]);

const dragging = ref(false);
const uploading = ref(false);
const message = ref("");
const messageType = ref("success");
const fileInput = ref(null);

async function uploadFile(file) {
  uploading.value = true;
  message.value = "";
  try {
    const fd = new FormData();
    fd.append("file", file);
    await apiClient.post(
      `/servers/${encodeURIComponent(props.serverId)}/plugins`,
      fd,
      {
        headers: { "Content-Type": "multipart/form-data" },
      },
    );
    message.value = `Uploaded ${file.name}`;
    messageType.value = "success";
    emit("uploaded");
  } catch (e) {
    message.value = `Failed: ${getApiErrorMessage(e)}`;
    messageType.value = "error";
  } finally {
    uploading.value = false;
    if (fileInput.value) fileInput.value.value = "";
  }
}

function handleFile(e) {
  const file = e.target.files[0];
  if (file) uploadFile(file);
}

function handleDrop(e) {
  dragging.value = false;
  const file = e.dataTransfer.files[0];
  if (file) uploadFile(file);
}
</script>

<style scoped>
.upload-area {
  margin-top: 8px;
}
.dropzone {
  border: 2px dashed var(--color-border);
  border-radius: var(--radius-lg);
  padding: 32px;
  text-align: center;
  transition: all 0.2s ease;
  cursor: pointer;
  background: var(--color-white);
}
.dropzone:hover,
.dropzone.dragging {
  border-color: var(--color-primary);
  background: rgba(16, 185, 129, 0.03);
}
.dropzone-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}
.drop-icon {
  color: var(--color-text-muted);
}
.drop-text {
  color: var(--color-text-secondary);
  font-weight: 500;
  font-size: 14px;
}
.upload-status {
  margin-top: 12px;
  text-align: center;
  font-weight: 500;
  color: var(--color-text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 14px;
}
.spin {
  animation: spin 1.2s linear infinite;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
.upload-message {
  margin-top: 10px;
  padding: 10px 14px;
  border-radius: var(--radius);
  font-weight: 500;
  font-size: 13px;
  text-align: center;
}
.upload-message.success {
  background: rgba(34, 197, 94, 0.08);
  color: #16a34a;
}
.upload-message.error {
  background: rgba(239, 68, 68, 0.08);
  color: #dc2626;
}
</style>
