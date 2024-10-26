<template>
    <div class="flex justify-center items-center min-h-screen">
        <div class="grid place-items-center w-full h-full rounded-2xl text-zinc-800">
            <div class="flex flex-col items-center justify-center gap-10vh">
                <div>&nbsp;</div>

                <div v-if="!jobID"
                    class="flex items-center gap-2 max-w-80vw text-lg cursor-pointer border-2 border-dashed border-primary p-8 rounded-md">
                    <i class="fa fa-image text-primary"></i>
                    <input ref="fileInput" class="hidden" type="file" @change="handleFileUpload">
                    <button @click="triggerFileInput" class=" text-primary">Paste an image, click to select, or drag and drop</button>
                </div>
                <div v-if="jobID">
                    <div class="flex w-full flex-col lg:flex-row">
                        <div class="rounded-box grid flex-grow place-items-center flex-1">
                            <img :src="beforeImageUrl" alt="Before Image" class="max-h-full max-w-full object-contain">
                        </div>
                        <div class="divider divider-primary lg:divider-horizontal"></div>
                        <div class="rounded-box grid flex-grow place-items-center flex-1">
                            <template v-if="afterImageUrl.endsWith('.pdf')">
                                <embed :src="afterImageUrl" type="application/pdf" width="100%" height="100%">
                            </template>
                            <template v-else>
                                <i class="fa-solid fa-spinner fa-spin-pulse text-primary"></i>
                            </template>
                        </div>
                    </div>
                    <div class="flex justify-center m-6 space-x-6">
                        <button @click="downloadResult" class="btn btn-info w-36"> <i class="fa-solid fa-download"></i> Download</button>
                        <button @click="clearJobID" class="btn btn-error w-36"> <i class="fa-solid fa-xmark"></i> Clear Job</button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">

const fileInput = ref<HTMLInputElement | null>(null);
const jobID = ref<string>('');
const beforeImageUrl = ref<string>('');
const afterImageUrl = ref<string>('');
const jobStatus = ref<string>('pending');

const triggerFileInput = () => {
    fileInput.value?.click();
};

const handleFileUpload = async (event: Event) => {
    const target = event.target as HTMLInputElement;
    const file = target.files?.[0];
    if (!file) return;

    beforeImageUrl.value = URL.createObjectURL(file);

    const formData = new FormData();
    formData.append('file', file);

    try {
        const response = await fetch('http://localhost:8080/upload', {
            method: 'POST',
            body: formData
        });

        const data = await response.json();
        jobID.value = data.jobID;

        pollJobStatus();

    } catch (error) {
        console.error('Error uploading file:', error);
    }
};

const pollJobStatus = async () => {
    const interval = setInterval(async () => {
        try {
            const response = await fetch(`http://localhost:8080/status/${jobID.value}`);
            const data = await response.json();
            jobStatus.value = data.status;

            if (jobStatus.value === 'completed') {
                clearInterval(interval);
                afterImageUrl.value = `http://localhost:8080/uploads/${jobID.value}.pdf`;
            }
        } catch (error) {
            console.error('Error checking job status:', error);
        }
    }, 2000);
};

const downloadResult = () => {
    const link = document.createElement('a');
    link.href = `http://localhost:8080/download/${jobID.value}.pdf`;
    link.download = `${jobID.value}.pdf`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
};

const clearJobID = () => {
    jobID.value = '';
    beforeImageUrl.value = '';
    afterImageUrl.value = '';
    jobStatus.value = 'pending';
};
</script>