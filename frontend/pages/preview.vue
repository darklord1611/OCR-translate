<template>
  <div class="pt-24 min-h-screen" style="background-image: url('/background.jpg');">

    <div class="controls flex justify-center gap-x-4 gap-y-4 h-24 flex-wrap">
      <select class="select select-bordered w-24 text-primary">
        <option>A4</option>
        <option>Letter</option>
      </select>
      <select class="select select-bordered w-24 text-primary">
        <option>Auto</option>
        <option>Portrait</option>
        <option>Landscape</option>
      </select>
      <select class="select select-bordered w-36 text-primary">
        <option>Small Margin</option>
        <option>No Margin</option>
        <option>Large Margin</option>
      </select>
      <button v-if="images.length === 0" class="btn btn-primary btn-disabled">Convert</button>
      <button v-else class="btn btn-primary" @click="convertImage">Convert</button>

    </div>

    <div class="flex justify-center flex-wrap gap-8 p-8">
      <div v-for="(image, index) in images" :key="index" class="flex justify-center flex-col items-center">

        <img :src="image.url" alt="Uploaded Image"
          class="h-72 w-full image-preview shadow-lg rounded-md object-cover transition-transform duration-300 ease-in-out hover:scale-105 hover:shadow-2xl m-2">
        <div
          class="buttons absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200 flex space-x-2">
          <button class="btn btn-secondary btn-xs" @click="zoomImage(image.url)">
            <i class="fas fa-search-plus"></i> <!-- Zoom Icon -->
          </button>
          <button class="btn btn-secondary btn-xs" @click="removeImage(index)">
            <i class="fas fa-trash-alt"></i> <!-- Delete Icon -->
          </button>
        </div>

        <p class="text-primary overflow-hidden whitespace-wrap text-ellipsis">{{ image.name }}</p>
      </div>

      <div
        class="h-72 w-72 flex items-center justify-center cursor-pointer 
        text-primary hover:text-secondary transition duration-200 border-2 border-dashed border-primary hover:border-secondary rounded-md"
        @click="triggerFileInput">
        <div class="flex flex-col justify-center items-center">
          <i class="fa-solid fa-image text-4xl"></i>
          <span class="text-center text-xl font-bold mt-2"><i class="fa-solid fa-plus"></i> Add Image</span>
        </div>
        <input type="file" class="hidden" ref="fileInput" @change="handleFileUpload">
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">

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

<style scoped></style>