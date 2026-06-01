<template>
  <header class="header">
    <div class="header-inner">
      <router-link to="/" class="logo">
        <PhCube :size="24" weight="duotone" class="logo-icon" />
        <span class="logo-text">MC Manage</span>
      </router-link>
      <nav class="nav">
        <router-link to="/" class="nav-link">
          <PhSquaresFour :size="18" weight="duotone" />
          Dashboard
        </router-link>
      </nav>
    </div>
  </header>
</template>

<script setup>
import { PhCube, PhSquaresFour } from "@phosphor-icons/vue";
</script>

<style scoped>
.header {
  position: sticky;
  top: 0;
  z-index: 100;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border);
  padding: 0 24px;
}

.header-inner {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
  color: var(--text-primary);
}

.logo-icon {
  color: var(--accent);
}

.logo-text {
  font-size: 1.1rem;
  font-weight: 800;
  letter-spacing: 0;
}

.nav {
  display: flex;
  gap: 4px;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border-radius: var(--radius-xs);
  text-decoration: none;
  color: var(--text-secondary);
  font-weight: 500;
  font-size: 0.875rem;
  transition: var(--transition);
}

.nav-link:hover,
.nav-link.router-link-active {
  background: var(--accent-pale);
  color: var(--accent);
}
</style>
