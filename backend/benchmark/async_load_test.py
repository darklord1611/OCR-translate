
# /teamspace/studios/this_studio/OCR-translate/backend/data/sample.png
import time
from locust import FastHttpUser, task, between, HttpUser, tag
import random
import requests
import os
from time import perf_counter


# Define the path to the directory containing images
image_directory = "./images"

# Get all files from the directory
all_images = [f for f in os.listdir(image_directory) if f.lower().endswith((".png", ".jpg", ".jpeg"))]

# Categorize images into groups based on prefixes
image_groups = {"L": [], "M": [], "S": []}
for image in all_images:
    prefix = image.split("-")[0].upper()  # Extract prefix (L, M, S) and normalize case
    if prefix in image_groups:
        image_groups[prefix].append(image)

# Define weights for each group
group_weights = {"L": 0.3, "M": 0.5, "S": 0.2}

# Create a weighted list of images
weighted_images = []
for group, weight in group_weights.items():
    weighted_images.extend(image_groups[group] * int(weight * 100))


def _get_image_part(file_path):
    import os
    file_name = os.path.basename(file_path)
    file_content = open(file_path, 'rb')
    return file_name, file_content


class MyUser(HttpUser):
    wait_time = between(1, 5)  # Wait between tasks to simulate user behavior

    @task
    @tag("upload")
    def upload_and_check_status(self):
        
        # Step 1: Upload a file
        payload = {
            "name": "John Doe",

        }
        
        files = {
            "file": _get_image_part("images/" + random.choice(weighted_images))
        }
        
        start_time = time.time()
        base_url = "https://8090-01j9vf08vxz2dsg2y5m0g74nxh.cloudspaces.litng.ai"
        with requests.post(f"{base_url}/upload", data=payload, files=files) as response:
            if response.status_code != 200:
                response.failure("File upload failed!")
                return
            # Assume response contains {"jobID": "12345"}
            job_id = response.json().get("jobID")
            if not job_id:
                response.failure("No jobID in response!")
                return
        
        # Step 2: Poll the status endpoint
        status = "pending"
        while status != "completed":
            with requests.get(f"{base_url}/status/{job_id}") as status_response:
                status = status_response.json().get("status")
                if status != "completed":
                    time.sleep(1)  # Wait before polling again

        # Step 3: Task is complete
        total_time = time.time() - start_time
        self.environment.events.request.fire(
            request_type="POST",
            name="Upload and Complete Task",
            response_time=total_time * 1000,  # Convert to ms
            response_length=len(response.content),
        )

