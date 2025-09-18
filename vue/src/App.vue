<template>
  <div id="app" style="min-height: 100vh; width: 100%;">
    <router-view></router-view>
  </div>
</template>

<script setup>
import { createApp } from 'vue';
import router from './router';

const app = createApp({
  router,
});

app.mount('#app');
</script>


