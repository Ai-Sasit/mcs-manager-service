<template>
  <div class="placeholder-view">
    <el-card>
      <div class="empty-state">
        <el-icon size="48" color="var(--color-text-muted)"><Monitor /></el-icon>
        <h2>{{ title }}</h2>
        <p>This module is currently under development. Check back soon for updates!</p>
        <el-button type="primary" @click="$router.push('/')">Return to Dashboard</el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { Monitor } from "@element-plus/icons-vue";
defineProps({
  title: {
    type: String,
    default: "Coming Soon",
  },
});
</script>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
  gap: 16px;
}

.empty-state h2 {
  margin: 0;
  font-size: 24px;
  color: var(--color-text);
}

.empty-state p {
  margin: 0;
  color: var(--color-text-secondary);
  max-width: 400px;
  line-height: 1.6;
}
</style>
