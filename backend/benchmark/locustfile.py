
# /teamspace/studios/this_studio/OCR-translate/backend/data/sample.png
import time
from locust import FastHttpUser, task, between, HttpUser, tag
import random
from time import perf_counter

sample_images = ["images/small-len-paragraph.PNG", "images/M-1.png", "images/L-1.png", "images/S-1.png", "images/medium-len-paragraph.PNG", "images/large-len-paragraph.PNG", "images/sample.png"]

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
        
        start_time = time.time()
        
        # Step 1: Upload a file
        payload = {
            "name": "John Doe",

        }
        
        files = {
            "file": _get_image_part(random.choice(sample_images))
        }
        
        with self.client.post("/upload", data=payload, files=files, catch_response=True) as response:
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
            with self.client.get(f"/status/{job_id}", catch_response=True) as status_response:

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
