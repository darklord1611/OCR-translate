<template>
  <div class="pt-24 min-h-screen" style="background-image: url('/background.jpg');">
    <div class="text-center">
      <h2 class="text-2xl font-bold text-primary mb-4">Conversion Result</h2>

      <div v-if="afterImageUrls.length">
        <div class="flex justify-center flex-wrap gap-x-6">
          <button @click="downloadResult" class="btn btn-primary"> <i class="fa-solid fa-download"></i> Download All
            PDFs</button>
          <nuxt-link to="/preview" class="btn btn-secondary"><i class="fa-solid fa-rotate-left"></i> Start
            Over</nuxt-link>
        </div>
  
        <div>
          <div class="flex justify-center flex-wrap gap-8 p-8">
            <iframe v-for="(url, index) in afterImageUrls" :key="index" :src="url" frameborder="0"
              class="w-96 h-72 border-2 border-gray-300 shadow-lg rounded-md transition-transform 
              duration-300 ease-in-out hover:scale-105 hover:shadow-2xl m-2"></iframe>
          </div>
        </div>
      </div>

      <div v-else>
        <p class="text-lg text-secondary">Your PDFs are being generated. Please wait ...</p>
        <i class="fa-solid fa-spinner fa-spin text-secondary text-4xl p-4"></i>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">

const route = useRoute();
const afterImageUrls = ref<string[]>([]);
const jobStatus = ref<string>('pending');
const jobIDs = ref<string[]>([]);

const convertImagesToPDFs = async () => {
  const fileUrls = route.query.images as string[];

  for (const [index, url] of fileUrls.entries()) {
    const formData = new FormData();
    const fileBlob = await fetch(url).then(res => res.blob());
    formData.append(`file`, fileBlob, `image${index}.png`);

    try {
      const response = await fetch(`${import.meta.env.VITE_BACKEND_URL}/upload`, {
        method: 'POST',
        body: formData,
      });
      const data = await response.json();
      jobIDs.value.push(data.jobID);
    } catch (error) {
      console.error(`Error uploading file ${index + 1}:`, error);
    }
  }

  // Start polling job status for each job
  jobIDs.value.forEach(pollJobStatus);
};

const pollJobStatus = async (jobID: string) => {
  const interval = setInterval(async () => {
    try {
      const response = await fetch(`${import.meta.env.VITE_BACKEND_URL}/status/${jobID}`);
      const data = await response.json();
      const status = data.status;

      if (status === 'completed') {
        clearInterval(interval);
        afterImageUrls.value.push(`${import.meta.env.VITE_BACKEND_URL}/uploads/${jobID}.pdf`);
      }
    } catch (error) {
      console.error('Error checking job status:', error);
    }
  }, 2000);
};

const downloadResult = () => {
  jobIDs.value.forEach((jobID) => {
    const link = document.createElement('a');
    link.href = `${import.meta.env.VITE_BACKEND_URL}/download/${jobID}.pdf`;
    link.download = `${jobID}.pdf`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  })
};

onMounted(() => {
  convertImagesToPDFs();
});
</script>

<style scoped></style>
