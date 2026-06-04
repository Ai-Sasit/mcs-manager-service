<template>
  <div class="placeholder-view">
    <div class="placeholder-content">
      <PhDesktop :size="48" color="var(--color-text-muted)" />
      <h2>{{ title }}</h2>
      <p>{{ description }}</p>
      <el-button type="primary">
        {{ actionText }}
      </el-button>
    </div>
  </div>
</template>

<script setup>
import { PhDesktop } from "@phosphor-icons/vue";

defineProps({
  title: {
    type: String,
    default: "Coming Soon",
  },
  description: {
    type: String,
    default: "This feature is under development.",
  },
  actionText: {
    type: String,
    default: "Learn More",
  },
});
</script>

<style scoped>
.placeholder-view {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
  padding: 48px;
}

.placeholder-content {
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  max-width: 400px;
}

.placeholder-content h2 {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  color: var(--color-text);
}

.placeholder-content p {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 15px;
  line-height: 1.6;
}
</style>
