<template>
  <div id="app" style="height: 100vh; width: 100%; overflow: hidden;">
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


