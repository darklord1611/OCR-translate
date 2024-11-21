<template>
  <div class="container mx-auto py-10 flex justify-center items-center min-h-screen">
    <nuxt-link to="hhttps://github.com/darklord1611/OCR-translate" class="btn btn-primary m-6">About Us</nuxt-link>
    <nuxt-link to="/preview" class="btn btn-secondary m-6">Get Start</nuxt-link>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';

const fileInput = ref<HTMLInputElement | null>(null);
const router = useRouter();

const triggerFileInput = () => {
  fileInput.value?.click();
};

const handleFileUpload = async (event: Event) => {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  if (!file) return;

  // Pass the file data as a URL object to the preview page
  router.push({
    name: 'preview',
    query: { fileUrl: URL.createObjectURL(file), fileName: file.name }
  });
};
</script>

<style scoped>
/* Styles similar to Smallpdf */
.container {
  max-width: 800px;
}
.card {
  background: #fff;
  border-radius: 8px;
}
.upload-area {
  transition: background-color 0.3s ease;
  min-height: 200px; /* Increase the minimum height */
}
.upload-area:hover {
  background-color: #f3f4f6;
}
</style>
