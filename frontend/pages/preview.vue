<template>
  <div class="pt-24 min-h-screen" style="background-image: url('/background.jpg');">
    <div class="flex justify-center gap-4 h-24 flex-wrap">
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
      <button :disabled="images.length==0" class="btn btn-primary" @click="handleConvertClick">
        Convert
      </button>
    </div>

    <div class="flex justify-center flex-wrap gap-8 p-8">
      <div v-for="(image, index) in images" :key="index" class="flex justify-center flex-col items-center group relative">
        <img :src="image.url" alt="Uploaded Image"
          class="h-72 w-full shadow-lg rounded-md object-cover transition-transform duration-300 ease-in-out hover:scale-105 hover:shadow-2xl m-2">
        <div class="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200 flex space-x-2">
          <button class="btn btn-secondary btn-sm" @click="zoomImage(image.url)">
            <i class="fas fa-search-plus"></i>
          </button>
          <button class="btn btn-secondary btn-sm" @click="removeImage(index)">
            <i class="fas fa-trash-alt"></i>
          </button>
        </div>
        <p class="text-primary overflow-hidden text-ellipsis">{{ image.name }}</p>
      </div>

      <div v-if="images.length < 10" class="h-72 w-72 flex items-center justify-center cursor-pointer text-primary hover:text-secondary transition duration-200 border-2 border-dashed border-primary hover:border-secondary rounded-md" @click="triggerFileInput">
        <div class="flex flex-col justify-center items-center">
          <i class="fa-solid fa-image text-4xl"></i>
          <span class="text-center text-xl font-bold mt-2"><i class="fa-solid fa-plus"></i> Add Image</span>
        </div>
        <input type="file" class="hidden" ref="fileInput" @change="handleFileUpload" multiple>
      </div>
    </div>

    <div v-if="showLimitMessage" class="toast toast-end">
      <div class="alert alert-error">
        <span>Bạn chỉ có thể gửi tối đa 10 hình ảnh!</span>
      </div>
    </div>

    <div v-if="showConversionLimitMessage" class="toast toast-end">
      <div class="alert alert-error">
        <span>Bạn chỉ có thể gửi tối đa 3 yêu cầu trong 5 phút, xin hãy thử lại sau.</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';

const router = useRouter();
const fileInput = ref<HTMLInputElement | null>(null);
const images = ref<{ url: string; name: string }[]>([]);
const showLimitMessage = ref(false);
const showConversionLimitMessage = ref(false);
const conversionCount = ref(0);
const lastConversionTime = ref<Date | null>(null);
const isConvertDisabled = ref(false);

const removeImage = (index: number) => {
  images.value.splice(index, 1);
};

const zoomImage = (url: string) => {
  window.open(url, '_blank');
};

const handleConvertClick = () => {
  if (isConvertDisabled.value) {
    showConversionLimitMessage.value = true;
    setTimeout(() => {
      showConversionLimitMessage.value = false;
    }, 3000);
    console.log('Conversion limit reached');
  } else {
    convertImage();
  }
};

const convertImage = () => {
  const now = new Date();
  if (lastConversionTime.value && (now.getTime() - lastConversionTime.value.getTime()) > 5 * 60 * 1000) {
    conversionCount.value = 0;
  }

  if (conversionCount.value < 3) {
    conversionCount.value++;
    lastConversionTime.value = now;
    localStorage.setItem('conversionCount', conversionCount.value.toString());
    localStorage.setItem('lastConversionTime', now.toString());
    router.push({
      name: 'results',
      query: { images: images.value.map(image => image.url) },
    });
  } else {
    showConversionLimitMessage.value = true;
    setTimeout(() => {
      showConversionLimitMessage.value = false;
    }, 3000);
  }

  if (conversionCount.value >= 3) {
    isConvertDisabled.value = true;
    setTimeout(() => {
      isConvertDisabled.value = false;
      conversionCount.value = 0;
      localStorage.removeItem('conversionCount');
      localStorage.removeItem('lastConversionTime');
    }, 5 * 60 * 1000);
  }
};

const triggerFileInput = () => {
  fileInput.value?.click();
};

const handleFileUpload = (event: Event) => {
  const input = event.target as HTMLInputElement;
  if (input.files) {
    const remainingSlots = 10 - images.value.length;
    const filesToAdd = Array.from(input.files).slice(0, remainingSlots);
    if (input.files.length > remainingSlots) {
      showLimitMessage.value = true;
      setTimeout(() => {
        showLimitMessage.value = false;
      }, 3000);
    }
    for (const file of filesToAdd) {
      const url = URL.createObjectURL(file);
      const name = file.name;
      images.value.push({ url, name });
    }
  }
};

onMounted(() => {
  const storedConversionCount = localStorage.getItem('conversionCount');
  const storedLastConversionTime = localStorage.getItem('lastConversionTime');

  if (storedConversionCount) {
    conversionCount.value = parseInt(storedConversionCount, 10);
  }

  if (storedLastConversionTime) {
    lastConversionTime.value = new Date(storedLastConversionTime);
    const now = new Date();
    if ((now.getTime() - lastConversionTime.value.getTime()) > 5 * 60 * 1000) {
      conversionCount.value = 0;
      localStorage.removeItem('conversionCount');
      localStorage.removeItem('lastConversionTime');
    } else if (conversionCount.value >= 3) {
      isConvertDisabled.value = true;
      setTimeout(() => {
        isConvertDisabled.value = false;
        conversionCount.value = 0;
        localStorage.removeItem('conversionCount');
        localStorage.removeItem('lastConversionTime');
      }, 5 * 60 * 1000 - (now.getTime() - lastConversionTime.value.getTime()));
    }
  }
});
</script>

<style scoped></style>