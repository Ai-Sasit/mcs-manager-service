<template>
  <div class="firewall-view">
    <!-- Status Card -->
    <el-card class="status-card" shadow="never">
      <div class="status-row">
        <div class="status-info">
          <h3>UFW Firewall</h3>
          <template v-if="!available">
            <el-tag type="info" size="large">Not Available</el-tag>
            <span class="status-hint">{{ unavailableMessage }}</span>
          </template>
          <template v-else>
            <el-tag :type="enabled ? 'success' : 'danger'" size="large">
              {{ enabled ? "Active" : "Inactive" }}
            </el-tag>
          </template>
        </div>
        <div class="status-actions" v-if="available">
          <el-button
            :type="enabled ? 'danger' : 'success'"
            @click="handleToggle"
            :loading="toggling">
            {{ enabled ? "Disable UFW" : "Enable UFW" }}
          </el-button>
          <el-button type="primary" @click="showAddDialog = true">
            <el-icon><Plus /></el-icon>
            Add Rule
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- Quick Add -->
    <el-card v-if="available" class="quick-card" shadow="never">
      <template #header>
        <span>Quick Add Common Ports</span>
      </template>
      <div class="quick-buttons">
        <el-button @click="quickAdd(25565, 'tcp')"
          >Minecraft Java (25565/tcp)</el-button
        >
        <el-button @click="quickAdd(19132, 'udp')"
          >Minecraft Bedrock (19132/udp)</el-button
        >
        <el-button @click="quickAdd(22, 'tcp')">SSH (22/tcp)</el-button>
        <el-button @click="quickAdd(80, 'tcp')">HTTP (80/tcp)</el-button>
        <el-button @click="quickAdd(443, 'tcp')">HTTPS (443/tcp)</el-button>
        <el-button @click="quickAdd(8080, 'tcp')">Panel (8080/tcp)</el-button>
      </div>
    </el-card>

    <!-- Rules Table -->
    <el-card v-if="available" shadow="never">
      <template #header>
        <div class="card-header">
          <span>Firewall Rules</span>
          <el-button text @click="loadRules" :loading="loadingRules">
            <el-icon><Refresh /></el-icon>
            Refresh
          </el-button>
        </div>
      </template>

      <el-table
        :data="rules"
        v-loading="loadingRules"
        empty-text="No rules configured"
        stripe>
        <el-table-column prop="number" label="#" width="60" />
        <el-table-column prop="to" label="To (Port/Protocol)" min-width="200" />
        <el-table-column prop="action" label="Action" width="120">
          <template #default="{ row }">
            <el-tag
              :type="
                row.action === 'ALLOW'
                  ? 'success'
                  : row.action === 'DENY'
                    ? 'danger'
                    : 'warning'
              "
              size="small">
              {{ row.action }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="from" label="From" min-width="150" />
        <el-table-column label="Actions" width="100" fixed="right">
          <template #default="{ row }">
            <el-popconfirm
              title="Delete this rule?"
              confirm-button-text="Delete"
              cancel-button-text="Cancel"
              @confirm="handleDeleteRule(row.number)">
              <template #reference>
                <el-button type="danger" size="small" text>
                  <el-icon><Delete /></el-icon>
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Add Rule Dialog -->
    <el-dialog
      v-model="showAddDialog"
      title="Add Firewall Rule"
      width="420px"
      @close="resetForm">
      <el-form :model="form" label-position="top">
        <el-form-item label="Port" required>
          <el-input-number
            v-model="form.port"
            :min="1"
            :max="65535"
            :step="1"
            controls-position="right"
            style="width: 100%" />
        </el-form-item>
        <el-form-item label="Protocol">
          <el-radio-group v-model="form.protocol">
            <el-radio-button value="">Both</el-radio-button>
            <el-radio-button value="tcp">TCP</el-radio-button>
            <el-radio-button value="udp">UDP</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="Action">
          <el-radio-group v-model="form.action">
            <el-radio-button value="allow">Allow</el-radio-button>
            <el-radio-button value="deny">Deny</el-radio-button>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">Cancel</el-button>
        <el-button type="primary" @click="handleAddRule" :loading="adding"
          >Add Rule</el-button
        >
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { Plus, Delete, Refresh } from "@element-plus/icons-vue";
import { ElMessage } from "element-plus";
import api from "../api";

const available = ref(false);
const enabled = ref(false);
const unavailableMessage = ref("");
const rules = ref([]);
const loadingRules = ref(false);
const toggling = ref(false);
const adding = ref(false);
const showAddDialog = ref(false);

const form = ref({
  port: 25565,
  protocol: "tcp",
  action: "allow",
});

function resetForm() {
  form.value = { port: 25565, protocol: "tcp", action: "allow" };
}

async function loadStatus() {
  try {
    const { data } = await api.getUfwStatus();
    available.value = data.available;
    enabled.value = data.enabled || false;
    if (!data.available) {
      unavailableMessage.value =
        data.message || "UFW is not available on this system";
    }
  } catch {
    available.value = false;
    unavailableMessage.value = "Failed to check UFW status";
  }
}

async function loadRules() {
  loadingRules.value = true;
  try {
    const { data } = await api.getUfwRules();
    rules.value = data.data || [];
  } catch {
    ElMessage.error("Failed to load firewall rules");
  } finally {
    loadingRules.value = false;
  }
}

async function handleToggle() {
  toggling.value = true;
  try {
    const { data } = await api.toggleUfw(!enabled.value);
    ElMessage.success(
      data.message || (enabled.value ? "UFW disabled" : "UFW enabled"),
    );
    await loadStatus();
    await loadRules();
  } catch {
    ElMessage.error("Failed to toggle UFW");
  } finally {
    toggling.value = false;
  }
}

async function handleAddRule() {
  if (!form.value.port || form.value.port < 1 || form.value.port > 65535) {
    ElMessage.warning("Please enter a valid port (1-65535)");
    return;
  }
  adding.value = true;
  try {
    const payload = { port: form.value.port, protocol: form.value.protocol };
    const apiFn =
      form.value.action === "allow" ? api.allowUfwRule : api.denyUfwRule;
    const { data } = await apiFn(payload);
    ElMessage.success(data.message || "Rule added");
    showAddDialog.value = false;
    resetForm();
    await loadRules();
  } catch {
    ElMessage.error("Failed to add rule");
  } finally {
    adding.value = false;
  }
}

async function handleDeleteRule(number) {
  try {
    const { data } = await api.deleteUfwRule(number);
    ElMessage.success(data.message || "Rule deleted");
    await loadRules();
  } catch {
    ElMessage.error("Failed to delete rule");
  }
}

async function quickAdd(port, protocol) {
  try {
    const { data } = await api.allowUfwRule({ port, protocol });
    ElMessage.success(data.message || `Allowed ${port}/${protocol}`);
    await loadRules();
  } catch {
    ElMessage.error("Failed to add rule");
  }
}

onMounted(async () => {
  await loadStatus();
  if (available.value) {
    await loadRules();
  }
});
</script>

<style scoped>
.firewall-view {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.status-card .status-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.status-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.status-info h3 {
  margin: 0;
  font-size: 18px;
}

.status-hint {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.status-actions {
  display: flex;
  gap: 8px;
}

.quick-card .quick-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
</style>
