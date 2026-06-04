<template>
  <div class="users-page">
    <div class="page-toolbar card">
      <div class="toolbar-left">
        <h3>Panel Users</h3>
        <el-tag size="small" type="info">{{ users.length }} user(s)</el-tag>
      </div>
      <el-button type="primary" @click="openCreate"
        ><template #icon><PhPlus /></template
        >Add User</el-button
      >
    </div>
    <div class="card" style="padding: 24px">
      <el-table :data="users" style="width: 100%" v-loading="loading">
        <el-table-column label="Username" prop="username" />
        <el-table-column label="Role" width="140">
          <template #default="{ row }">
            <el-tag
              :type="row.role === 'admin' ? 'danger' : 'info'"
              effect="light"
              size="small"
              >{{ row.role }}</el-tag
            >
          </template>
        </el-table-column>
        <el-table-column label="Created" width="200">
          <template #default="{ row }">{{
            formatDate(row.created_at)
          }}</template>
        </el-table-column>
        <el-table-column label="Actions" width="160" align="center">
          <template #default="{ row }">
            <el-button
              size="small"
              circle
              @click="openEdit(row)"
              ><template #icon><PhPencilSimple /></template
            />
            <el-button
              size="small"
              type="danger"
              circle
              ><template #icon><PhTrash /></template
              @click="handleDelete(row)"
              :disabled="row.username === currentUsername"
            />
          </template>
        </el-table-column>
      </el-table>
    </div>
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? 'Edit User' : 'Create User'"
      width="420px"
      :close-on-click-modal="false"
    >
      <el-form :model="form" label-position="top">
        <el-form-item label="Username"
          ><el-input
            v-model="form.username"
            placeholder="Enter username"
            :disabled="isEdit"
        /></el-form-item>
        <el-form-item
          :label="isEdit ? 'New Password (leave blank to keep)' : 'Password'"
        >
          <el-input
            v-model="form.password"
            type="password"
            placeholder="Enter password"
            show-password
          />
        </el-form-item>
        <el-form-item label="Role">
          <el-select v-model="form.role" style="width: 100%">
            <el-option label="Admin" value="admin" />
            <el-option label="Viewer" value="viewer" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">{{
          isEdit ? "Update" : "Create"
        }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { PhPlus, PhPencilSimple, PhTrash } from "@phosphor-icons/vue";
import { ElMessage, ElMessageBox } from "element-plus";
import apiClient from "@/api/client";
import { getApiErrorMessage } from "@/utils/apiError";

const users = ref([]);
const loading = ref(false);
const dialogVisible = ref(false);
const isEdit = ref(false);
const editId = ref("");
const saving = ref(false);
const currentUsername = localStorage.getItem("mc_username") || "admin";
const form = ref({ username: "", password: "", role: "viewer" });

async function loadUsers() {
  loading.value = true;
  try {
    const { data } = await apiClient.get("/users");
    users.value = data.data || [];
  } catch (e) {
    users.value = [];
    ElMessage.error("Failed to load users: " + getApiErrorMessage(e));
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  isEdit.value = false;
  editId.value = "";
  form.value = { username: "", password: "", role: "viewer" };
  dialogVisible.value = true;
}
function openEdit(row) {
  isEdit.value = true;
  editId.value = row.id;
  form.value = { username: row.username, password: "", role: row.role };
  dialogVisible.value = true;
}

async function handleSave() {
  saving.value = true;
  try {
    if (isEdit.value) {
      await apiClient.put(`/users/${encodeURIComponent(editId.value)}`, {
        username: form.value.username,
        password: form.value.password || undefined,
        role: form.value.role,
      });
      ElMessage.success("User updated.");
    } else {
      if (!form.value.username || !form.value.password) {
        ElMessage.warning("Username and password are required.");
        saving.value = false;
        return;
      }
      await apiClient.post("/users", form.value);
      ElMessage.success("User created.");
    }
    dialogVisible.value = false;
    loadUsers();
  } catch (e) {
    ElMessage.error(getApiErrorMessage(e));
  } finally {
    saving.value = false;
  }
}

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(
      `Delete user "${row.username}"? This cannot be undone.`,
      "Confirm Delete",
      {
        confirmButtonText: "Delete",
        cancelButtonText: "Cancel",
        type: "warning",
        confirmButtonClass: "el-button--danger",
      },
    );
    await apiClient.delete(`/users/${encodeURIComponent(row.id)}`);
    ElMessage.success("User deleted.");
    loadUsers();
  } catch (e) {
    if (e !== "cancel") ElMessage.error(getApiErrorMessage(e));
  }
}

function formatDate(iso) {
  if (!iso) return "-";
  return new Date(iso).toLocaleString();
}
onMounted(loadUsers);
</script>

<style scoped>
.users-page {
  animation: fadeIn 0.3s ease-out;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.page-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
}
.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.toolbar-left h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}
</style>
