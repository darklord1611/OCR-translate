<template>
  <div class="container mx-auto pt-12 flex">
    <div class="w-full">
      <div class="navbar controls flex justify-between items-center rounded-lg mb-4 w-full">
        <div class="flex items-center space-x-4">
          <select class="select select-bordered w-24">
            <option>A4</option>
            <option>Letter</option>
          </select>
          <select class="select select-bordered w-24">
            <option>Auto</option>
            <option>Portrait</option>
            <option>Landscape</option>
          </select>
          <select class="select select-bordered w-36">
            <option>Small Margin</option>
            <option>No Margin</option>
            <option>Large Margin</option>
          </select>
        </div>
        <button v-if="images.length === 0" class="btn" >Convert</button>
        <button v-else class="btn btn-primary" @click="convertImage">Convert</button>

      </div>

      <div class="images-container grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-y-4">
        <div
          v-for="(image, index) in images"
          :key="index"
          class="file-preview relative group w-48"
          :class="{'justify-self-center': 'true'}"
        >
          <img :src="image.url" alt="Uploaded Image" class="image-preview shadow-lg rounded-md object-cover transition duration-200">

          <div class="buttons absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200 flex space-x-2">
            <button class="btn btn-secondary btn-xs" @click="zoomImage(image.url)">
              <i class="fas fa-search-plus"></i> <!-- Zoom Icon -->
            </button>
            <button class="btn btn-secondary btn-xs" @click="removeImage(index)">
              <i class="fas fa-trash-alt"></i> <!-- Delete Icon -->
            </button>

          </div>

          <p class="image-name text-center text-gray-600 mt-2">{{ image.name }}</p>
        </div>

        <div class="add-image-button file-preview relative group w-48 flex items-center justify-center cursor-pointer text-gray-500 hover:bg-gray-100 transition duration-200 border-2 border-dashed border-gray-300 rounded-md justify-self-center" @click="triggerFileInput">
          <span class="text-center">+ Add Image</span>
          <input type="file" class="hidden" ref="fileInput" @change="handleFileUpload">
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';

const router = useRouter();
const fileInput = ref<HTMLInputElement | null>(null);
const images = ref<{ url: string; name: string }[]>([]);

const removeImage = (index: number) => {
  images.value.splice(index, 1);
};

const zoomImage = (url: string) => {
  window.open(url, '_blank'); 
};

const convertImage = () => {
  router.push({
    name: 'results',
    query: { images: images.value.map(image => image.url) },
  });
};

const triggerFileInput = () => {
  fileInput.value?.click();
};

const handleFileUpload = (event: Event) => {
  const input = event.target as HTMLInputElement;
  if (input.files && input.files[0]) {
    const file = input.files[0];
    const url = URL.createObjectURL(file);
    const name = file.name;
    images.value.push({ url, name });
  }
};
</script>

<style scoped>
/* Controls Section */
.controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border-radius: 0.5rem;
  box-shadow: 3px 3px 20px rgba(170, 74, 74, 0.35);
  margin-bottom: 1rem;
  margin-top: 2rem;
  width: 100%;
}

.controls .flex {
  width: 100%;
}

.controls select {
  font-weight: bold;
  width: 9rem;
}

.controls button {
  width: auto;
}

.images-container {
  display: grid;
  gap: 1rem;
}

.file-preview {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 12rem;
}

.image-preview {
  width: 12rem;
  height: 12rem;
  object-fit: cover;
  transition: box-shadow 0.3s ease, transform 0.3s ease;
}

.image-preview:hover {
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  transform: scale(1.02);
}

.buttons {
  position: absolute;
  top: 8px;
  right: 8px;
}

.image-name {
  width: 100%;
  text-align: center;
  font-weight: 500;
  margin-top: 0.5rem;
}

.add-image-button {
  width: 12rem;
  height: 12rem;
}

</style>
