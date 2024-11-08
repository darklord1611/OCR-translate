<template>
  <div class="container mx-auto p-4">
    <div class="upload-area text-center p-6 border border-dashed rounded-lg cursor-pointer" @click="triggerFileInput">
      <i class="fa fa-cloud-upload-alt text-4xl text-primary mb-4"></i>
      <p class="text-lg">Click to select or drag and drop a file</p>
      <input ref="fileInput" type="file" class="hidden" @change="handleFileUpload" />
    </div>
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

  // Navigate to the preview page and pass the file information
  router.push({ path: `/preview` });
};
</script>

<style scoped>
.upload-area {
  transition: background-color 0.3s ease;
}
.upload-area:hover {
  background-color: #f3f4f6;
}
</style>
