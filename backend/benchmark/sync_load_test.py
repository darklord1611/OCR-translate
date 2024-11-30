
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

class QuickstartUser(HttpUser):
    wait_time = between(1, 5)

    @task
    def test_upload(self):
        
        payload = {
            "name": "John Doe",

        }
        
        files = {
            "file": _get_image_part("images/" + random.choice(weighted_images))
        }
        
        start_time = time.time()
        url = "https://8081-01j9vf08vxz2dsg2y5m0g74nxh.cloudspaces.litng.ai/upload"
        response = requests.post(url, data=payload, files=files)
        
        # Step 3: Task is complete2
        total_time = time.time() - start_time
        self.environment.events.request.fire(
            request_type="POST",
            name="Upload and Complete Task",
            response_time=total_time * 1000,  # Convert to ms
            response_length=len(response.content),
        )
        