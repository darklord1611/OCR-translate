<template>
  <div class="hero bg-white min-h-screen" style="background-image: url('/background.jpg');">
    <div class="hero-content flex-col lg:flex-row">
      <img src="/landing.svg" alt="OCR Translate" class="transition-transform 
            duration-300 ease-in-out hover:scale-105 m-2" />
      <div class="p-12">
        <h1 class="text-5xl font-bold text-red-500 transition-transform 
            duration-300 ease-in-out hover:scale-105 m-2">OCR - Translator</h1>
        <p class="py-6">
          Press the button below to get started with OCR Translate. Upload an image containing text and we'll convert it
          to text for you.
        </p>
        <div class="flex space-x-6">
          <nuxt-link to="/preview" class="btn btn-primary"><i 
            class="fa-solid fa-rocket"></i> Get Started</nuxt-link>
          <nuxt-link to="https://github.com/darklord1611/OCR-translate" class="btn btn-secondary"> <i
              class="fa-brands fa-github"></i> About Us</nuxt-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">

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

<style scoped></style>
