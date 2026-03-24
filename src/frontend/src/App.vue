<script setup lang="ts">
import { Toaster } from 'vue-sonner'
import 'vue-sonner/style.css'
</script>

<template>
  <Toaster rich-colors close-button position="bottom-right" />
  <!-- component matched by the route will render here -->
  <router-view />
</template>
