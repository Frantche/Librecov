<template>
  <div class="project-view">
    <div class="header-section">
      <h2>Project Details</h2>
      <router-link :to="`/projects/${$route.params.id}/settings`" class="btn btn-secondary">
        Project Settings
      </router-link>
    </div>
    <p>Project ID: {{ $route.params.id }}</p>
    <p>This view will show project details, builds, and coverage history.</p>
  </div>
</template>

<script setup lang="ts">
// Project detail view to be implemented
</script>

<style scoped>
.project-view {
  padding: 1rem 0;
}

.header-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
  text-decoration: none;
  display: inline-block;
}

.btn-secondary {
  background: #95a5a6;
  color: white;
}

.btn-secondary:hover {
  background: #7f8c8d;
}
</style>

