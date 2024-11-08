<template>
  <div class="container mx-auto py-10">
    <div class="card p-8 shadow-lg text-center">
      <h2 class="text-2xl font-bold mb-4">Conversion Result</h2>

      <div v-if="afterImageUrls.length">
        <div class="images-result grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 xl:g gap-4 mb-6">
          <iframe
            v-for="(url, index) in afterImageUrls"
            :key="index"
            :src="url"
            class="pdf-viewer"
            frameborder="0"
          ></iframe>
        </div>
        <button @click="downloadResult" class="btn btn-primary m-6">Download All PDFs</button>
        <nuxt-link to="/preview" class="btn btn-secondary m-6">Start Over</nuxt-link>
      </div>
      <div v-else>
        <p class="text-lg">Your PDFs are being generated. Please wait...</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';

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
      const response = await fetch('http://localhost:8080/upload', {
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
      const response = await fetch(`http://localhost:8080/status/${jobID}`);
      const data = await response.json();
      const status = data.status;

      if (status === 'completed') {
        clearInterval(interval);
        afterImageUrls.value.push(`http://localhost:8080/uploads/${jobID}.pdf`);
      }
    } catch (error) {
      console.error('Error checking job status:', error);
    }
  }, 2000);
};

const downloadResult = () => {
  jobIDs.value.forEach((jobID) => {
    const link = document.createElement('a');
    link.href = `http://localhost:8080/download/${jobID}.pdf`;
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

<style scoped>
.container {
  max-width: 800px;
}
.card {
  border-radius: 8px;
}
.pdf-viewer {
  width: 100%;
  height: 300px;
  border: 1px solid #ddd;
}
.images-result {
  display: grid;
  gap: 1rem;
}
.btn {
  padding: 10px 20px;
  border-radius: 4px;
  cursor: pointer;
}
</style>
