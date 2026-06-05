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
        multiple
        @change="handleFiles"
        :accept="edition === 'java' ? '.jar' : '.mcpack,.mcaddon,.zip'"
        style="display: none"
      />
      <div class="dropzone-content">
        <PhUpload :size="36" class="drop-icon" />
        <p class="drop-text">
          Drop {{ edition === "java" ? "plugin (.jar)" : "addon" }} files here
        </p>
        <el-button size="small" @click="$refs.fileInput.click()">
          Browse Files
        </el-button>
      </div>
    </div>

    <!-- Upload queue -->
    <div v-if="queue.length > 0" class="upload-queue">
      <div
        v-for="(item, idx) in queue"
        :key="idx"
        class="queue-row"
        :class="item.status"
      >
        <div class="queue-info">
          <span class="queue-name">{{ item.file.name }}</span>
          <span class="queue-size">{{ formatSize(item.file.size) }}</span>
        </div>
        <div class="queue-progress">
          <div class="progress-bar-track">
            <div
              class="progress-bar-fill"
              :class="item.status"
              :style="{ width: item.percent + '%' }"
            ></div>
          </div>
          <span class="progress-text">
            <template v-if="item.status === 'uploading'">
              {{ item.percent }}%
            </template>
            <template v-else-if="item.status === 'done'">
              <PhCheck :size="14" class="status-icon" />
            </template>
            <template v-else-if="item.status === 'error'">
              <PhX :size="14" class="status-icon" />
              <span class="error-msg">{{ item.error }}</span>
            </template>
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, nextTick } from "vue";
import { PhUpload, PhCheck, PhX } from "@phosphor-icons/vue";
import apiClient from "@/api/client";
import { getApiErrorMessage } from "@/utils/apiError";

const props = defineProps({
  serverId: String,
  edition: String,
});

const emit = defineEmits(["uploaded"]);

const dragging = ref(false);
const fileInput = ref(null);
const queue = ref([]);

function formatSize(bytes) {
  if (!bytes) return "0 B";
  if (bytes < 1024) return bytes + " B";
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + " KB";
  return (bytes / 1048576).toFixed(1) + " MB";
}

function enqueueFiles(files) {
  const newItems = Array.from(files).map((file) => ({
    file,
    percent: 0,
    status: "pending", // pending | uploading | done | error
    error: "",
  }));
  queue.value.push(...newItems);
  nextTick(() => processQueue());
}

let processing = false;
async function processQueue() {
  if (processing) return;
  processing = true;

  while (queue.value.some((q) => q.status === "pending")) {
    const item = queue.value.find((q) => q.status === "pending");
    if (!item) break;

    item.status = "uploading";
    item.percent = 0;

    try {
      const fd = new FormData();
      fd.append("file", item.file);

      await apiClient.post(
        `/servers/${encodeURIComponent(props.serverId)}/plugins`,
        fd,
        {
          onUploadProgress(progressEvent) {
            if (progressEvent.total) {
              item.percent = Math.round(
                (progressEvent.loaded * 100) / progressEvent.total,
              );
            }
          },
        },
      );

      item.status = "done";
      item.percent = 100;
      emit("uploaded");
    } catch (e) {
      item.status = "error";
      item.error = getApiErrorMessage(e);
    }
  }

  processing = false;

  // Clear queue if everything is done/failed
  if (queue.value.every((q) => q.status === "done" || q.status === "error")) {
    setTimeout(() => {
      queue.value = [];
    }, 4000);
  }
}

function handleFiles(e) {
  enqueueFiles(e.target.files);
  if (fileInput.value) fileInput.value.value = "";
}

function handleDrop(e) {
  dragging.value = false;
  enqueueFiles(e.dataTransfer.files);
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

.upload-queue {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.queue-row {
  background: var(--color-white);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 12px 16px;
}
.queue-row.error {
  border-color: rgba(239, 68, 68, 0.3);
  background: rgba(239, 68, 68, 0.03);
}
.queue-row.done {
  border-color: rgba(34, 197, 94, 0.3);
  background: rgba(34, 197, 94, 0.03);
}
.queue-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}
.queue-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 70%;
}
.queue-size {
  font-size: 12px;
  color: var(--color-text-muted);
  flex-shrink: 0;
}
.queue-progress {
  display: flex;
  align-items: center;
  gap: 10px;
}
.progress-bar-track {
  flex: 1;
  height: 6px;
  background: var(--color-bg-secondary, #f1f5f9);
  border-radius: 3px;
  overflow: hidden;
}
.progress-bar-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.3s ease;
  background: var(--color-primary, #10b981);
}
.progress-bar-fill.error {
  background: #ef4444;
}
.progress-bar-fill.done {
  background: #22c55e;
}
.progress-text {
  font-size: 12px;
  font-weight: 500;
  color: var(--color-text-muted);
  min-width: 60px;
  display: flex;
  align-items: center;
  gap: 4px;
  justify-content: flex-end;
}
.status-icon {
  flex-shrink: 0;
}
.done .status-icon {
  color: #22c55e;
}
.error .status-icon {
  color: #ef4444;
}
.error-msg {
  color: #ef4444;
  font-size: 11px;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
